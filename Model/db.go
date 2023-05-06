package Model

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

// connect db
func InitDb(dbAddr string, port int, password string) {

	dsn := "root:" + password + "@tcp(" + dbAddr + ":" + strconv.Itoa(port) + ")/forBlog?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		},
	})

	if err != nil {
		fmt.Println("failed to connented DB", err)
	}

	db.AutoMigrate(&User{}, &Category{}, &Article{})

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(10 * time.Second)

}
