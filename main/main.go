package main

import (
	"LittlePudding/models"
	conf "LittlePudding/modules/config"
	"LittlePudding/modules/dispatcher"
	"LittlePudding/modules/logger"
	"LittlePudding/modules/routers"
	"LittlePudding/modules/rpc/auth"
	"LittlePudding/modules/rpc/client"
	pb "LittlePudding/modules/rpc/proto"
	"LittlePudding/modules/rpc/server"
	"LittlePudding/modules/utils"
	"crypto/tls"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

/*
job := NewJob(0, "CMD go", "D:\\GoPro\\LittlePudding\\scripts\\hello.go")
job_2 := NewJob(1, "CMD python", "D:\\GoPro\\LittlePudding\\scripts\\hello.py 1 2")
*/

func StartServer(setting *conf.Setting) {
	serverAddr, err := models.GetAllAddress()
	if err != nil {
		log.Fatalf("获取服务地址失败")
		return
	}
	enableTLS := setting.EnableTLS
	certFile := setting.CertFile
	keyFile := setting.KeyFile

	if enableTLS {
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
		CertFile: strings.TrimSpace(certFile),
		KeyFile:  strings.TrimSpace(keyFile),
	}
	for _, addr := range serverAddr {
		go server.Start(addr, enableTLS, certificate)
	}
}

func InitModule() (setting *conf.Setting) {
	setting, err := conf.ReadConfig("config.ini")
	if err != nil {
		logger.Fatal("读取应用配置失败", err)
	}
	// 初始化DB
	models.Db = models.InitDb(setting)
	return
}

func InitGin() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	routers.Register(r)

	f, err := os.Create("gin.log")
	if err != nil {
		log.Fatal(err)
	}
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	err = r.Run("127.0.0.1:8051")
	if err != nil {
		logger.Fatal("无法启动服务器", err)
	}
}

func testExecutor(setting *conf.Setting) {
	// 创建请求
	req := &pb.TaskRequest{
		Id:            1,
		Priority:      1,
		ExecTime:      time.Now().Unix(),
		RouteStrategy: "RoundRobin",
		TaskType:      "CMD go",
		TaskParam:     "D:\\GoPro\\LittlePudding\\scripts\\hello.go",
		TaskTimeout:   10,
	}

	// 发送请求
	resp, err := client.Exec("127.0.0.1:50051", req)
	if err != nil {
		log.Fatalf("Could not run task: %v", err)
	}
	log.Printf("Task response: %v", resp)
}

func loadTLSCredentials(certFile, keyFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: true, // 跳过证书验证（仅在开发和测试环境中使用）
	}

	return credentials.NewTLS(config), nil
}

func StartDispatcher(setting *conf.Setting) {
	disp := dispatcher.NewDispatcher()
	err := dispatcher.UpdateNextRunTime()
	if err != nil {
		log.Fatalf("更新任务下次执行时间失败: %v", err)
	}
	disp.Start()
}

func main() {
	location, _ := time.LoadLocation("Asia/Shanghai")
	log.Printf("in %v", location)
	setting := InitModule()

	go StartServer(setting)
	go StartDispatcher(setting)

	//启动gin
	_ = InitModule()
	InitGin()
	//time.Sleep(time.Second * 2)
	//testExecutor(setting)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs
	log.Infof("收到信号: %v，正在关闭程序...", sig)
}
