package upload

import (
	"BlogProject/Shares/errmsg"
	"context"
	"mime/multipart"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

const (
	AccessKey string = "69xqoOmXpS3fL9yb5o3olferH4rW2ZUN1PHveIfy"
	SecretKey string = "i0gdD2-S3WBC26QHxPnSxJeueDb6oj5zjA0H4sne"
	Bucket    string = "for-blog-imgs"
	ImgsUrl   string = "cdn.leelennin.top/"
)

func UploadFile(file multipart.File, fileSize int64) (string, int) {
	putPolicy := storage.PutPolicy{
		Scope: Bucket,
	}
	mac := qbox.NewMac(AccessKey, SecretKey)
	upToken := putPolicy.UploadToken(mac)

	cfg := storage.Config{
		Zone:     &storage.ZoneXinjiapo,
		UseHTTPS: false,
	}

	putExtra := storage.PutExtra{}

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	err := formUploader.PutWithoutKey(context.Background(), &ret, upToken, file, fileSize, &putExtra)
	if err != nil {
		return "", errmsg.ERROR
	}
	url := ImgsUrl + ret.Key
	return url, errmsg.SUCCESS
}
