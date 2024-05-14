package service

import (
	"container/heap"
	"math/rand"
	"sync"
	"time"
)

/*
执行器组件设计文档：
目前暂定为调度中心生产出来的是task，由选择的jobRunner去消费
大量任务被发往该执行器组中
快堆慢堆 快堆用来存放有优先级的任务，慢堆用来存放无优先级的任务
任务根据预计执行时间与当前时间的差值来进行排序

*/

/*
id int, runType string, cmdType string, cmdParam, glueParam, cmdParam interface{}
*/

type RouteStrategy int

const (
	Random     RouteStrategy = iota // 随机
	RoundRobin                      // 轮询
	LeastTask                       // 最少任务
)

type Task struct {
	Id            int         // 任务ID
	Priority      int         // 任务优先级
	ExecTime      time.Time   // 预计执行时间
	RouteStrategy string      // 路由策略
	TaskType      string      // 任务类型
	TaskParam     string      // 任务参数
	TaskResult    *TaskResult // 任务结果
	TaskStatus    Status      // 任务状态
}

type TaskResult JobResult

type TaskHeap []*Task

func (h TaskHeap) Len() int { return len(h) }
func (h TaskHeap) Less(i, j int) bool {
	if h[i].Priority == h[j].Priority {
		return h[i].ExecTime.Before(h[j].ExecTime)
	}
	return h[i].Priority < h[j].Priority
}
func (h TaskHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *TaskHeap) Push(x interface{}) {
	*h = append(*h, x.(*Task))
}

func (h *TaskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

type Executor struct {
	Id         int
	mutex      sync.RWMutex       // 执行器读写锁
	isFull     *sync.Cond         // 执行器是否满
	FastHeap   TaskHeap           // 快堆，存放有优先级的任务
	SlowHeap   TaskHeap           // 慢堆，存放无优先级的任务
	taskChan   chan *Task         // 任务缓冲通道
	taskMap    map[int]*Task      // 任务映射
	ResultChan chan *JobResult    // 任务结果通道
	jobRunners map[int]*JobRunner // 执行器组
	lastJobRun int                // 轮询上次执行的执行器
	Status     Status             // 执行器状态
}

func NewExecutor(id, numJobRunners, fastHeapSize, slowHeapSize, taskMapNum int) *Executor {
	fastHeap := make(TaskHeap, 0, fastHeapSize)
	slowHeap := make(TaskHeap, 0, slowHeapSize)
	taskChan := make(chan *Task, 100)
	resultChan := make(chan *JobResult, 100)
	taskMap := make(map[int]*Task, taskMapNum)
	heap.Init(&fastHeap)
	heap.Init(&slowHeap)
	jobRunners := make(map[int]*JobRunner, numJobRunners)
	for i := 1; i < numJobRunners; i++ {
		jobRunners[i] = NewJobRunner(i, 5)
		go jobRunners[i].Run()
	}
	return &Executor{
		Id:         id,
		FastHeap:   fastHeap,
		SlowHeap:   slowHeap,
		jobRunners: jobRunners,
		taskChan:   taskChan,
		ResultChan: resultChan,
		taskMap:    taskMap,
		lastJobRun: 0,
		Status:     Running,
		mutex:      sync.RWMutex{},
		isFull:     sync.NewCond(&sync.Mutex{}),
	}
}

// AddTasksToExec 生产任务
func (e *Executor) AddTasksToExec() {
	for {
		// 从快堆取出优先级最高的任务执行
		if e.FastHeap.Len() > 0 {
			task := heap.Pop(&e.FastHeap).(*Task)
			// 任务添加到缓冲区
			e.addTaskToExec(task)
		}
		// 如果快堆为空，则从慢堆取出任务执行
		if e.SlowHeap.Len() > 0 {
			task := heap.Pop(&e.SlowHeap).(*Task)
			// 任务添加到缓冲区
			e.addTaskToExec(task)
		}
		// 如果快慢堆都为空，则等待
		if e.FastHeap.Len() == 0 && e.SlowHeap.Len() == 0 {
			e.isFull.Signal()
		}
	}
}

// ExecuteTask 消费任务
func (e *Executor) ExecuteTask() {
	for {
		availableJobRunners := e.GetAvailableJobRunners()
		// 如果没有可用的执行器，则等待
		if len(availableJobRunners) == 0 {
			e.isFull.L.Lock()
			e.isFull.Wait()
			e.isFull.L.Unlock()
			continue
		}

		select {
		case task := <-e.taskChan:
			e.dispatchTask(*task, availableJobRunners)
		case result := <-e.ResultChan:
			// 任务执行结果与map中的任务进行关联
			if task, ok := e.taskMap[result.Id]; ok {
				task.TaskResult = (*TaskResult)(result)
				task.TaskStatus = Finished
			}
			e.isFull.Signal()
		case <-time.After(1 * time.Second):
			availableJobRunners = e.GetAvailableJobRunners()
			if len(availableJobRunners) == 0 {
				e.isFull.L.Lock()
				e.isFull.Wait()
				e.isFull.L.Unlock()
			}
		}
	}
}

// Run 执行器运行
func (e *Executor) Run() {
	// 启动生产者
	for i := 0; i < 2; i++ {
		go e.AddTasksToExec()
	}
	for i := 0; i < 2; i++ {
		go e.ExecuteTask()
	}
}

func (e *Executor) dispatchTask(task Task, jobRunners []int) {
	job := NewJob(task.Id, task.TaskType, task.TaskParam)
	switch task.RouteStrategy {
	case "Random":
		// 随机选择一个执行器
		e.mutex.RLock()
		defer e.mutex.RUnlock()
		e.jobRunners[jobRunners[rand.Intn(len(jobRunners))]].AddJob(job)
	case "RoundRobin":
		// 读锁
		e.mutex.Lock()
		defer e.mutex.Unlock()
		for i := 0; i < len(e.jobRunners); i++ {
			idx := (e.lastJobRun + i) % len(e.jobRunners)
			if e.jobRunners[idx].IsAvailable() {
				e.jobRunners[idx].AddJob(job)
				e.lastJobRun = idx
				return
			}
		}
	case "LeastTask":
		// 选择任务最少的执行器
		e.mutex.Lock()
		defer e.mutex.Unlock()
		minJobRunner := 0
		for i := 0; i < len(e.jobRunners); i++ {
			if len(e.jobRunners[i].JobChan) < len(e.jobRunners[minJobRunner].JobChan) {
				minJobRunner = i
			}
		}
		e.jobRunners[minJobRunner].AddJob(job)
	}
}

// GetAvailableJobRunners 获取所有可用的执行器
func (e *Executor) GetAvailableJobRunners() (jobRunners []int) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	for _, jobRunner := range e.jobRunners {
		if jobRunner.IsAvailable() {
			jobRunners = append(jobRunners, jobRunner.Id)
		}
	}
	return
}

func (e *Executor) addTaskToExec(task *Task) {
	// 执行任务
	e.taskChan <- task
}

func (e *Executor) RunTask(id int, priority int, execTime time.Time, routeStrategy, taskType, taskParam string) {
	task := &Task{
		Id:            id,
		Priority:      priority,
		ExecTime:      execTime,
		RouteStrategy: routeStrategy,
		TaskType:      taskType,
		TaskParam:     taskParam,
		TaskStatus:    Queuing,
	}
	if task.Priority > 0 {
		heap.Push(&e.FastHeap, task)
	} else {
		heap.Push(&e.SlowHeap, task)
	}
	//在taskMap中添加
	e.taskMap[task.Id] = task
}
