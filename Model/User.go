package Model

import (
	"BlogProject/Shares/encryption"
	"BlogProject/Shares/errmsg"
	"errors"
	"log"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(20);not null" json:"username" validate:"required,min=4,max=12" label:"用户名"`
	Password string `gorm:"type:varchar(200);not null" json:"password" validate:"required,min=6,max=30" label:"密码"`
	Role     int    `gorm:"type:int;DEFAULT=2" json:"role" validate:"required,gte=1" label:"权限"`
}

func CheckUserExist(username string) int {
	var data User
	db.Select("id").Where("username=?", username).First(&data)
	if data.ID > 0 {
		return errmsg.ERROR_USERNAME_USED //1001
	}
	return errmsg.SUCCESS
}

func CreateUser(data *User) int {
	unEncryptionPwd := data.Password
	data.Password = encryption.ScryptPassword(data.Password)
	result := db.Create(data)

	if result.Error != nil {
		data.Password = unEncryptionPwd
		return errmsg.ERROR // 500
	}
	return errmsg.SUCCESS // 200
}

func GetUsers(pageSize int, pageNum int) ([]User, int) {
	var users = []User{}

	offset := (pageNum - 1) * pageSize
	if pageNum == -1 && pageSize == -1 {
		offset = -1
	}
	if pageSize > 50 {
		pageSize = 50
	}

	var total_num int64
	result := db.Limit(pageSize).Offset(offset).Find(&users).Count(&total_num)
	if result.Error != nil {
		return nil, 0
	}
	log.Printf("Query success, rowAffected:%v \n", result.RowsAffected)
	return users, int(total_num)
}

func RemoveUser(id int) int {
	var user User
	result := db.Where("id=?", id).Delete(&user)
	if result.Error != nil || result.RowsAffected == 0 {
		log.Printf("cannot delete user. id:%v", id)
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// user update hook
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	code := CheckUserExist(u.Username)
	if code != 200 {
		return errors.New(errmsg.GetErrMsg(code))
	}
	return nil
}
func UpdateUser(id int, data *User) int {
	var user User
	var maps = make(map[string]interface{})
	maps["username"] = data.Username
	maps["role"] = data.Role
	result := db.Model(&user).Where("id=?", id).Updates(maps)
	if result.Error != nil || result.RowsAffected == 0 {
		log.Println("Update failed; ", result.Error)
		return errmsg.ERROR
	}
	return errmsg.SUCCESS
}

// login authentication

func AuthLogin(username string, password string) (code int) {
	var user User
	db.Where("username=?", username).First(&user)
	if user.ID == 0 {
		return errmsg.ERROR_USER_NOT_EXIST
	}
	if encryption.ScryptPassword(password) != user.Password {
		return errmsg.ERROR_PASSWORD_WRONG
	}
	return errmsg.SUCCESS
}
