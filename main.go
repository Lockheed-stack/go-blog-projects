package main

import (
	"BlogProject/Middlewares"
	"BlogProject/Model"
	"BlogProject/Routes"
	"flag"
	"time"
)

var DbAddress = flag.String("a", "localhost", "The MySql address")
var DbPort = flag.Int("port", 3306, "MySql port")
var DbPassword = flag.String("pwd", "", "MySql password")
var RedisMasterName = flag.String("Rname", "master", "The redis master name")
var JwtKey = flag.String("jwtkey", "your_jwt_key", "The jwt key")

func main() {
	flag.Parse()

	Model.InitDb(*DbAddress, *DbPort, *DbPassword)
	Model.InitRedis(*RedisMasterName)
	Middlewares.InitJWTkey(*JwtKey)

	// persisten relevant value
	tickerForRedisPersistencePv := time.NewTicker(time.Minute * 30)
	go Model.RedisPersistenceToDb(tickerForRedisPersistencePv)
	defer tickerForRedisPersistencePv.Stop()

	Routes.InitRouter()

}
