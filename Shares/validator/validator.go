package validator

import (
	"BlogProject/Shares/errmsg"
	"log"
	"reflect"

	"github.com/go-playground/locales/zh_Hans_CN"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/zh"
)

func Validate(data interface{}) (string, int) {
	validate := validator.New()
	uniTrans := ut.New(zh_Hans_CN.New())
	trans, _ := uniTrans.GetTranslator("zh_Hans_CN")

	err := zh.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		log.Println(err)
	}

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		label := field.Tag.Get("label")
		return label
	})

	err = validate.Struct(data)
	if err != nil {
		for _, v := range err.(validator.ValidationErrors) {
			return v.Translate(trans), errmsg.ERROR
		}
	}

	return "", errmsg.SUCCESS
}
