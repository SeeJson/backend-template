package config

import (
	"strings"

	handler "github.com/SeeJson/account/cmd/account/handler/http"
	httpserver "github.com/SeeJson/account/cmd/account/server/http"
	rpcserver "github.com/SeeJson/account/cmd/account/server/rpc"
	"github.com/SeeJson/account/service"
	"github.com/SeeJson/account/util/captcha"
	"github.com/SeeJson/account/util/jwt"
	mlog "github.com/SeeJson/account/util/log"
	mongodao "github.com/SeeJson/account/util/mongo"
	redisdao "github.com/SeeJson/account/util/redis"
	"github.com/spf13/viper"
)

const (
	ConfigEnvPrefix = "SR"
)

type Config struct {
	LogConfig     mlog.Config        `mapstructure:"log_config"`
	HttpConfig    httpserver.Config  `mapstructure:"http_config"`
	RpcConfig     rpcserver.Config   `mapstructure:"rpc_config"`
	MongoConfig   mongodao.Config    `mapstructure:"mongo_config"`
	JwtConfig     jwt.Config         `mapstructure:"jwt_config"`
	CaptchaConfig captcha.Config     `mapstructure:"captcha_config"`
	UserConfig    service.UserConfig `mapstructure:"user_config"`
	RedisConfig   redisdao.Config    `mapstructure:"redis_config"`
	HandlerConfig handler.Config     `mapstructure:"handler_config"`
}

/*
 * @param cfgPath: 配置文件所在路径，如：./conf
 * @param cfgFilename: 配置文件的文件名（不包含后缀），如：config
 * @param cfgType: 配置文件的文件类型，如：yaml
 */
func Load(cfgPath string, cfgFilename string, cfgType string) error {
	viper.AddConfigPath(cfgPath)
	viper.SetConfigName(cfgFilename)
	viper.SetConfigType(cfgType)
	viper.SetEnvPrefix(ConfigEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}

	// apply
	applyConfig(cfg)

	return nil
}

func applyConfig(cfg *Config) {
	// Set Config
	mlog.SetConfig(cfg.LogConfig)
	httpserver.SetConfig(cfg.HttpConfig)
	rpcserver.SetConfig(cfg.RpcConfig)
	mongodao.SetConfig(cfg.MongoConfig)
	jwt.SetConfig(cfg.JwtConfig)
	captcha.SetConfig(cfg.CaptchaConfig)
	service.SetUserConfig(cfg.UserConfig)
	redisdao.SetConfig(cfg.RedisConfig)
	handler.SetConfig(cfg.HandlerConfig)
}
