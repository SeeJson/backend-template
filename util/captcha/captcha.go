package captcha

import (
	"github.com/mojocn/base64Captcha"
)

type Config struct {
	Height     int     `mapstructure:"height"`
	Width      int     `mapstructure:"width"`
	MaxSkew    float64 `mapstructure:"max_skew"`
	DotCount   int     `mapstructure:"dot_count"`
	CaptchaLen int     `mapstructure:"captcha_len"`
}

var cfg Config

var store = base64Captcha.DefaultMemStore

func SetConfig(c Config) {
	cfg = c
}

func Generate() (id string, b64s string, err error) {
	driver := base64Captcha.NewDriverDigit(
		cfg.Height,
		cfg.Width,
		cfg.CaptchaLen,
		cfg.MaxSkew,
		cfg.DotCount,
	)
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, err = c.Generate()
	return
}

func Verify(id string, answer string) bool {
	return store.Verify(id, answer, true)
}
