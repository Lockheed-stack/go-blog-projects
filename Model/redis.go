package Model

import (
	"BlogProject/Shares/errmsg"
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx context.Context
var masterName string
var rdb *redis.Client
var sentinelClient *redis.SentinelClient
var replicas ReplicaQueue

// ----------------------------- ReplicatRedis struct ---------------------------------
type ReplicaRedis struct {
	Addr string
	Rdb  *redis.Client
}
type ReplicaQueue []ReplicaRedis

func (r *ReplicaQueue) ReplicaQueueInit(master_name string) {
	slaves, _ := sentinelClient.Replicas(ctx, master_name).Result()
	for _, s := range slaves {
		addr := s["name"]
		tmp := redis.NewClient(&redis.Options{
			Addr: addr,
			DB:   0,
		})
		_, err := tmp.Ping(ctx).Result()
		if err != nil {
			log.Printf("failed to connect replica: %v\n", addr)
		} else {
			*r = append(*r, ReplicaRedis{
				Addr: addr,
				Rdb:  tmp,
			})
		}
	}
	log.Println("replica rdb info: ", replicas)
}

func (r *ReplicaQueue) replicaQueuePop() (replicaInfo ReplicaRedis) {
	replicaInfo = ReplicaRedis{}

	if len(*r) == 0 {
		return replicaInfo
	}

	replicaInfo.Addr = (*r)[0].Addr
	replicaInfo.Rdb = (*r)[0].Rdb
	*r = (*r)[1:]
	return replicaInfo
}

func (r *ReplicaQueue) replicaQueuePush(replicaInfo ReplicaRedis) {
	*r = append(*r, replicaInfo)
}

func (r *ReplicaQueue) GetReplicaClient() *redis.Client {
	if len(*r) == 0 {
		log.Println("redis: no available replica client, try to search...")
		r.ReplicaQueueInit(masterName)

		if len(*r) == 0 {
			log.Println("redis: failed to find available replica client, return master client")
			return rdb
		}
	}

	// get replica
	replicaInfo := r.replicaQueuePop()
	// check replica
	s, err := replicaInfo.Rdb.Do(ctx, "info", "replication").Result()
	if err != nil {
		// failed to get slave
		log.Println(err)
		return rdb
	}

	res := s.(string)
	role := strings.Split(strings.Split(res, "\r")[1], ":")[1]
	if role != "slave" { // the master offline and slave become the new master
		// try to find another available replica client
		r.ReplicaQueueInit(masterName)
	} else {
		// put it back to queue
		r.replicaQueuePush(replicaInfo)
	}
	return replicaInfo.Rdb
}

// ----------------------------------------- end ----------------------------------------

func InitRedis(master_Name string) {

	masterName = master_Name

	ctx = context.Background()
	// --------------------------------------------------
	// init redis client
	rdb = redis.NewFailoverClient(
		&redis.FailoverOptions{
			MasterName:    masterName, // 去 sentinel.conf 配置文件中查看
			SentinelAddrs: []string{"redis-sentinel-1:26379"},
			DB:            0,
			PoolFIFO:      true,
		},
	)
	pong, err := rdb.Ping(ctx).Result()

	if err != nil {
		log.Fatalf("failed to connect redis, err:%v\n", err)
	} else {
		log.Printf("Redis Connected: %v\n", pong)
	}
	// --------------------------------------------------
	// init sentinel client
	sentinelClient = redis.NewSentinelClient(
		&redis.Options{
			Addr: "redis-sentinel-1:26379",
		},
	)
	_, err2 := sentinelClient.Ping(ctx).Result()
	if err2 != nil {
		log.Fatalln(err2)
	}
	log.Printf("sentinel client connected")

	// --------------------------------------------------
	// init replication redis addr
	replicas.ReplicaQueueInit(masterName)

	// --------------------------------------------------
	// init page value
	articles, errCode, _ := GetArticles("", -1, -1)
	if errCode != errmsg.SUCCESS {
		log.Fatalln("redis: failed to init page view")
	}

	var id_pv []string
	for _, a := range articles {
		key := strconv.Itoa(int(a.ID))
		value := strconv.Itoa(int(a.PageView))
		id_pv = append(id_pv, key, value)
	}
	s, err3 := rdb.MSet(ctx, id_pv).Result()
	if err3 != nil {
		log.Println(err3)
	}
	log.Println(s)
}

func GetRedisMasterClient() *redis.Client {
	return rdb
}
func GetRedisReplicaQueue() *ReplicaQueue {
	return &replicas
}

// redis catalog relevant
func RedisCatalogHset(data interface{}, pageSize int, pageNum int) {

	field := "pagenum=" + strconv.Itoa(pageNum) + "pagesize=" + strconv.Itoa(pageSize)
	switch t := data.(type) {
	case CategoryList:
		{
			key := "/controller/categories"
			_, err := rdb.HSet(ctx, key, map[string]interface{}{
				field: &t,
			}).Result()
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("redis: Hset key \"%v\" successfully\n", key)
			}
		}
	case CategoryArticleList:
		{
			key := "/controller/categories/list"
			_, err := rdb.HSet(ctx, key, map[string]interface{}{
				field: &t,
			}).Result()
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("redis: Hset key \"%v\" successfully\n", key)
			}
		}
	}
}
func RedisCatalogDelKey() {
	_, err := rdb.Del(ctx, "/controller/categories", "/controller/categories/list").Result()
	if err != nil {
		log.Printf("redis: failed to delete keys")
	}
}
func RedisCatalogNumSet(total_num int) {
	if err := rdb.Set(ctx, "catalog_total_num", total_num, 0).Err(); err != nil {
		log.Println(err)
	}
}

