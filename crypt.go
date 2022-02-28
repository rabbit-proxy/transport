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
	key          []byte
	encryptSteam cipher.Stream
	decryptSteam cipher.Stream
}

func NewAesEncryption(mode string, commonIV []byte, encryptKey string) *AesEncryption {
	return &AesEncryption{
		cryptMode:  mode,
		commonIV:   commonIV,
		encryptKey: encryptKey,
	}
}

func (cryptMode *AesEncryption) getAESCipher() cipher.Block {
	if cryptMode.key == nil {
		hash := sha256.New()
		hash.Write([]byte(cryptMode.encryptKey))
		hashedKey := hash.Sum(nil)

		switch cryptMode.cryptMode {
		case "aes128":
			cryptMode.key = hashedKey[:16]
		case "aes192":
			cryptMode.key = hashedKey[:24]
		case "aes256":
			cryptMode.key = hashedKey[:32]
		}
	}

	aesCipher, err := aes.NewCipher(cryptMode.key)
	if err != nil {
		zap.L().Panic("get aesCipher failed", zap.Error(err))
	}

	return aesCipher
}

func (cryptMode *AesEncryption) getEnc() cipher.Stream {
	aesCipher := cryptMode.getAESCipher()
	cfbEnc := cipher.NewCFBEncrypter(aesCipher, cryptMode.commonIV)
	return cfbEnc
}

func (cryptMode *AesEncryption) getDec() cipher.Stream {
	aesCipher := cryptMode.getAESCipher()
	cfbDec := cipher.NewCFBDecrypter(aesCipher, cryptMode.commonIV)
	return cfbDec
}

func (cryptMode *AesEncryption) Encrypt(plainText []byte) {

	if cryptMode.encryptSteam == nil { // todo 这个地方可能不是并发安全的
		cryptMode.encryptSteam = cryptMode.getEnc()
	}

	cipherText := GetBuffer()
	defer PutBuffer(cipherText)

	cryptMode.encryptSteam.XORKeyStream(cipherText, plainText)
	copy(plainText, cipherText)
}

func (cryptMode *AesEncryption) Decrypt(cipherText []byte) {

	if cryptMode.decryptSteam == nil {
		cryptMode.decryptSteam = cryptMode.getDec()
	}

	plaintext := GetBuffer()
	defer PutBuffer(plaintext)

	cryptMode.decryptSteam.XORKeyStream(plaintext, cipherText)
	copy(cipherText, plaintext)
}

// PlainEncryption plain mode, 不再对内容进行加密，安全性由下层协议进行完成
type PlainEncryption struct{}

func (cryptMode *PlainEncryption) Encrypt(plainText []byte)  {}
func (cryptMode *PlainEncryption) Decrypt(cipherText []byte) {}
