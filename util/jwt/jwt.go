package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/SeeJson/account/util/crypt"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	PublicKeyPath  string `mapstructure:"public_key_path"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
	MaxAge         int    `mapstructure:"max_age"` // 会话有效期，单位：秒
	KeyFactory     string `mapstructure:"key_factory"`

	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

var cfg Config

func SetConfig(c Config) {
	cfg = c
	err := load()
	if err != nil {
		log.Fatalf("fail to load jwt key: %v", err) // fatal
	}
}

func load() error {
	data, err := ioutil.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		log.Errorf("fail to read private key file: %v", err)
		return err
	}

	key, err := crypt.Decrypt(cfg.KeyFactory, string(data))
	if err != nil {
		log.Errorf("fail to decrypt private key:%v", err)
		return err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		log.Errorf("invalid private key: %v", err)
		return err
	}

	cfg.PrivateKey = privateKey

	key, err = ioutil.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		log.Errorf("fail to read public key file: %v", err)
		return err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(key)
	if err != nil {
		log.Errorf("invalid private key: %v", err)
		return err
	}

	cfg.PublicKey = publicKey

	return nil
}

type Claims struct {
	Iat     int64
	Exp     int64
	Payload string
}

func GenBase64Token(payload string) (string, error) {
	claims := jwt.MapClaims{
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Duration(cfg.MaxAge) * time.Second).Unix(),
		"payload": payload,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token, err := jwtToken.SignedString(cfg.PrivateKey)
	if err != nil {
		log.Errorf("fail to sign token: %v", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString([]byte(token)), nil
}

func DecodeB64Token(b64Token string) (*Claims, error) {
	b, err := base64.StdEncoding.DecodeString(b64Token)
	if err != nil {
		log.Errorf("invalid base64 token: %v", b64Token)
		return nil, err
	}

	token, err := jwt.Parse(string(b), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return cfg.PublicKey, nil
	})
	if err != nil {
		log.Errorf("fail to parse token: %v", err)
		return nil, err
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Errorf("invalid token: %v", b64Token)
		return nil, err
	}
	claims := Claims{
		Iat:     int64(mapClaims["iat"].(float64)),
		Exp:     int64(mapClaims["exp"].(float64)),
		Payload: mapClaims["payload"].(string),
	}
	return &claims, nil
}
