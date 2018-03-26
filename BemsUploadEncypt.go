package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

func bemsPKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func aesEncrypt(data, key []byte) ([]byte, error) {
	myCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipherBlockSize := myCipher.BlockSize()
	data = bemsPKCS5Padding(data, cipherBlockSize)
	encrypter := cipher.NewCBCEncrypter(myCipher, key)
	crypted := make([]byte, len(data))
	encrypter.CryptBlocks(crypted, data)
	return crypted, nil
}

func bemsUploadEncrypt(b []byte) ([]byte, error) {
	b, err := aesEncrypt(b, []byte("useruseruseruser"))
	if err != nil {
		return []byte(""), err
	}
	return b, nil
}

func aesDecrypt(data, key []byte) ([]byte, error) {
	myCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(myCipher, key)
	decrypted := make([]byte, len(data))
	decrypter.CryptBlocks(decrypted, data)
	return decrypted, nil
}

func bemsUploadDecrypt(b []byte) ([]byte, error) {
	b, err := aesDecrypt(b, []byte("useruseruseruser"))
	if err != nil {
		return []byte(""), err
	}
	return b, nil
}
