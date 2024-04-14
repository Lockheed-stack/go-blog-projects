package Middlewares

import (
	"BlogProject/Shares/errmsg"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JwtKey []byte

const (
	AccessTokenExpiredDuration  = 2 * time.Hour
	RefreshTokenExpiredDuration = 30 * 24 * time.Hour
	TokenIssuer                 = ""
)

type MyClaims struct {
	Username string `json:"username"`
	// Password string `json:"password"`
	jwt.RegisteredClaims
}

func GetJWTtime(t time.Duration) *jwt.NumericDate {
	return jwt.NewNumericDate(time.Now().Add(t))
}

// init JwtKey
func InitJWTkey(jwtkey string) {
	JwtKey = []byte(jwtkey)
}

// generate token
func GenerateToken(username string, password string) (string, int) {

	RegisteredClaims := jwt.RegisteredClaims{
		ExpiresAt: GetJWTtime(AccessTokenExpiredDuration),
		Issuer:    "ginBlog",
	}
	SetClaims := MyClaims{
		username,
		// password,
		RegisteredClaims,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, SetClaims)
	signedString, err := accessToken.SignedString(JwtKey)
	if err != nil {
		log.Println("Token signed error: ", err)
		return "", errmsg.ERROR_TOKEN_WRONG
	}
	return signedString, errmsg.SUCCESS
}

// authenticate token
func AuthToken(token string) (*MyClaims, int) {
	parseToken, err := jwt.ParseWithClaims(token, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	if err != nil {
		return nil, errmsg.ERROR_TOKEN_WRONG
	}

	if claim, ok := parseToken.Claims.(*MyClaims); ok && parseToken.Valid {
		log.Println("Token is valid: ", claim)
		return claim, errmsg.SUCCESS
	} else {
		log.Println("Fail to authenticate token. ", err)
		return nil, errmsg.ERROR_TOKEN_WRONG
	}
}

// jwt middleware
func JwtToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		tokenHeader := ctx.Request.Header.Get("Authorization")

		if tokenHeader == "" {
			log.Println("invalid token header")
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"tokenCode": errmsg.ERROR,
				"message":   errmsg.GetErrMsg(errmsg.ERROR),
			})
			ctx.Abort()
			return
		}

		checkToken := strings.Split(tokenHeader, " ")
		log.Println(checkToken)
		if len(checkToken) != 2 && checkToken[0] != "Bearer" {
			log.Println(errmsg.GetErrMsg(errmsg.ERROR_TOKEN_TYPE_WRONG))
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"tokenCode": errmsg.ERROR_TOKEN_TYPE_WRONG,
				"message":   errmsg.GetErrMsg(errmsg.ERROR_TOKEN_TYPE_WRONG),
			})
			ctx.Abort()
			return
		}

		key, tokenCode := AuthToken(checkToken[1])
		if tokenCode != errmsg.SUCCESS {
			log.Println(errmsg.GetErrMsg(tokenCode))
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"tokenCode": errmsg.ERROR_TOKEN_WRONG,
				"message":   errmsg.GetErrMsg(errmsg.ERROR_TOKEN_WRONG),
			})
			ctx.Abort()
			return
		} else {
			if time.Now().Unix() > key.ExpiresAt.Unix() {
				log.Println(errmsg.GetErrMsg(errmsg.ERROR_TOKEN_TIMEOUT))
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"tokenCode": tokenCode,
					"message":   errmsg.GetErrMsg(tokenCode),
				})
				ctx.Abort()
				return
				// } else {
				// 	ctx.JSON(http.StatusOK, gin.H{
				// 		"tokenCode": tokenCode,
				// 		"message":   errmsg.GetErrMsg(tokenCode),
				// 	})
			}
		}
		ctx.Set("username", key.Username)
		ctx.Next()
	}
}
