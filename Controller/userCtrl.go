package Controller

import (
	"BlogProject/Model"
	"BlogProject/Shares/errmsg"
	"BlogProject/Shares/validator"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUser(ctx *gin.Context) {
	pageSize, err1 := strconv.Atoi(ctx.Query("pagesize"))
	pageNum, err2 := strconv.Atoi(ctx.Query("pagenum"))
	if err1 != nil || err2 != nil {
		log.Printf("Get user error: %v; \n %v;\n", err1, err2)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad requset",
		})
	} else {
		if pageSize == 0 {
			pageSize = -1
		}
		if pageNum == 0 {
			pageNum = -1
		}

		var data, total_num = Model.GetUsers(pageSize, pageNum)
		var code int
		if data != nil {
			code = errmsg.SUCCESS
		} else {
			code = errmsg.ERROR_USER_NOT_EXIST
		}
		ctx.JSON(http.StatusOK, gin.H{
			"status":  code,
			"data":    data,
			"total":   total_num,
			"message": errmsg.GetErrMsg(code),
		})
	}
}

func AddUser(ctx *gin.Context) {
	var data Model.User
	err := ctx.ShouldBindJSON(&data)
	msg, code := validator.Validate(&data)

	if err != nil || code != errmsg.SUCCESS {
		log.Println("Binding Json error: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    data,
			"message": msg,
			"error":   err,
		})
	} else {
		code := Model.CheckUserExist(data.Username)
		if code == errmsg.SUCCESS {
			create_code := Model.CreateUser(&data)
			if create_code != errmsg.SUCCESS {
				log.Printf("cannot create user %v\n", data.Username)
				ctx.JSON(http.StatusBadRequest, gin.H{
					"status":  create_code,
					"data":    data,
					"message": errmsg.GetErrMsg(create_code),
				})
			} else {
				ctx.JSON(http.StatusOK, gin.H{
					"status":  create_code,
					"data":    data,
					"message": errmsg.GetErrMsg(create_code),
				})
			}

		} else {
			log.Println(errmsg.GetErrMsg(code))
			ctx.JSON(http.StatusOK, gin.H{
				"status":  code,
				"data":    data,
				"message": errmsg.GetErrMsg(code),
			})
		}
	}

}

// func EditUser(ctx *gin.Context) {
// 	var data Model.User
// 	err := ctx.ShouldBindJSON(&data)
// 	if err != nil {
// 		log.Println("Binding Json error: ", err)
// 		ctx.JSON(http.StatusBadRequest, gin.H{
// 			"data":    data,
// 			"message": err,
// 		})
// 	} else {

// 	}

// }

func RemoveUser(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {
		log.Println("request param error")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"id":      id,
			"message": "bad requset",
		})
	} else {
		var code = Model.RemoveUser(id)
		var httpStatusCode int
		switch {
		case code == 200:
			httpStatusCode = http.StatusOK
		case code == 500:
			httpStatusCode = http.StatusBadRequest
		default:
			httpStatusCode = http.StatusForbidden
		}
		ctx.JSON(httpStatusCode, gin.H{
			"status":  code,
			"message": errmsg.GetErrMsg(code),
		})
	}
}