// func RedisCatalogNumGet() int {
// 	s, err := rdb.Get(ctx, "catalog_total_num").Result()
// 	if err == redis.Nil || err != nil {
// 		log.Println(err)
// 		return -1
// 	}
// 	i, err2 := strconv.Atoi(s)
// 	if err2 != nil {
// 		log.Printf("redis: failed to convert \"%v\" to int\n", s)
// 		return -1
// 	}
// 	return i
// }

// redis article relevant
func RedisArticleSet(key string, data Article) {
	if key == "" {
		log.Println("redis: Skip to set because key is nil")
		return
	}

	_, err := rdb.Set(ctx, key, &data, 0).Result()
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("redis Set key \"%v\" successfully\n", key)

	result, err2 := rdb.SetNX(ctx, strconv.Itoa(int(data.ID)), data.PageView+1, 0).Result()
	if err2 != nil {
		log.Println(result)
	}
	if !result { // key 'id' exists
		rdb.Incr(ctx, key)
	}

}

func RedisArticleHset(key string, data ArticleList, field string) {

	if key == "" {
		log.Println("redis: Skip to Hset because key is nil")
		return
	}

	_, err := rdb.HSet(ctx, key, map[string]interface{}{
		field: &data,
	}).Result()

	if err != nil {
		log.Println(err)
	} else {
		log.Printf("redis: Hset key \"%v\" successfully\n", key)
		if field == "hot" { // set expire time for hottest articles
			_, err2 := rdb.Expire(ctx, key, time.Hour*6).Result()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Printf("redis: key \"%v\" will expire after 6 hour", key)
			}
		}
	}

}

func RedisGetArticleById(id string, data *Article) int {
	if err := rdb.Get(ctx, id).Scan(data); err != nil {
		log.Println(err)
		return errmsg.ERROR_KEY_NOT_FOUND
	}
	return errmsg.SUCCESS
}

func RedisArticleDelKey(id string, cid string) {

	var keys []string

	switch cid {
	case "":
		keys = []string{
			"/controller/article/list",
			"/controller/article/list/",
			"/controller/article/" + id,
			id,
		}

	default:
		keys = []string{
			"/controller/article/list",
			"/controller/article/list/",
			"/controller/article/" + id,
			"/controller/article/list/" + cid,
			"/controller/categories",
			"/controller/categories/list",
		}
	}

	_, err := rdb.Del(ctx, keys...).Result()
	if err != nil {
		log.Println(err)
	}
}

func RedisArticleNumSet(total_num int) {
	if err := rdb.Set(ctx, "article_total_num", total_num, 0).Err(); err != nil {
		log.Println(err)
	}
}

// func RedisArticleNumGet() int {
// 	s, err := rdb.Get(ctx, "article_total_num").Result()
// 	if err == redis.Nil || err != nil {
// 		log.Println(err)
// 		return -1
// 	}
// 	i, err2 := strconv.Atoi(s)
// 	if err2 != nil {
// 		log.Printf("redis: failed to convert \"%v\" to int\n", s)
// 		return -1
// 	}
// 	return i
// }

// persistence pv
func RedisPersistenceToDb(t *time.Ticker) {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	for range t.C {
		articles, errCode, total_num := GetArticles("", -1, -1)
		if errCode != errmsg.SUCCESS {
			log.Println("failed to get articles")
		} else {

			pipe := rdb.Pipeline()
			var values_sql string

			for _, a := range articles {
				s, err := rdb.Get(ctx, strconv.Itoa(int(a.ID))).Result()
				if err == redis.Nil || err != nil {
					log.Printf("redis: no such article id:\"%v\"", a.ID)
				} else {
					pv, err2 := strconv.Atoi(s)
					if err2 != nil {
						log.Println(err2)
					} else {

						// update article pv in redis
						var redisKey string = "/controller/article/" + strconv.Itoa(int(a.ID))
						tmp := a
						tmp.PageView = uint(pv)

						pipe.Set(ctx, redisKey, &tmp, 0)

						// concat sql
						values_sql += "(" + strconv.Itoa(int(a.ID)) + "," + strconv.Itoa(pv) + "),"
					}
				}
			}

			// exec sql
			affected_row := UpdateArticlePv(values_sql[:len(values_sql)-1])
			// update hottest article in next request and exec redis pipe
			pipe.Del(ctx, "/controller/article/list/hot")
			_, err := pipe.Exec(ctx)
			if err != nil {
				log.Println(err)
			}
			log.Printf("PV persistence finished. Total article:%v ; %v articles get new value\n", total_num, affected_row)
		}
	}
}
