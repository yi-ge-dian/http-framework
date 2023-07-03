package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HTTPOption func(h *HTTPServer)

type Server interface {
	// Handler 硬性要求，必须要组合 http.Handler
	http.Handler
	// Start 启动服务
	Start(address string) error
	// Stop	关闭服务
	Stop() error
}

type HTTPServer struct {
	// http 包下内置的 Server
	srv *http.Server
	// 优雅关闭的属性，通过 Option 设计模式来确定
	stop func() error
}

func WithHTTPServerStop(fn func() error) HTTPOption {
	return func(h *HTTPServer) {
		if fn == nil {
			fn = func() error {
				fmt.Println("Default stop function")
				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit
				log.Println("Shutdown server ...")

				// 创建一个超时上下文，给它 5s 的时间去关闭
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 服务器关机
				if err := h.srv.Shutdown(ctx); err != nil {
					log.Fatal("Server shutdown:", err)
				}
				select {
				case <-ctx.Done():
					log.Println("Spend 5 seconds")
					return nil
				default:
					return nil
				}
			}
		}
		h.stop = fn
	}
}

func NewHTTP(opts ...HTTPOption) *HTTPServer {
	h := &HTTPServer{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// ServeHTTP 接受请求，转发请求、
// 接收请求: 接收前端传过来的请求
// 转发请求: 转发前端过来的请求到自定义的框架中
func (h *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
}

// Start 开始服务
func (h *HTTPServer) Start(addr string) error {
	h.srv = &http.Server{
		Addr:    addr,
		Handler: h,
	}
	return h.srv.ListenAndServe()
}

// Stop 结束服务
func (h *HTTPServer) Stop() error {
	return h.stop()
}

func main() {
	h := NewHTTP(WithHTTPServerStop(nil))
	go func() {
		err := h.Start(":8080")
		if err != nil && err != http.ErrServerClosed {
			panic("启动失败")
		}
	}()
	err := h.Stop()
	if err != nil {
		panic("关闭失败")
	}
}
