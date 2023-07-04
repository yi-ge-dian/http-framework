package http_framework

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

// HTTPOption 配置模式
type HTTPOption func(h *HTTPServer)

// HandleFunc 视图函数的签名
type HandleFunc func(w http.ResponseWriter, r *http.Request)

type Server interface {
	// Handler 硬性要求，必须要组合 http.Handler
	http.Handler
	// Start 启动服务
	Start(address string) error
	// Stop	关闭服务
	Stop() error
	// add Router 注册路由，核心 API
	addRouter(method string, pattern string, handlefunc HandleFunc)
}

type HTTPServer struct {
	// http 包下内置的 Server
	srv *http.Server
	// 优雅关闭的属性，通过 Option 设计模式来确定
	stop func() error
	// routers，临时存放路由的位置
	routers map[string]HandleFunc
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
	h := &HTTPServer{
		routers: make(map[string]HandleFunc, 0),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// ServeHTTP 接受请求，转发请求
// 接收请求: 接收前端传过来的请求
// 转发请求: 转发前端过来的请求到自定义的框架中
// ServeHTTP 方法向前对接前端请求，向后对接咱们的框架
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. 匹配路由
	key := fmt.Sprintf("%s-%s", r.Method, r.URL.Path)
	handlerFunc, ok := h.routers[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 NOT FOUND"))
		return
	}
	// 2. 转发路由
	handlerFunc(w, r)
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

// addRouter 注册路由
// 注册路由的时机:就是项目启动的时候注册，项目启动之后就不不能注册了
func (h *HTTPServer) addRouter(method string, pattern string, handlefunc HandleFunc) {
	// 构建唯一的key
	key := fmt.Sprintf("%s-%s", method, pattern)
	fmt.Printf("add router %s - %s\n", method, pattern)
	h.routers[key] = handlefunc
}

func (h *HTTPServer) GET(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodGet, pattern, handleFunc)
}

func (h *HTTPServer) POST(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodPost, pattern, handleFunc)
}

func (h *HTTPServer) DElETE(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodDelete, pattern, handleFunc)
}

// func main() {
// 	h := NewHTTP(WithHTTPServerStop(nil))
// 	go func() {
// 		err := h.Start(":8080")
// 		if err != nil && err != http.ErrServerClosed {
// 			panic("启动失败")
// 		}
// 	}()
// 	err := h.Stop()
// 	if err != nil {
// 		panic("关闭失败")
// 	}
// }
