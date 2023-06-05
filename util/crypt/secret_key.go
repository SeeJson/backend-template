package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

var (
	//初始向量
	initVector = []byte{0x25, 0x68, 0x82, 0x9B, 0xF9, 0x06, 0x29, 0x04, 0x16, 0x43, 0x14, 0x20, 0xD2, 0x5A, 0x25, 0x4C}
	//迭代次数
	PBKDF2_ITERATIONS = 50000
)

func init() {
	//初始化初始向量
	initVector[0] = initVector[15] & initVector[13]
	initVector[1] = initVector[6] ^ initVector[4]
	initVector[2] = initVector[0] & initVector[11]
	initVector[3] = initVector[15] | initVector[6]
	initVector[4] = initVector[7] & initVector[10]
	initVector[5] = initVector[12]
	initVector[6] = initVector[10] ^ initVector[7]
	initVector[7] = initVector[7] >> initVector[8]
	initVector[8] = initVector[12] ^ initVector[1]
	initVector[9] = initVector[2] & initVector[12]
	initVector[10] = initVector[8] & initVector[1]
	initVector[11] = initVector[2] + initVector[0]
	initVector[12] = initVector[12] - initVector[6]
	initVector[13] = initVector[15] | initVector[0]
	initVector[14] = initVector[9] & initVector[5]
	initVector[15] = initVector[9] | initVector[10]
}

func Encrypt(factor, data string) (string, error) {
	bytes := pbkdf2.Key([]byte(factor), initVector, PBKDF2_ITERATIONS, sha256.Size, sha256.New)
	encrypt, err := GCMEncrypt(bytes, []byte(data))
	if err != nil {
		return "", err
	}
	encryptStr := base64.StdEncoding.EncodeToString(encrypt)

	return encryptStr, nil

}

func Decrypt(factor, data string) ([]byte, error) {
	key := pbkdf2.Key([]byte(factor), initVector, PBKDF2_ITERATIONS, sha256.Size, sha256.New)
	bytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	decrypt, err := GCMDecrypt(key, bytes)
	if err != nil {
		return nil, err
	}
	return decrypt, nil
}

func GCMEncrypt(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesgcm.NonceSize())

	out := aesgcm.Seal(nonce, nonce, data, nil)
	return out, nil
}

func GCMDecrypt(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, ciphertext := data[:aesgcm.NonceSize()], data[aesgcm.NonceSize():]
	out, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return out, nil
}
