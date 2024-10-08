package service

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Status int8
type Result int8
type JobFunc func()

const ( // 完成
	Success Result = iota // 成功
	Failure               // 失败
	Expired               // 执行超时
	Overdue               // 任务过期
)

const (
	Queuing  Status = iota //排队
	Running                // 运行中
	Finished               // 完成
)

type Job struct {
	Id        int
	RunType   string        //任务类型 GLUE 脚本
	glueParam string        //cobra参数
	cmdType   string        //脚本类型
	cmdParam  string        //脚本参数
	Status    Status        //执行状态
	Timeout   time.Duration //超时时间
}

type JobResult struct {
	Id       int
	ExecTime time.Time
	Result   Result
	Err      error
}

// JobRunner 执行器
type JobRunner struct {
	Id              int
	JobChan         chan *Job
	mutexJobRunners sync.Mutex
	quit            chan struct{}
	wait            sync.WaitGroup
	maxJobNum       int
}

// NewJob 创建任务
func NewJob(id int, runType string, runParam interface{}) *Job {
	if runType == "GLUE" {
		return &Job{
			Id:        id,
			RunType:   runType,
			glueParam: runParam.(string),
		}
	} else {
		// CMD类型任务的runType格式为"CMD 脚本类型"
		cmdType := strings.Split(runType, " ")[1]
		return &Job{
			Id:       id,
			RunType:  runType,
			cmdType:  cmdType,
			cmdParam: runParam.(string),
		}
	}
}

// Run 任务执行
// TODO: GLUE类型任务执行
func (t *Job) Run(ctx context.Context) (err error) {
	if t.RunType == "GLUE" {
		// 执行脚本
	}
	// 执行命令
	err = t.RunByCmd(ctx)
	return
}

// RunByCmd 执行命令
func (t *Job) RunByCmd(ctx context.Context) (err error) {
	var cmd *exec.Cmd
	args := strings.Split(t.cmdParam, " ")
	switch t.cmdType {
	case "go":
		// 执行go脚本
		cmd = exec.CommandContext(ctx, "go", append([]string{"run"}, args...)...)
	case "python":
		// 执行python脚本
		cmd = exec.CommandContext(ctx, "python", args...)
	}
	output, err := cmd.Output()
	log.Printf("执行成功,输出结果:%s\n", output)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		err = ctx.Err()
		return
	}
	return
}

// NewJobRunner 创建执行器
func NewJobRunner(id int, maxJobNum interface{}) *JobRunner {
	// 默认最大任务数
	if maxJobNum == nil {
		maxJobNum = 10
	}
	return &JobRunner{
		Id:        id,
		JobChan:   make(chan *Job, maxJobNum.(int)),
		quit:      make(chan struct{}),
		wait:      sync.WaitGroup{},
		maxJobNum: maxJobNum.(int),
	}
}

// Run 执行队列
func (t *JobRunner) Run(resultChan chan *JobResult) {
	t.wait.Add(1)
	go func() {
		defer t.wait.Done()
		for {
			select {
			case job := <-t.JobChan:
				job.Status = Running
				defer func() {
					job.Status = Finished
				}()
				ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
				defer cancel()
				execTime := time.Now()
				err := job.Run(ctx)
				if err != nil {
					resultChan <- &JobResult{
						Id:       job.Id,
						ExecTime: execTime,
						Result:   Failure,
						Err:      err,
					}
				} else if errors.Is(err, context.DeadlineExceeded) {
					resultChan <- &JobResult{
						Id:       job.Id,
						ExecTime: execTime,
						Result:   Expired,
						Err:      err,
					}
				}
				resultChan <- &JobResult{
					Id:       job.Id,
					ExecTime: execTime,
					Result:   Success,
					Err:      nil,
				}
				log.Println("任务执行完成")
			case <-t.quit:
				return
			case <-time.After(1 * time.Second):
				continue
			}
		}
	}()
}

func (t *JobRunner) Stop() {
	close(t.quit)
	t.wait.Wait()
}

func (t *JobRunner) AddJob(job *Job) {
	t.mutexJobRunners.Lock()
	defer t.mutexJobRunners.Unlock()
	job.Status = Queuing
	t.JobChan <- job
}

// IsAvailable 是否可用
func (t *JobRunner) IsAvailable() bool {
	return len(t.JobChan) <= t.maxJobNum
}

func (t *JobRunner) Wait() {
	t.wait.Wait()
}
