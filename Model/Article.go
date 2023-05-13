package Model

import (
	"BlogProject/Shares/errmsg"
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	Catalog Category `gorm:"foreignkey:Cid"`

	Title   string `gorm:"type:varchar(100);not null" json:"title"`
	Cid     int    `gorm:"type:int;not null" json:"cid"`
	Desc    string `gorm:"type:varchar(200)" json:"desc"`
	Content string `gorm:"type:longtext" json:"content"`
	Img     string `gorm:"type:longtext" json:"img"`

	PageView uint `gorm:"type:uint;defualt:0" json:"pv"`
}
type ArticleList []Article

// implement encoding.BinaryMarshaler for redis
func (a *Article) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}
func (a *ArticleList) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

// implement implementing BinaryUnmarshaler for redis
func (a *Article) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}
func (a *ArticleList) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

// check the validity
func CheckArticleUpdateValidity(requestData Article) int {
	var article Article
	sqlRes := db.First(&article, requestData.ID)
	if sqlRes.Error != nil || sqlRes.RowsAffected == 0 {
		log.Printf("Invalid article id:%v. ERROR: %v\n", requestData.ID, sqlRes.Error.Error())
		return errmsg.ERROR_ARTICLE_INVALID_ID
	}
	//  else if article.Title != requestData.Title {
	// 	log.Printf("Invalid article id or title.\n")
	// 	return errmsg.ERROR_ARTICLE_INVALID_ID_OR_TITLE
	// }
	return errmsg.SUCCESS
}

// update hook
func (article *Article) BeforeUpdate(tx *gorm.DB) (err error) {
	if CheckArticleUpdateValidity(*article) != errmsg.SUCCESS {
		err = errors.New("before update checked failed")
	}
	return
}

func AddArticle(data Article) int {
	result := db.Create(&data)
	if result.Error != nil {
		log.Println(result.Error.Error())
		return errmsg.ERROR
	}

	// add to redis cache
	RedisArticleSet("/controller/article/"+strconv.Itoa(int(data.ID)), data)
	// remove old redis cache
	cid := strconv.Itoa(data.Cid)
	RedisArticleDelKey("", cid)

	log.Printf("Create a new article: %v\n", data.Title)
	return errmsg.SUCCESS
}

// query Article
func GetArticles(redisKey string, pageSize int, pageNum int) ([]Article, int, int) {
	var articleList ArticleList = []Article{}

	offset := (pageNum - 1) * pageSize
	if pageNum <= -1 || pageSize <= -1 {
		offset = -1
	}
	if pageSize > 50 {
		pageSize = 50
	}
	var total_num int64
	db.Model(&Article{}).Count(&total_num)
	result := db.Preload("Catalog").Limit(pageSize).Offset(offset).Find(&articleList)
	if result.Error != nil {
		log.Println(result.Error.Error())
		return nil, errmsg.ERROR_ARTICLE_NOT_EXIST, 0
	}

	// cache request in redis
	field := "pagesize=" + strconv.Itoa(pageSize) + "pagenum=" + strconv.Itoa(pageNum)
	RedisArticleHset(redisKey, articleList, field)
	RedisArticleNumSet(int(total_num))

	log.Printf("Query articles success, rowAffected:%v \n", result.RowsAffected)
	return articleList, errmsg.SUCCESS, int(total_num)
}

// query specify article
func GetSingleArticle(id int) (Article, int) {
	var article Article
	result := db.Preload("Catalog").Where("id=?", id).First(&article)
	if result.Error != nil {
		log.Println(result.Error.Error())
		return Article{}, errmsg.ERROR_ARTICLE_NOT_EXIST
	}

	// cache request in redis
	var redisKey string = "/controller/article/" + strconv.Itoa(id)
	RedisArticleSet(redisKey, article)

	log.Printf("Query article success, rowAffected:%v \n", result.RowsAffected)
	return article, errmsg.SUCCESS
}

