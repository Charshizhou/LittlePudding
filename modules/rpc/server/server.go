package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"LittlePudding/modules/rpc/auth"
	pb "LittlePudding/modules/rpc/proto"
	"LittlePudding/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	Executor *service.Executor
	pb.UnimplementedTaskServiceServer
}

var keepAlivePolicy = keepalive.EnforcementPolicy{
	MinTime:             10 * time.Second,
	PermitWithoutStream: true,
}

var keepAliveParams = keepalive.ServerParameters{
	MaxConnectionIdle: 30 * time.Second,
	Time:              30 * time.Second,
	Timeout:           3 * time.Second,
}

func (s *Server) RunTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	log.Infof("received task: %v", req)
	task := &service.Task{
		Id:            int(req.Id),
		Priority:      int(req.Priority),
		ExecTime:      time.Unix(req.ExecTime, 0),
		RouteStrategy: req.RouteStrategy,
		TaskType:      req.TaskType,
		TaskParam:     req.TaskParam,
		TaskStatus:    service.Queuing,
	}
	var result *service.TaskResult
	result, err := s.Executor.RunTask(context.Background(), task)
	if err != nil {
		log.Errorf("run task failed: %v", err)
		return nil, err
	}
	resp := &pb.TaskResponse{
		Id:     int32(result.Id),
		Result: int32(result.Result),
		Error:  "",
	}
	return resp, err
}

func Start(addr string, enableTLS bool, certificate auth.Certificate) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepAliveParams),
		grpc.KeepaliveEnforcementPolicy(keepAlivePolicy),
	}
	if enableTLS {
		tlsConfig, err := certificate.GetTLSConfigForServer()
		if err != nil {
			log.Fatal(err)
		}
		opt := grpc.Creds(credentials.NewTLS(tlsConfig))
		opts = append(opts, opt)
	}
	server := grpc.NewServer(opts...)
	executor := service.NewExecutor(0, 5, 50, 50, 100)
	executor.Run()
	pb.RegisterTaskServiceServer(server, &Server{Executor: executor})
	log.Infof("server listen on %s", addr)

	go func() {
		if err = server.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		s := <-c
		log.Infoln("收到信号 -- ", s)
		switch s {
		case syscall.SIGHUP:
			log.Infoln("收到终端断开信号, 忽略")
		case syscall.SIGINT, syscall.SIGTERM:
			log.Info("应用准备退出")
			server.GracefulStop()
			return
		}
	}
}
