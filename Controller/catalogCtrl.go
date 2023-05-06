package Controller

import (
	"BlogProject/Model"
	"BlogProject/Shares/errmsg"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddCatalog(ctx *gin.Context) {
	var data Model.Category
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		log.Println("Create catalog error, ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    data,
			"message": err,
		})
	} else {
		code := Model.AddCatalog(&data)
		if code != errmsg.SUCCESS {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"data":    data,
				"message": errmsg.GetErrMsg(code),
			})
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"data":    data,
				"message": errmsg.GetErrMsg(code),
			})
		}
	}
}

func EditCatalog(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var data Model.Category
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		log.Println("Binding Json error: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    data,
			"message": "Bad requset",
		})
	} else {
		code := Model.UpdateCatalog(id, &data)
		var httpStatusCode int
		switch {
		case code == errmsg.SUCCESS:
			httpStatusCode = 200
		default:
			httpStatusCode = 400
		}
		ctx.JSON(httpStatusCode, gin.H{
			"data":    data,
			"message": errmsg.GetErrMsg(code),
		})
	}
}

func RemoveCatalog(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Println("request param error")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"id":      id,
			"message": "bad requset",
		})
	} else {
		var code = Model.RemoveCatalog(id)
		var httpStatusCode int
		switch {
		case code == 200:
			httpStatusCode = http.StatusOK
		case code == 500:
			httpStatusCode = http.StatusBadRequest
		default:
			httpStatusCode = http.StatusForbidden
		}
		log.Printf("try to delete catalog, id: %v; result: %v \n", id, errmsg.GetErrMsg(code))
		ctx.JSON(httpStatusCode, gin.H{
			"status":  code,
			"message": errmsg.GetErrMsg(code),
		})
	}
}

func QueryCatalog(ctx *gin.Context) {
	pageSize, err1 := strconv.Atoi(ctx.Query("pagesize"))
	pageNum, err2 := strconv.Atoi(ctx.Query("pagenum"))

	if err1 != nil || err2 != nil {
		log.Printf("Get catalog error: %v; \n %v;\n", err1, err2)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad requset",
		})
	} else {
		data, total_num := Model.GetCatalogs(pageSize, pageNum)
		var (
			code, HttpStatusCode int
		)

		switch {
		case data == nil:
			code = 500
			HttpStatusCode = 400
		default:
			code = 200
			HttpStatusCode = http.StatusOK
		}

		ctx.JSON(HttpStatusCode, gin.H{
			"data":    data,
			"total":   total_num,
			"message": errmsg.GetErrMsg(code),
		})
	}

}

func QueryAllCategoriesWithAllArticles(ctx *gin.Context) {
	pageSize, err1 := strconv.Atoi(ctx.Query("pagesize"))
	pageNum, err2 := strconv.Atoi(ctx.Query("pagenum"))

	if err1 != nil || err2 != nil {
		log.Printf("Get catalog error: %v; \n %v;\n", err1, err2)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad requset",
		})
	} else {
		data, total_num := Model.GetAllCatelogsWithAllArticles(pageSize, pageNum)
		var (
			code, HttpStatusCode int
		)

		switch {
		case data == nil:
			code = 500
			HttpStatusCode = 400
		default:
			code = 200
			HttpStatusCode = http.StatusOK
		}

		ctx.JSON(HttpStatusCode, gin.H{
			"data":    data,
			"total":   total_num,
			"message": errmsg.GetErrMsg(code),
		})
	}
}
