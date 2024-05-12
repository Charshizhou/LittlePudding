package service

import (
	"sync"
)

/*
设计文档：
目前暂定为调度中心生产出来的是task，由选择的jobRunner去消费
*/

type Task struct {
	Id          int
	TaskName    string
	TaskType    string
	TaskContent string
	TaskStatus  Status
}

type Executor struct {
	Id           int
	ExecutorName string
	mutex        sync.Mutex
	taskChan     chan Task
	jobRuns      map[int]*JobRunner
}

func NewExecutor(id int, executorName string) *Executor {
	return &Executor{
		Id:           id,
		ExecutorName: executorName,
		jobRuns:      make(map[int]*JobRunner),
	}
}

func (e *Executor) AddTaskRun(address string, maxJobNum int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.jobRuns[len(e.jobRuns)] = NewJobRunner(len(e.jobRuns), address, maxJobNum)
}

func (e *Executor) GetTaskRun(id int) *JobRunner {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.jobRuns[id]
}

func (e *Executor) RemoveTaskRun(id int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	delete(e.jobRuns, id)
}

func (e *Executor) GetTaskRuns() map[int]string {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	jobRuns := make(map[int]string)
	for id, jobRun := range e.jobRuns {
		jobRuns[id] = jobRun.address
	}
	return jobRuns
}

//func (e *Executor) RouteTaskRun()
