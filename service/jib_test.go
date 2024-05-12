package service

import "testing"

func TestJobRun(t *testing.T) {
	jobRunner := NewJobRun(0, "127.0.0.1:9000", 2)
	job := Job{
		Id:       0,
		GlueType: "CMD",
		CmdParam: "go run hello.go",
	}
	jobRunner.JobChan <- job
	jobRunner.Run()
}
