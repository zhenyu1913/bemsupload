package main

import (
    "bytes"
    "crypto/cipher"
    "crypto/aes"
)

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func AesEncrypt(data, key []byte) ([]byte, error) {
     myCipher, err := aes.NewCipher(key)
     if err != nil {
          return nil, err
     }
     cipherBlockSize := myCipher.BlockSize()
     data = PKCS5Padding(data, cipherBlockSize)
     encrypter := cipher.NewCBCEncrypter(myCipher, key)
     crypted := []byte{}
     encrypter.CryptBlocks(crypted, data)
     return crypted, nil
}

func BemsUploadEncrypt(b []byte) ([]byte, error) {
    b, err := AesEncrypt(b , []byte("aes.modeaes.mode"))
    if err != nil {
        return []byte(""), err
    }
    return b, nil
}

func AesDecrypt(data, key []byte) ([]byte, error) {
     myCipher, err := aes.NewCipher(key)
     if err != nil {
          return nil, err
     }
     decrypter := cipher.NewCBCDecrypter(myCipher, key)
     decrypted := []byte{}
     decrypter.CryptBlocks(decrypted, data)
     return decrypted, nil
}

func BemsUploadDecrypt(b []byte) ([]byte, error) {
    b, err := AesDecrypt(b , []byte("aes.modeaes.mode"))
    if err != nil {
        return []byte(""), err
    }
    return b, nil
}
