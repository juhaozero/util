package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/ratelimit"
	"go.uber.org/zap"
)

// Cors 设置跨域
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") // 请求头部
		if origin != "" {
			// 接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			// 服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			// 允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
			//  允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			// 设置缓存时间
			//  c.Header("Access-Control-Max-Age", "172800")
			// 允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 允许类型校验
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		defer func() {
			if err := recover(); err != nil {
				log.Print("Panic info is", zap.Any("err", err), zap.Any("stack\n", string(debug.Stack())))
			}
		}()
		c.Next()
	}
}

// RateLimitMiddleware gin 限流
func RateLimitMiddleware(fillInterval time.Duration, cap, quantum int64) gin.HandlerFunc {
	bucket := ratelimit.NewBucketWithQuantum(fillInterval, cap, quantum)
	return func(c *gin.Context) {
		if bucket.TakeAvailable(1) < 1 {
			c.String(http.StatusTooManyRequests, "rate limit...")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AccessLogHandler 日志打印
func AccessLogHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		blw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		var requestBody string

		b, _ := c.GetRawData()
		requestBody = string(b)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(b))

		log.Print("AccessLogHandler brfore", zap.String("method", c.Request.Method),
			zap.String("request", requestBody))

		c.Next()
		log.Print("AccessLogHandler after", zap.String("data", fmt.Sprintf("url=%s, status=%d, resp=%s",
			c.Request.URL, c.Writer.Status(), blw.body.String())))
	}
}

// Logs 限制日志输出
func Logs() gin.HandlerFunc {
	return func(c *gin.Context) {
		gin.LoggerWithConfig(gin.LoggerConfig{
			Formatter: func(param gin.LogFormatterParams) string {
				//健康检查不输出
				if param.Path == "/" {
					return ""
				}
				var statusColor, methodColor, resetColor string
				if param.IsOutputColor() {
					statusColor = param.StatusCodeColor()
					methodColor = param.MethodColor()
					resetColor = param.ResetColor()
				}
				if param.Latency > time.Minute {
					param.Latency = param.Latency - param.Latency%time.Second
				}
				return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
					param.TimeStamp.Format("2006/01/02 - 15:04:05"),
					statusColor, param.StatusCode, resetColor,
					param.Latency,
					param.ClientIP,
					methodColor, param.Method, resetColor,
					param.Path,
					param.ErrorMessage,
				)
			},
			Output: gin.DefaultWriter,
		})
	}
}
