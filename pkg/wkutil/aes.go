package wkutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

func AesEncryptPkcs7(origData []byte, key []byte, iv []byte) ([]byte, error) {
	return AesEncrypt(origData, key, iv, PKCS7Padding)
}

func AesEncryptPkcs7Base64(origData []byte, key []byte, iv []byte) ([]byte, error) {
	data, err := AesEncrypt(origData, key, iv, PKCS7Padding)
	if err != nil {
		return data, err
	}
	dataStr := base64.StdEncoding.EncodeToString(data)
	return []byte(dataStr), nil
}

func AesEncrypt(origData []byte, key []byte, iv []byte, paddingFunc func([]byte, int) []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = paddingFunc(origData, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func AesDecryptPkcs5(crypted []byte, key []byte, iv []byte) ([]byte, error) {
	return AesDecrypt(crypted, key, iv, PKCS5UnPadding)
}

func AesDecryptPkcs7(crypted []byte, key []byte, iv []byte) ([]byte, error) {
	return AesDecrypt(crypted, key, iv, PKCS7UnPadding)
}

func AesDecrypt(crypted, key []byte, iv []byte, unPaddingFunc func([]byte) []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = unPaddingFunc(origData)
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if length < unpadding {
		return []byte("unpadding error")
	}
	return origData[:(length - unpadding)]
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	totalLen := len(ciphertext) + padding

	if cap(ciphertext) < totalLen {
		newCap := cap(ciphertext) * 2
		if newCap < totalLen {
			newCap = totalLen
		}
		newSlice := make([]byte, len(ciphertext), newCap)
		copy(newSlice, ciphertext)
		ciphertext = newSlice
	}

	ciphertext = ciphertext[:totalLen]

	for i := len(ciphertext) - padding; i < len(ciphertext); i++ {
		ciphertext[i] = byte(padding)
	}

	return ciphertext
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)

	unpadding := int(origData[length-1])

	return origData[:(length - unpadding)]

}
