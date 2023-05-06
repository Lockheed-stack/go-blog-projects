package Controller

import (
	"BlogProject/Middlewares"
	"BlogProject/Model"
	"BlogProject/Shares/errmsg"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	var data Model.User
	ctx.ShouldBindJSON(&data)

	code := Model.AuthLogin(data.Username, data.Password)

	if code == errmsg.SUCCESS {
		token, tokenCode := Middlewares.GenerateToken(data.Username, data.Password)
		if tokenCode != errmsg.SUCCESS {
			ctx.JSON(http.StatusOK, gin.H{
				"status":  tokenCode,
				"message": errmsg.GetErrMsg(code),
				"token":   token,
			})
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"status":  tokenCode,
				"message": errmsg.GetErrMsg(code),
				"token":   token,
			})
		}
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Bad Request",
			"token":   "",
		})
	}

}

func TokenCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  200,
		"message": errmsg.GetErrMsg(errmsg.SUCCESS),
	})
}
