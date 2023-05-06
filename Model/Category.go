package Model

import (
	"BlogProject/Shares/errmsg"
	"encoding/json"
	"errors"
	"log"

	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name string `gorm:"type:varchar(20);not null" json:"name"`
}
type CategoryWithArticle struct {
	Name  string `json:"cname"`
	Cid   int    `json:"cid"`
	Title string `json:"title"`
	Id    int    `json:"id"`
}
type CategoryList []Category
type CategoryArticleList []CategoryWithArticle

// implement encoding.BinaryMarshaler for redis
func (c *Category) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}
func (c *CategoryList) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}
func (c *CategoryWithArticle) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}
func (c *CategoryArticleList) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// implement implementing BinaryUnmarshaler for redis
func (c *Category) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}
func (c *CategoryList) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}
func (c *CategoryWithArticle) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}
func (c *CategoryArticleList) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// check the validity
func CheckCatalog(name string) int {
	var catalog Category
	db.Select("id").Where("name=?", name).First(&catalog)
	if catalog.ID > 0 {
		return errmsg.ERROR_CATALOG_USED
	}
	return errmsg.SUCCESS // name doesn't exist
}
func CheckCatalogById(id int) int {
	var catalog Category
	catalog_check := db.First(&catalog, id)
	if catalog_check.Error != nil || catalog_check.RowsAffected == 0 {
		log.Printf("Invalid cid. Error: %v", catalog_check.Error.Error())
		return errmsg.ERROR_CATALOG_NOT_EXIST
	}
	return errmsg.SUCCESS
}
func CheckCatalogByIdAndName(id uint, name string) int {
	var catalog []Category

	sqlRes := db.Select("id").Where("name=?", name).Find(&catalog)
	if sqlRes.Error != nil || len(catalog) > 1 || sqlRes.RowsAffected == 0 {
		log.Printf("Invalid cid or catalog name. ERROR: %v\n", sqlRes.Error.Error())
		return errmsg.ERROR_CATALOG_INVALID_NAME
	} else if len(catalog) == 1 && catalog[0].ID != id {
		return errmsg.ERROR_CATALOG_INVALID_CID
	}

	return errmsg.SUCCESS
}

// create hook
func (catalog *Category) BeforeCreate(tx *gorm.DB) (err error) {
	var code = CheckCatalog(catalog.Name)
	if code != errmsg.SUCCESS {
		return errors.New(errmsg.GetErrMsg(code))
	}
	return nil
}

func AddCatalog(data *Category) int {
	result := db.Create(data)
	if result.Error != nil {
		log.Println(result.Error.Error())
		return errmsg.ERROR
	}

	// remove redis keys. cache will update after first query
	RedisCatalogDelKey()

	log.Printf("Create a new catalog: %v\n", data.Name)
	return errmsg.SUCCESS
}

func GetCatalogs(pageSize int, pageNum int) ([]Category, int) {
	var catalogs CategoryList = []Category{}

	offset := (pageNum - 1) * pageSize
	if pageNum <= -1 || pageSize <= -1 {
		offset = -1
	}
	if pageSize > 50 {
		pageSize = 50
	}

	var total_num int64
	db.Model(&Category{}).Count(&total_num)
	result := db.Limit(pageSize).Offset(offset).Find(&catalogs)
	if result.Error != nil {
		log.Println(result.Error.Error())
		return nil, 0
	}

	// cache request in redis
	RedisCatalogHset(catalogs, pageSize, pageNum)
	RedisCatalogNumSet(int(total_num))

	log.Printf("Query category success, rowAffected:%v \n", result.RowsAffected)
	return catalogs, int(total_num)
}

func GetAllCatelogsWithAllArticles(pageSize int, pageNum int) ([]CategoryWithArticle, int) {
	var result CategoryArticleList = []CategoryWithArticle{}
	offset := (pageNum - 1) * pageSize
	if pageNum <= -1 || pageSize <= -1 {
		offset = -1
	}
	if pageSize > 50 {
		pageSize = 50
	}

	var total int64
	db.Model(&Category{}).
		Joins("left join article on (category.id=article.cid and article.deleted_at is null)").
		Count(&total)

	sqlRes := db.Limit(pageSize).Offset(offset).Model(&Category{}).
		Select("category.name,category.id as cid,article.title,article.id").
		Joins("left join article on category.id=article.cid and article.deleted_at is null").Scan(&result)

	if sqlRes.Error != nil {
		log.Println(sqlRes.Error.Error())
		return nil, 0
	}

	//cache request in redis
	RedisCatalogHset(result, pageSize, pageNum)

	log.Printf("Query all category with all articles success, rowAffected: %v\n", sqlRes.RowsAffected)
	return result, int(total)
}

func RemoveCatalog(id int) int {
	var catalog Category
	result := db.Where("id=?", id).Delete(&catalog)
	if result.Error != nil || result.RowsAffected == 0 {
		log.Printf("cannot delete catalog. id:%v", id)
		return errmsg.ERROR
	}

	// remove redis keys. cache will update after first query
	RedisCatalogDelKey()
	return errmsg.SUCCESS
}

// update hook
// func (catalog *Category) BeforeUpdate(tx *gorm.DB) (err error) {
// 	code := CheckCatalog(catalog.Name)

// 	// New catalog name is avaliable
// 	if code == errmsg.SUCCESS {
// 		return nil
// 	}
// 	return errors.New(errmsg.GetErrMsg(code))
// }

func UpdateCatalog(id int, data *Category) int {
	var catalog Category
	var maps = make(map[string]interface{})
	maps["name"] = data.Name

	code := CheckCatalogByIdAndName(uint(id), data.Name)
	if code != errmsg.SUCCESS {
		log.Println("Catalog Check failed; ", errmsg.GetErrMsg(code))
		return errmsg.ERROR
	} else {
		result := db.Model(&catalog).Where("id=?", id).Updates(maps)
		if result.Error != nil || result.RowsAffected == 0 {
			log.Println("Update catalog failed; ", result.Error)
			return errmsg.ERROR
		}
	}

	// remove redis keys. cache will update after first query
	RedisCatalogDelKey()
	return errmsg.SUCCESS
}
