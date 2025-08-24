package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-cloud-disk/auth"
	"go-cloud-disk/cache"
	"go-cloud-disk/conf"
	"go-cloud-disk/disk"
	"go-cloud-disk/model"
	"go-cloud-disk/rabbitMQ"
	"go-cloud-disk/rabbitMQ/script"
	"go-cloud-disk/server"
	"go-cloud-disk/task"
	"go-cloud-disk/utils/logger"

	"github.com/gin-gonic/gin"
)

// initServer init server that server needed
func initServer() {
	// set cloud disk
	disk.SetBaseCloudDisk()
	// set log
	logger.BuildLogger()

	// connect database
	model.Database()

	// connect redis
	cache.Redis()

	// start regular task
	task.CronJob()

	// start casbin
	auth.InitCasbin()

	// start rabbitmq
	rabbitMQ.InitRabbitMq()
}

func loadingScript() {
	ctx := context.Background()
	go script.SendConfirmEmailSync(ctx)
}

func main() {
	// 配置初始化
	conf.Init()
	initServer()
	loadingScript()

	// 设置路由
	gin.SetMode(conf.GinMode)
	r := server.NewRouter()

	// 创建服务
	srv := &http.Server{
		Addr:    ":" + conf.ServerPort,
		Handler: r,
	}

	go func() {
		log.Println("go-cloud-disk server start")
		// 启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待系统退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	// 设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		log.Println("Server exiting")
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}

// func main() {
// 	filePath := "【哲风壁纸】动漫角色-战斗姿势.png"

// 	if _, err := os.Stat(filePath); os.IsNotExist(err) {
// 		fmt.Printf("❌ 文件不存在: %s\n", filePath)
// 		return
// 	}

// 	// 调用库函数
// 	chunkFiles, err := test.SplitFileToChunks(filePath)
// 	if err != nil {
// 		fmt.Printf("❌ 切片失败: %v\n", err)
// 		return
// 	}

// 	// 打印信息
// 	test.PrintChunkInfo(filePath, chunkFiles)
// }
