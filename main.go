package main

import (
	"context"
	"flag"
	"fmt"
	"go-web-app/dao"
	"go-web-app/logger"
	"go-web-app/routes"
	"go-web-app/settings"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Go Web开发比较通用的脚手架模板

func main() {
	var appConfigpath string
	flag.StringVar(&appConfigpath, "c", "", "Configuration file path")
	flag.Parse()
	//1. 加载配置文件

	if err := settings.Init(appConfigpath); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}
	//2. 初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			zap.L().Error("L.Sync failed...")
		}
	}(zap.L())
	zap.L().Debug("logger init success...")
	//加载数据采集程序

	go func() {
		Hostid, err := dao.ServerConfirm(settings.Conf.ServerConfig)
		if err != nil {
			zap.L().Error("init ServerConfirm failed, err:%v\n", zap.Error(err))
			return
		}
		fmt.Println(Hostid)
		if err := dao.Processing(Hostid, settings.Conf); err != nil {
			zap.L().Error("init Processing failed, err:%v\n", zap.Error(err))
			return
		}
	}()

	//3. 注册路由
	r := routes.Setup()
	//4. 启动服务 （优雅关机）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: r,
	}

	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown: ", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
