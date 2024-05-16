package client

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/status"

	"LittlePudding/modules/logger"
	"LittlePudding/modules/rpc/grpcpool"
	pb "LittlePudding/modules/rpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
)

var (
	taskMap sync.Map
)

var (
	errUnavailable = errors.New("无法连接远程服务器")
)

func generateTaskUniqueKey(ip string, port int, id int32) string {
	return fmt.Sprintf("%s:%d:%d", ip, port, id)
}

func Stop(ip string, port int, id int32) {
	key := generateTaskUniqueKey(ip, port, id)
	cancel, ok := taskMap.Load(key)
	if !ok {
		return
	}
	cancel.(context.CancelFunc)()
}

func Exec(ip string, port int, taskReq *pb.TaskRequest) (string, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("panic#rpc/client.go:Exec#", err)
		}
	}()
	addr := fmt.Sprintf("%s:%d", ip, port)
	c, err := grpcpool.Pool.Get(addr)
	if err != nil {
		return "", err
	}
	if taskReq.TaskTimeout <= 0 || taskReq.TaskTimeout > 86400 {
		taskReq.TaskTimeout = 86400
	}
	timeout := time.Duration(taskReq.TaskTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	taskUniqueKey := generateTaskUniqueKey(ip, port, taskReq.Id)
	taskMap.Store(taskUniqueKey, cancel)
	defer taskMap.Delete(taskUniqueKey)

	resp, err := c.RunTask(ctx, taskReq)
	var result string
	if resp.Result == 0 {
		result = "Success"
	} else {
		result = "Failed"
	}
	if err != nil {
		return parseGRPCError(err)
	}

	if resp.Error == "" {
		return result, nil
	}

	return result, errors.New(resp.Error)
}

func parseGRPCError(err error) (string, error) {
	switch status.Code(err) {
	case codes.Unavailable:
		return "", errUnavailable
	case codes.DeadlineExceeded:
		return "", errors.New("执行超时, 强制结束")
	case codes.Canceled:
		return "", errors.New("手动停止")
	}
	return "", err
}
