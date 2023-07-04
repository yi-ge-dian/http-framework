package http_framework

import (
	"net/http"
	"testing"
)

// 注册路由是在启动服务之前完成
func Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login 请求成功!"))
}
func Register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login 请求成功!"))
}

func TestHTTP_Start(t *testing.T) {
	h := NewHTTP()
	h.GET("/login", Login)
	h.POST("/register", Register)

	err := h.Start(":8080")
	if err != nil {
		t.Fail()
	}
}
