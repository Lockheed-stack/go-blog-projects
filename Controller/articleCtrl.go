package Controller

import (
	"BlogProject/Model"
	"BlogProject/Shares/errmsg"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddArticle(ctx *gin.Context) {
	var data Model.Article
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		log.Println("Create article error, ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    data,
			"message": err,
		})
	} else {
		code := Model.AddArticle(data)
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

func EditArticle(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var data Model.Article
	err := ctx.ShouldBindJSON(&data)
	data.ID = uint(id)

	if err != nil {
		log.Println("Binding Json error: ", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data":    data,
			"message": "Bad requset",
		})
	} else {
		code := Model.CheckCatalogById(data.Cid)
		if code != errmsg.SUCCESS {
			ctx.JSON(http.StatusOK, gin.H{
				"data":    data,
				"message": errmsg.GetErrMsg(code),
			})
		} else {
			articleCode := Model.UpdateArticle(&data)
			ctx.JSON(http.StatusOK, gin.H{
				"data":    data,
				"message": errmsg.GetErrMsg(articleCode),
			})
		}
	}
}

func RemoveArticle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Println("request param error")
		ctx.JSON(http.StatusOK, gin.H{
			"id":      id,
			"message": "bad requset",
		})
	} else {
		var code = Model.RemoveArticle(id)

		log.Printf("try to delete article, id: %v; result: %v \n", id, errmsg.GetErrMsg(code))
		ctx.JSON(http.StatusOK, gin.H{
			"status":  code,
			"message": errmsg.GetErrMsg(code),
		})
	}
}

func QueryArticles(ctx *gin.Context) {
	pageSize, err1 := strconv.Atoi(ctx.Query("pagesize"))
	pageNum, err2 := strconv.Atoi(ctx.Query("pagenum"))

	if err1 != nil || err2 != nil {
		log.Printf("Get catalog error: %v; \n %v;\n", err1, err2)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad requset",
		})
	} else {
		data, code, total_num := Model.GetArticles(ctx.Request.URL.Path, pageSize, pageNum)
		var HttpStatusCode int

		switch {
		case code == errmsg.ERROR_ARTICLE_NOT_EXIST:
			HttpStatusCode = 400
		default:
			HttpStatusCode = http.StatusOK
		}

		ctx.JSON(HttpStatusCode, gin.H{
			"data":    data,
			"total":   total_num,
			"message": errmsg.GetErrMsg(code),
		})
	}
}

func QueryLast3Articles(ctx *gin.Context) {
	data, code := Model.GetLast3Articles()
	var HttpStatusCode int

	switch {
	case code == errmsg.ERROR_ARTICLE_NOT_EXIST:
		HttpStatusCode = 400
	default:
		HttpStatusCode = http.StatusOK
	}

	ctx.JSON(HttpStatusCode, gin.H{
		"data":    data,
		"total":   3,
		"message": errmsg.GetErrMsg(code),
	})
}
func QueryHot3Articles(ctx *gin.Context) {
	data, code := Model.GetHot3Articles()
	var HttpStatusCode int

	switch {
	case code == errmsg.ERROR_ARTICLE_NOT_EXIST:
		HttpStatusCode = 400
	default:
		HttpStatusCode = http.StatusOK
	}

	ctx.JSON(HttpStatusCode, gin.H{
		"data":    data,
		"total":   3,
		"message": errmsg.GetErrMsg(code),
	})
}

func QuerySingleArticle(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("Get article error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad requset",
		})
	} else {
		data, code := Model.GetSingleArticle(id)
		var HttpStatusCode int

		switch {
		case code == errmsg.ERROR_ARTICLE_NOT_EXIST:
			HttpStatusCode = 400
		default:
			HttpStatusCode = http.StatusOK
		}

		ctx.JSON(HttpStatusCode, gin.H{
			"data":    data,
			"message": errmsg.GetErrMsg(code),
		})
	}
}
func QueryArticlesInSameCatalog(ctx *gin.Context) {
	cid, err := strconv.Atoi(ctx.Param("cid"))
	pageSize, err2 := strconv.Atoi(ctx.Query("pagesize"))
	pageNum, err3 := strconv.Atoi(ctx.Query("pagenum"))

	if err2 != nil || err3 != nil {
		log.Printf("Get articles error: %v; \n %v;\n", err2, err2)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad requset",
		})
	}
	if err != nil {
		log.Printf("No such catalog. Get article error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad requset",
		})
	} else {
		data, code, total_num := Model.GetArticlesInSameCatalog(ctx.Request.URL.Path, pageSize, pageNum, cid)
		var HttpStatusCode int

		switch {
		case code == errmsg.ERROR_ARTICLE_NOT_EXIST:
			HttpStatusCode = 400
		default:
			HttpStatusCode = http.StatusOK
		}

		ctx.JSON(HttpStatusCode, gin.H{
			"data":    data,
			"total":   total_num,
			"message": errmsg.GetErrMsg(code),
		})
	}
}
