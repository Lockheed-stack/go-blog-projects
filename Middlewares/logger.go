package Middlewares

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {

	filePath := "log/ginBlog.log"
	src, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		log.Printf("fail to open %v, error: %v\n", filePath, err)
	}
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(src)

	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		duration := fmt.Sprintf("%.3f ms", float64(time.Since(startTime).Nanoseconds()/1e6))
		hostName, err := os.Hostname()
		if err != nil {
			hostName = "Unknown"
		}

		statusCode := ctx.Writer.Status()
		clientIp := ctx.ClientIP()
		userAgent := ctx.Request.UserAgent()
		method := ctx.Request.Method
		path := ctx.Request.RequestURI
		dataSize := ctx.Writer.Size()
		if dataSize < 0 {
			dataSize = 0
		}
		entry := logger.WithFields(logrus.Fields{
			"hostName": hostName,
			"status":   statusCode,
			"duration": duration,
			"IP":       clientIp,
			"method":   method,
			"path":     path,
			"dataSize": dataSize,
			"Agent":    userAgent,
		})

		if len(ctx.Errors) > 0 {
			entry.Error(ctx.Errors.ByType(gin.ErrorTypePrivate).String())
		}
		if statusCode >= 500 {
			entry.Error()
		} else if statusCode >= 400 {
			entry.Warn()
		} else {
			entry.Info()
		}
	}
}
