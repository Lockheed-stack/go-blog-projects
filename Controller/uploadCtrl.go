package Controller

import (
	"BlogProject/Shares/errmsg"
	"BlogProject/Shares/upload"
	"log"

	"github.com/gin-gonic/gin"
)

func Upload(ctx *gin.Context) {
	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Println("err", err)
		return
	}
	fileSize := fileHeader.Size

	log.Println(file, fileSize)

	url, code := upload.UploadFile(file, fileSize)
	var httpStatusCode int
	switch {
	case code == errmsg.ERROR:
		httpStatusCode = 400
	default:
		httpStatusCode = 200
	}
	ctx.JSON(httpStatusCode, gin.H{
		"status":  code,
		"message": errmsg.GetErrMsg(code),
		"url":     url,
	})
}
