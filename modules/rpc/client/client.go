package client

import (
	"LittlePudding/service"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"

	"LittlePudding/modules/rpc/grpcpool"
	pb "LittlePudding/modules/rpc/proto"
	"golang.org/x/net/context"
)

var (
	taskMap sync.Map
)

var (
	errUnavailable = errors.New("无法连接远程服务器")
)

func generateTaskUniqueKey(addr string, id int32) string {
	return fmt.Sprintf("%s:%d", addr, id)
}

func Stop(addr string, id int32) {
	key := generateTaskUniqueKey(addr, id)
	cancel, ok := taskMap.Load(key)
	if !ok {
		return
	}
	cancel.(context.CancelFunc)()
}

func Exec(addr string, taskReq *pb.TaskRequest) (*service.TaskResult, error) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Error("panic#rpc/client.go:Exec#", err)
		}
	}()
	c, err := grpcpool.Pool.Get(addr)
	if err != nil {
		logrus.Infof("pool err: %v", err)
		return &service.TaskResult{}, err
	}
	if taskReq.TaskTimeout <= 0 || taskReq.TaskTimeout > 86400 {
		taskReq.TaskTimeout = 86400
	}
	// 设置任务超时时间 + 30s
	timeout := time.Duration(taskReq.TaskTimeout+30) * time.Second
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	taskUniqueKey := generateTaskUniqueKey(addr, taskReq.Id)
	taskMap.Store(taskUniqueKey, cancel)
	defer taskMap.Delete(taskUniqueKey)
	logrus.Infof("taks run %v", taskReq.Id)
	resp, err := c.RunTask(context.Background(), taskReq)
	logrus.Infof("task result: %v", resp)
	logrus.Infof("time: %v    %v", resp.ExecTime, resp.DispatchTime)
	if err != nil {
		return &service.TaskResult{}, err
	}
	taskResult := &service.TaskResult{
		Id:           int(resp.Id),
		ExecTime:     time.Unix(resp.ExecTime, 0),
		DispatchTime: time.Unix(resp.DispatchTime, 0),
		Result:       service.Result(resp.Result),
		Err:          errors.New(resp.Error),
	}

	if resp.Error == "" {
		return taskResult, nil
	}
	return taskResult, errors.New(resp.Error)
}
