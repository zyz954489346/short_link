package libraries

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"runtime"
)

// Response 通用的返回体
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// 获取方法调用者的名称
func getCallerName() string {
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		return runtime.FuncForPC(pc).Name()
	}
	return ""
}

// Ok http 返回成功
func Ok(c *gin.Context, resp *Response) {
	// 默认使用函数名作为msg
	if resp.Message == "" {
		if caller := getCallerName(); caller != "" {
			resp.Message = caller
		}
	}
	c.JSON(http.StatusOK, resp)
}

// Err http 返回错误
func Err(c *gin.Context, resp *Response) {
	if resp.Code == 0 {
		resp.Code = -1
	}

	if resp.Message == "" {
		if caller := getCallerName(); caller != "" {
			resp.Message = caller
		}
	}

	if resp.Message == "EOF" {
		resp.Message = "参数不能为空"
	}

	logrus.Error(fmt.Sprintf("\033[41m %s \033[0m", resp.Message))

	c.JSON(http.StatusOK, resp)
}
