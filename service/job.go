package service

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
)

type Status int8
type Result int8
type JobFunc func()

const ( // 完成
	Success Result = iota // 成功
	Failure               // 失败
)

const (
	Queuing  Status = iota //排队
	Running                // 运行中
	Finished               // 完成
)

type Job struct {
	Id        int
	RunType   string //任务类型 GLUE 脚本
	GlueParam string //cobra参数
	CmdType   string //脚本类型
	CmdParam  string //脚本参数
	CmdPath   string //脚本路径
	Status    Status //执行状态
	Result    Result //执行结果
	errInfo   string //错误信息
}

// JobRunner 执行器
type JobRunner struct {
	Id        int
	address   string
	JobChan   chan *Job
	quit      chan struct{}
	Wait      sync.WaitGroup
	maxJobNum int
}

// NewJob 创建任务
func NewJob(id int, runType string, cmdType string, cmdParam, glueParam, cmdPath interface{}) *Job {
	if runType == "GLUE" {
		return &Job{
			Id:        id,
			RunType:   runType,
			GlueParam: glueParam.(string),
		}
	} else {
		return &Job{
			Id:       id,
			RunType:  runType,
			CmdType:  cmdType,
			CmdParam: cmdParam.(string),
			CmdPath:  cmdPath.(string),
		}
	}
}

// Run 任务执行
// TODO: GLUE类型任务执行
func (t *Job) Run() (err error) {
	if t.RunType == "GLUE" {
		// 执行脚本
	}
	// 执行命令
	err = t.RunByCmd()
	if err != nil {
		// 执行失败
		fmt.Println("执行失败", err)
		return err
	}
	return nil
}

// RunByCmd 执行命令
func (t *Job) RunByCmd() (err error) {
	var cmd *exec.Cmd
	switch t.CmdType {
	case "go":
		// 执行go脚本
		if t.CmdParam == "" {
			cmd = exec.Command("go", "run", t.CmdPath)
		} else {
			cmd = exec.Command("go", "run", t.CmdParam, t.CmdPath)
		}
		output, err := cmd.Output()
		if err != nil {
			// 执行失败
			fmt.Println("执行失败", err)
			return err
		}
		fmt.Printf("执行成功,输出结果:%s\n", output)
	case "python":
		// 执行python脚本
		if t.CmdParam == "" {
			cmd = exec.Command("python", t.CmdPath)
		} else {
			cmd = exec.Command("python", t.CmdParam, t.CmdPath)
		}
		output, err := cmd.Output()
		if err != nil {
			// 执行失败
			fmt.Println("执行失败", err)
			return err
		}
		fmt.Printf("执行成功,输出结果:%s\n", output)
	}
	return nil
}

// NewJobRunner 创建执行器
func NewJobRunner(id int, address string, maxJobNum interface{}) *JobRunner {
	// 默认最大任务数
	if maxJobNum == nil {
		maxJobNum = 10
	}
	return &JobRunner{
		Id:        id,
		address:   address,
		JobChan:   make(chan *Job, maxJobNum.(int)),
		quit:      make(chan struct{}),
		Wait:      sync.WaitGroup{},
		maxJobNum: maxJobNum.(int),
	}
}

// Run 执行队列
func (t *JobRunner) Run() {
	t.Wait.Add(1)
	go func() {
		defer t.Wait.Done()
		for {
			select {
			case job := <-t.JobChan:
				go func() {
					job.Status = Running
					err := job.Run()
					if err != nil {
						job.Result = Failure
						job.errInfo = err.Error()
					}
					job.Result = Success
					job.Status = Finished
				}()
			case <-t.quit:
				return
			default:
				// 为空则继续接收 1ms
				time.Sleep(time.Millisecond)
			}
		}
	}()
}

func (t *JobRunner) Stop() {
	close(t.quit)
	t.Wait.Wait()
}

func (t *JobRunner) AddJob(job *Job) {
	job.Status = Queuing
	t.JobChan <- job
}

func (t *JobRunner) IsFull() bool {
	return len(t.JobChan) >= t.maxJobNum
}
