package dispatcher

import (
	"LittlePudding/models"
	"LittlePudding/modules/rpc/client"
	pb "LittlePudding/modules/rpc/proto"
	"LittlePudding/service"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Dispatcher struct {
	wait sync.WaitGroup
}

/*
调度器
*/

// NewDispatcher 创建调度器
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		wait: sync.WaitGroup{},
	}
}

// UpdateNextRunTime 更新所有任务的下次执行时间
func UpdateNextRunTime() error {
	location, _ := time.LoadLocation("Asia/Shanghai")
	// 查询所有激活任务
	tasks := make([]models.Task, 0)
	err := models.Db.Where("status = ?", models.Enabled).Find(&tasks)
	if err != nil {
		return err
	}
	// 根据cron表达式计算下次执行时间
	for _, task := range tasks {
		task.NextRunTime, err = task.CalculateNextRunTimeWithLocation(location)
		utcNextRunTime := task.NextRunTime.Add(+8 * time.Hour)
		_, err = task.Update(task.Id, models.CommonMap{"next_run_time": utcNextRunTime})
		if err != nil {
			return err
		}
	}
	logrus.Infoln("更新任务NextRunTime成功")
	return nil
}

func (disp *Dispatcher) dispatchTasks() error {
	currentTime := time.Now()
	futureTime := currentTime.Add(30 * time.Second)

	tasks := make([]models.Task, 0)
	err := models.Db.Table("task").
		Select("task.*, executor.address AS executor_addr").
		Join("LEFT", "executor", "task.executor_id = executor.id").
		Where("task.next_run_time BETWEEN ? AND ?", currentTime, futureTime).
		Find(&tasks)
	errChan := make(chan error, len(tasks))

	for _, task := range tasks {
		disp.wait.Add(1)
		go func(task models.Task) {
			defer disp.wait.Done()
			taskReq := &pb.TaskRequest{
				Id:            int32(task.Id),
				Priority:      int32(task.Priority),
				ExecTime:      task.NextRunTime.Unix(),
				RouteStrategy: task.ExecutorRouteStrategy,
				TaskType:      task.TaskType,
				TaskParam:     task.TaskParam,
				TaskTimeout:   int64(task.ExecuteTimeout),
			}
			row, err := models.Db.Query("SELECT executor.address AS executor_addr FROM executor WHERE executor.id = ?", task.ExecutorId)
			task.ExecutorAddr = string(row[0]["executor_addr"])
			logrus.Infof("task running: %v", task.Name)
			result, err := client.Exec(task.ExecutorAddr, taskReq)
			if result == nil {
				result = &service.TaskResult{
					Result: service.Failure,
				}
			}
			if task.ExecuteFailRetryCount > 0 && result.Result == service.Failure {
				for i := 0; i < task.ExecuteFailRetryCount; i++ {
					result, err = client.Exec(task.ExecutorAddr, taskReq)
					if result == nil {
						result = &service.TaskResult{
							Result: service.Failure,
						}
					}
					if err == nil && result.Result == service.Success {
						task.ExecuteFailRetryCount = i + 1
						break
					}
				}
			} else {
				task.ExecuteFailRetryCount = 0
			}
			task.TaskResult = (*models.TaskResult)(result)
			if _, err := models.CreateTaskLog(task); err != nil {
				errChan <- err
			}
		}(task)
	}

	disp.wait.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return err
}

func (disp *Dispatcher) Start() {
	// 启动时立即执行一次 dispatchTasks
	if err := disp.dispatchTasks(); err != nil {
		log.Printf("调度任务失败: %v", err)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := disp.dispatchTasks(); err != nil {
					log.Printf("调度任务失败: %v", err)
				}
			}
		}
	}()

	go func() {
		for {
			now := time.Now()
			next15Min := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 15, 0, 0, now.Location())
			next45Min := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, now.Location())

			if now.After(next15Min) {
				next15Min = next15Min.Add(1 * time.Hour)
			}
			if now.After(next45Min) {
				next45Min = next45Min.Add(1 * time.Hour)
			}

			timer15Min := time.NewTimer(next15Min.Sub(now))
			timer45Min := time.NewTimer(next45Min.Sub(now))
			select {
			case <-timer15Min.C:
				if err := UpdateNextRunTime(); err != nil {
					log.Printf("更新任务NextRunTime失败: %v", err)
				}
				timer45Min.Reset(next45Min.Sub(time.Now()))
			case <-timer45Min.C:
				if err := UpdateNextRunTime(); err != nil {
					log.Printf("更新任务NextRunTime失败: %v", err)
				}
				timer15Min.Reset(next15Min.Add(1 * time.Hour).Sub(time.Now()))
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		log.Printf("收到信号 -- %v", s)
		switch s {
		case syscall.SIGHUP:
			log.Printf("收到终端断开信号, 忽略")
		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("应用准备退出")
			return
		}
	}
}
