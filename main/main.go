package main

import (
	. "LittlePudding/service"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func testJobRun() {
	jobRunner := NewJobRunner(0, "127.0.0.1:9000", 2)
	jobRunner.Run()
	job := &Job{
		Id:      0,
		RunType: "CMD",
		CmdType: "go",
		CmdPath: "D:\\GoPro\\LittlePudding\\scripts\\hello.go",
	}
	job_2 := &Job{
		Id:      1,
		RunType: "CMD",
		CmdType: "python",
		CmdPath: "D:\\GoPro\\LittlePudding\\scripts\\hello.py",
	}
	jobRunner.JobChan <- job
	jobRunner.JobChan <- job_2
	time.Sleep(10 * time.Second)
	jobRunner.Stop()
	jobRunner.Wait.Wait()
	fmt.Println("%v", job)
	fmt.Println("%v", job_2)
}

func main() {
	testJobRun()
}