// query last 3 articles
func GetLast3Articles() ([]Article, int) {
	var articles ArticleList = []Article{}
	result := db.Preload("Catalog").Order("id desc").Limit(3).Find(&articles)
	if result.Error != nil {
		log.Println(result.Error.Error())
		return nil, errmsg.ERROR_ARTICLE_NOT_EXIST
	}

	// cache request in redis
	field := "last"
	RedisArticleHset("/controller/article/list/", articles, field)

	log.Printf("Query articles success, rowAffected:%v \n", result.RowsAffected)
	return articles, errmsg.SUCCESS
}
func GetHot3Articles() (ArticleList, int) {
	var articles ArticleList = []Article{}

	result := db.Order("page_view desc").Limit(3).Find(&articles)
	if result.Error != nil {
		log.Println(result.Error.Error())
		return nil, errmsg.ERROR_ARTICLE_NOT_EXIST
	}

	// cache request in redis
	field := "hot"
	RedisArticleHset("/controller/article/list/", articles, field)

	log.Printf("Query articles success, rowAffected:%v \n", result.RowsAffected)
	return articles, errmsg.SUCCESS
}

// query articles under specify catalog
func GetArticlesInSameCatalog(redisKey string, pageSize int, pageNum int, cid int) ([]Article, int, int) {
	var articles ArticleList = []Article{}

	offset := (pageNum - 1) * pageSize
	if pageNum <= -1 || pageSize <= -1 {
		offset = -1
	}
	if pageSize > 50 {
		pageSize = 50
	}
	// check cid
	checkResultCode := CheckCatalogById(cid)
	if checkResultCode != errmsg.SUCCESS {
		return nil, checkResultCode, 0
	}

	var total_num int64
	db.Preload("Catalog").Where("cid=?", cid).Find(&Article{}).Count(&total_num)
	result := db.Preload("Catalog").Where("cid=?", cid).Limit(pageSize).Offset(offset).Find(&articles)

	if result.Error != nil {
		log.Println(result.Error.Error())
		return nil, errmsg.ERROR_ARTICLE_NOT_EXIST, 0
	}

	// cache request in redis
	field := "pagesize=" + strconv.Itoa(pageSize) + "pagenum=" + strconv.Itoa(pageNum)
	RedisArticleHset(redisKey, articles, field)

	log.Printf("Query article success, rowAffected:%v \n", result.RowsAffected)
	return articles, errmsg.SUCCESS, int(total_num)

}

func RemoveArticle(id int) int {
	var article Article

	// get article cid from redis if key exists
	if err := RedisGetArticleById(strconv.Itoa(id), &article); err != errmsg.SUCCESS {
		log.Println("redis: ", errmsg.GetErrMsg(err))
		db.Select("cid").Where("id=?", id).First(&article)
	}

	result := db.Where("id=?", id).Delete(&Article{})
	if result.Error != nil || result.RowsAffected == 0 {
		log.Printf("cannot delete article. id:%v", id)
		return errmsg.ERROR
	}

	log.Printf("article: %v\n", article)
	// remove redis keys. cache will update after first query
	cid := strconv.Itoa(article.Cid)
	RedisArticleDelKey(strconv.Itoa(id), cid)

	return errmsg.SUCCESS
}

func UpdateArticle(data *Article) int {
	var article Article
	var maps = make(map[string]interface{})
	article.ID = data.ID

	maps["id"] = data.ID
	maps["title"] = data.Title
	maps["cid"] = data.Cid
	maps["desc"] = data.Desc
	maps["content"] = data.Content
	maps["img"] = data.Img

	result := db.Model(&article).Updates(maps)
	if result.Error != nil || result.RowsAffected == 0 {
		log.Println("Update article failed; ", result.Error)
		return errmsg.ERROR
	}

	// remove redis keys. cache will update after first query
	cid := strconv.Itoa(data.Cid)
	RedisArticleDelKey(strconv.Itoa(int(data.ID)), cid)
	return errmsg.SUCCESS
}

func UpdateArticlePv(values_sql string) int {

	db.Exec(`create temporary table if not exists tmp (
		id  int primary key,
		page_view int
	)`)
	db.Exec("insert into tmp (id,page_view) values " + values_sql + " on duplicate key update page_view=values(page_view)")
	res := db.Exec("update article,tmp set article.page_view = tmp.page_view where article.id = tmp.id")

	if err := res.Error; err != nil {
		log.Printf("err: %v\n", err)
	}
	return int(res.RowsAffected)
}
