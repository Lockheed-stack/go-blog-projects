package errmsg

const (
	SUCCESS = 200
	ERROR   = 500

	// code: 1000+  user error
	ERROR_USERNAME_USED      = 1001
	ERROR_PASSWORD_WRONG     = 1002
	ERROR_USER_NOT_EXIST     = 1003
	ERROR_TOKEN_EXIST        = 1004
	ERROR_TOKEN_TIMEOUT      = 1005
	ERROR_TOKEN_WRONG        = 1006
	ERROR_TOKEN_FORMAT_WRONG = 1007
	ERROR_TOKEN_TYPE_WRONG   = 1008
	// code: 2000+ article error
	ERROR_ARTICLE_NOT_EXIST           = 2001
	ERROR_ARTICLE_INVALID_ID          = 2002
	ERROR_ARTICLE_INVALID_ID_OR_TITLE = 2003
	// code: 3000+ catalog error
	ERROR_CATALOG_USED         = 3001
	ERROR_CATALOG_NOT_EXIST    = 3002
	ERROR_CATALOG_INVALID_NAME = 3003
	ERROR_CATALOG_INVALID_CID  = 3004

	// code: 4000+ redis error
	ERROR_KEY_NOT_FOUND = 4001
)

var codemsg = map[int]string{
	SUCCESS: "OK",
	ERROR:   "FAIL",

	//user
	ERROR_USERNAME_USED:  "User name existed",
	ERROR_PASSWORD_WRONG: "password error",
	ERROR_USER_NOT_EXIST: "User doesn't exist",
	//user Token
	ERROR_TOKEN_EXIST:        "TOKEN doesn't exist",
	ERROR_TOKEN_TIMEOUT:      "TOKEN timeout",
	ERROR_TOKEN_WRONG:        "TOKEN wrong",
	ERROR_TOKEN_FORMAT_WRONG: "TOKEN Format wrong",
	ERROR_TOKEN_TYPE_WRONG:   "TOKEN type wrong",
	//catalog
	ERROR_CATALOG_USED:         "catalog name existed",
	ERROR_CATALOG_NOT_EXIST:    "catalog doesn't exist",
	ERROR_CATALOG_INVALID_NAME: "catalog name invalid",
	ERROR_CATALOG_INVALID_CID:  "catalog's cid is invalid",

	// article
	ERROR_ARTICLE_NOT_EXIST:           "article doesn't exist",
	ERROR_ARTICLE_INVALID_ID:          "article id is invalid",
	ERROR_ARTICLE_INVALID_ID_OR_TITLE: "article id or title is invalid",

	// redis
	ERROR_KEY_NOT_FOUND: "HKEY is not exist",
}

func GetErrMsg(code int) string {
	return codemsg[code]
}
