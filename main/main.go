package main

import (
	"LittlePudding/modules/config"
	"LittlePudding/modules/rpc/auth"
	pb "LittlePudding/modules/rpc/proto"
	"LittlePudding/modules/rpc/server"
	"LittlePudding/modules/utils"
	"context"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"strings"
	"time"
)

/*
job := NewJob(0, "CMD go", "D:\\GoPro\\LittlePudding\\scripts\\hello.go")
job_2 := NewJob(1, "CMD python", "D:\\GoPro\\LittlePudding\\scripts\\hello.py 1 2")
*/

func testExecutor() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	log.Infof("connected to server: %v", "localhost:50051")
	defer conn.Close()

	c := pb.NewTaskServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req := &pb.TaskRequest{
		Id:            1,
		Priority:      1,
		ExecTime:      time.Now().Unix(),
		RouteStrategy: "RoundRobin",
		TaskType:      "CMD python",
		TaskParam:     "D:\\GoPro\\LittlePudding\\scripts\\hello.py 1 2",
	}

	res, err := c.RunTask(ctx, req)
	if err != nil {
		log.Fatalf("could not run task: %v", err)
	}

	log.Printf("Task Response: %v", res)
}

func StartServer() {

	var serverAddr string
	var allowRoot bool
	var version bool
	var CAFile string
	var certFile string
	var keyFile string
	var enableTLS bool
	var logLevel string
	flag.BoolVar(&allowRoot, "allow-root", false, "./gocron-node -allow-root")
	flag.StringVar(&serverAddr, "s", "0.0.0.0:5921", "./gocron-node -s ip:port")
	flag.BoolVar(&version, "v", false, "./gocron-node -v")
	flag.BoolVar(&enableTLS, "enable-tls", false, "./gocron-node -enable-tls")
	flag.StringVar(&CAFile, "ca-file", "", "./gocron-node -ca-file path")
	flag.StringVar(&certFile, "cert-file", "", "./gocron-node -cert-file path")
	flag.StringVar(&keyFile, "key-file", "", "./gocron-node -key-file path")
	flag.StringVar(&logLevel, "log-level", "info", "-log-level error")
	flag.Parse()
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)

	if enableTLS {
		if !utils.FileExist(CAFile) {
			log.Fatalf("failed to read ca cert file: %s", CAFile)
		}
		if !utils.FileExist(certFile) {
			log.Fatalf("failed to read server cert file: %s", certFile)
			return
		}
		if !utils.FileExist(keyFile) {
			log.Fatalf("failed to read server key file: %s", keyFile)
			return
		}
	}

	certificate := auth.Certificate{
		CAFile:   strings.TrimSpace(CAFile),
		CertFile: strings.TrimSpace(certFile),
		KeyFile:  strings.TrimSpace(keyFile),
	}
	go server.Start("localhost:50051", false, certificate)
}

func main() {
	var s *config.Setting
	s, err := config.ReadConfig("config.ini")
	if err != nil {
		log.Fatalf("读取配置文件失败-%s", err)
	}
	fmt.Println(s)
}
