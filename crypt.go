package transport

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"go.uber.org/zap"
)

type Encryption interface {
	Encrypt(plainText []byte)
	Decrypt(cipherText []byte)
}

// AesEncryption 默认的 aes 加密方式
type AesEncryption struct {
	cryptMode    string
	commonIV     []byte
	encryptKey   string
	encryptSteam cipher.Stream
	decryptSteam cipher.Stream
}

func NewAesEncryption(mode string, commonIV []byte, encryptKey string) *AesEncryption {
	return &AesEncryption{
		cryptMode:  mode,
		commonIV:   commonIV,
		encryptKey: encryptKey,
		encryptSteam: getEnc(getAESCipher(encryptKey, mode), commonIV),
		decryptSteam: getDec(getAESCipher(encryptKey, mode), commonIV),
	}
}

func getAESCipher(encryptKey string, encryptMode string) cipher.Block {
	hash := sha256.New()
	hash.Write([]byte(encryptKey))
	hashedKey := hash.Sum(nil)

	var key []byte
	switch encryptMode {
	case "aes128":
		key = hashedKey[:16]
	case "aes192":
		key = hashedKey[:24]
	case "aes256":
		key = hashedKey[:32]
	}

	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		zap.L().Panic("get aesCipher failed", zap.Error(err))
	}
	return aesCipher
}

func  getEnc(aesCipher cipher.Block, commonIV []byte) cipher.Stream {
	cfbEnc := cipher.NewCFBEncrypter(aesCipher, commonIV)
	return cfbEnc
}

func  getDec(aesCipher cipher.Block, commonIV []byte) cipher.Stream {
	cfbEnc := cipher.NewCFBDecrypter(aesCipher, commonIV)
	return cfbEnc
}

func (cryptMode *AesEncryption) Encrypt(plainText []byte) {
	cipherText := GetBuffer()
	defer PutBuffer(cipherText)

	cryptMode.encryptSteam.XORKeyStream(cipherText, plainText)
	copy(plainText, cipherText)
}

func (cryptMode *AesEncryption) Decrypt(cipherText []byte) {
	plaintext := GetBuffer()
	defer PutBuffer(plaintext)

	cryptMode.decryptSteam.XORKeyStream(plaintext, cipherText)
	copy(cipherText, plaintext)
}

// PlainEncryption plain mode, 不再对内容进行加密，安全性由下层协议进行完成
type PlainEncryption struct{}

func (cryptMode *PlainEncryption) Encrypt(plainText []byte)  {}
func (cryptMode *PlainEncryption) Decrypt(cipherText []byte) {}
