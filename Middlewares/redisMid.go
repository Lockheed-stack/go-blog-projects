package Middlewares

import (
	"BlogProject/Model"
	"BlogProject/Shares/errmsg"
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RedisCatalogs() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		client := Model.GetRedisReplicaQueue().GetReplicaClient()
		rdb_ctx := context.Background()

		field := "pagenum=" + ctx.Query("pagenum") + "pagesize=" + ctx.Query("pagesize")
		key := ctx.Request.URL.Path

		var data interface{}
		var total_num int = 0
		var redisHgetErr error

		switch key {
		case "/controller/categories/list":
			tmp := Model.CategoryArticleList{}
			redisHgetErr = client.HGet(rdb_ctx, key, field).Scan(&tmp)
			total_num = len(tmp)
			data = tmp
		case "/controller/categories":
			tmp := Model.CategoryList{}
			redisHgetErr = client.HGet(rdb_ctx, key, field).Scan(&tmp)
			s, _ := client.Get(rdb_ctx, "catalog_total_num").Result()
			total_num, _ = strconv.Atoi(s)
			data = tmp
		default:
			ctx.JSON(http.StatusOK, gin.H{
				"message": "bad request",
			})
			ctx.Abort()
		}

		if redisHgetErr == redis.Nil {
			log.Printf("redis: no such key \"%v\"\n", key)
		} else if redisHgetErr != nil {
			log.Printf("redis err: %v\n", redisHgetErr)
			ctx.JSON(http.StatusOK, gin.H{
				"message": "bad request",
			})
			ctx.Abort()
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"data":    data,
				"total":   total_num,
				"message": errmsg.GetErrMsg(errmsg.SUCCESS),
			})
			ctx.Abort()
		}

	}
}

func RedisArticles() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		client := Model.GetRedisReplicaQueue().GetReplicaClient()
		rdb_ctx := context.Background()

		key := ctx.Request.URL.Path
		queryPagesize := ctx.Query("pagesize")
		queryPagenum := ctx.Query("pagenum")

		if ctx.Param("id") != "" {
			var data Model.Article

			err := client.Get(rdb_ctx, key).Scan(&data)

			if err == redis.Nil {
				log.Printf("redis: no such key \"%v\"\n", key)
			} else if err != nil {
				log.Println(err)
			} else {
				ctx.JSON(http.StatusOK, gin.H{
					"data":    data,
					"message": errmsg.GetErrMsg(errmsg.SUCCESS),
				})
				go RedisCheckAndSetPV(ctx.Param("id"), ctx.ClientIP())

				ctx.Abort()
			}
		} else {

			data := Model.ArticleList{}

			if queryPagenum != "" && queryPagesize != "" {
				// query article (or article under same catalog) list
				field := "pagesize=" + queryPagesize + "pagenum=" + queryPagenum
				err := client.HGet(rdb_ctx, key, field).Scan(&data)
				var total_num int = 0

				if ctx.Param("cid") != "" { //article under same catalog
					total_num = len(data)
				} else {
					s, err := client.Get(rdb_ctx, "article_total_num").Result()
					if err == redis.Nil || err != nil {
						log.Println(err)
					}
					total_num, _ = strconv.Atoi(s)
				}

				if err == redis.Nil {
					log.Printf("redis Hget err: no such key \"%v\" or field \"%v\"\n", key, field)
				} else if err != nil {
					log.Println(err)
				} else {
					ctx.JSON(http.StatusOK, gin.H{
						"data":    data,
						"total":   total_num,
						"message": errmsg.GetErrMsg(errmsg.SUCCESS),
					})
					ctx.Abort()
				}
			} else {
				// query last or hot articles
				var field string = ""
				switch key {
				case "/controller/article/list/last":
					{
						field = "last"
					}
				case "/controller/article/list/hot":
					{
						field = "hot"
					}
				default:
					// bad request
					ctx.JSON(http.StatusOK, gin.H{
						"message": "bad request",
					})
					ctx.Abort()
					return
				}

				err := client.HGet(rdb_ctx, "/controller/article/list/", field).Scan(&data)
				if err == redis.Nil {
					log.Printf("redis Hget err: no such key \"%v\" or field \"%v\"\n", key, field)
				} else if err != nil {
					log.Println(err)
				} else {
					ctx.JSON(http.StatusOK, gin.H{
						"data":    data,
						"total":   len(data),
						"message": errmsg.GetErrMsg(errmsg.SUCCESS),
					})
					ctx.Abort()
				}
			}
		}

	}
}

func RedisCheckAndSetPV(id string, ip string) (errcode int) {

	if id != "" {

		rdb := Model.GetRedisMasterClient()
		rdb_ctx := context.Background()

		// check ip
		b, err2 := rdb.HExists(rdb_ctx, ip, id).Result()
		if err2 != nil {
			log.Println(err2)
			return errmsg.ERROR
		}
		if !b { // key-field pair 'ip:id' doesn't exist

			article := Model.Article{}
			pipe := rdb.Pipeline()

			err := rdb.Get(rdb_ctx, "/controller/article/"+id).Scan(&article)
			if err != nil {
				log.Println(err)
			} else {
				article.PageView++
				pipe.Set(rdb_ctx, "/controller/article/"+id, &article, 0)
			}

			pv, _ := pipe.Incr(rdb_ctx, id).Result()
			pipe.HSet(rdb_ctx, ip, id, pv).Result()
			pipe.Expire(rdb_ctx, ip, time.Hour*24).Result()
			_, err3 := pipe.Exec(rdb_ctx)
			if err3 != nil {
				log.Println(err3)
				return errmsg.ERROR
			}
		}
		return errmsg.SUCCESS
	}
	return errmsg.ERROR
}
