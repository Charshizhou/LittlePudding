package main

import (
	. "LittlePudding/service"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func testJobRun() {
	jobRunner := NewJobRunner(0, 2)
	jobRunner.Run()
	job := NewJob(0, "CMD go", "D:\\GoPro\\LittlePudding\\scripts\\hello.go")
	job_2 := NewJob(1, "CMD python", "D:\\GoPro\\LittlePudding\\scripts\\hello.py 1 2")
	jobRunner.JobChan <- job
	jobRunner.JobChan <- job_2
	time.Sleep(10 * time.Second)
	jobRunner.Stop()
	jobRunner.Wait()
	fmt.Println("%v", job)
	fmt.Println("%v", job_2)
}

func main() {
	testJobRun()
}
