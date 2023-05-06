package encryption

import (
	"encoding/base64"
	"log"
	"time"

	"golang.org/x/crypto/scrypt"
)

// encryption password

func ScryptPassword(password string) string {
	const KeyLen = 20
	salt := make([]byte, 8)
	copy(salt, []byte{0x12, 0x44, 0x04, 0x23, 0xa1, 0x78, 0xaf, 0x5b})

	dk, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, KeyLen)
	if err != nil {
		log.Println(err)
	}
	time_start := time.Now()
	encryptionPwd := base64.StdEncoding.EncodeToString(dk)
	log.Printf("length: %v; encryption time: %v\n", len(encryptionPwd), time.Since(time_start))
	return encryptionPwd
}
