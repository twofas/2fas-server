package config

import (
	"strings"

	"github.com/spf13/viper"
	"github.com/twofas/2fas-server/internal/common/logging"
)

var Config Configuration

type Configuration struct {
	Debug     bool            `json:"debug"`
	Env       string          `json:"env"`
	Aws       AwsConfig       `json:"aws"`
	Db        DbConf          `json:"db"`
	Redis     RedisConf       `json:"redis"`
	App       AppConfig       `json:"app"`
	Websocket WebsocketConfig `json:"websocket"`
	Security  SecurityConfig  `json:"security"`
	Icons     IconsConfig     `json:"icons"`
}

func (c *Configuration) IsTestingEnv() bool {
	return strings.ToLower(c.Env) == "testing"
}

type DbConf struct {
	Host     string `mapstructure:"mysql_host" json:"host"`
	Port     int    `mapstructure:"mysql_port" json:"port"`
	Username string `mapstructure:"mysql_username" json:"username"`
	Password string `mapstructure:"mysql_password" json:"password"`
	Database string `mapstructure:"mysql_database" json:"database"`
}

type RedisConf struct {
	ServiceUrl           string `mapstructure:"service_url" json:"service_url"`
	Port                 int    `mapstructure:"port" json:"port"`
	PersistentConnection bool   `mapstructure:"persistent_connection" json:"persistent_connection"`
}

type AppConfig struct {
	ListenAddr string `mapstructure:"listen_addr" json:"listen_addr"`
}

type SecurityConfig struct {
	RateLimitIP     int `mapstructure:"rate_limit_ip" json:"rate_limit_ip"`
	RateLimitMobile int `mapstructure:"rate_limit_mobile" json:"rate_limit_mobile"`
	RateLimitBE     int `mapstructure:"rate_limit_be" json:"rate_limit_be"`
}

type WebsocketConfig struct {
	ListenAddr string `mapstructure:"listen_addr" json:"listen_addr"`
	ApiUrl     string `mapstructure:"url" json:"api_url"`
}

type AwsConfig struct {
	Region            string `mapstructure:"region" json:"region"`
	Profile           string `mapstructure:"profile" json:"profile"`
	S3AccessKeyId     string `mapstructure:"s3_access_key_id" json:"s3_access_key_id"`
	S3AccessSecretKey string `mapstructure:"s3_access_secret_key" json:"s3_access_secret_key"`
}

type IconsConfig struct {
	AwsProfile        string `mapstructure:"aws_profile" json:"aws_profile"`
	S3Bucket          string `mapstructure:"s3_bucket" json:"s3_bucket"`
	S3AccessKeyId     string `mapstructure:"s3_access_key_id" json:"s3_access_key_id"`
	S3AccessSecretKey string `mapstructure:"s3_access_secret_key" json:"s3_access_secret_key"`
}

func initViper(configFilePath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.BindEnv("aws.profile", "AWS_PROFILE")

	viper.BindEnv("db.mysql_host", "MYSQL_HOST")
	viper.BindEnv("db.mysql_port", "MYSQL_PORT")
	viper.BindEnv("db.mysql_database", "MYSQL_DATABASE")
	viper.BindEnv("db.mysql_username", "MYSQL_USERNAME")
	viper.BindEnv("db.mysql_password", "MYSQL_PASSWORD")

	viper.BindEnv("redis.service_url", "REDIS_SERVICE_URL")

	viper.BindEnv("app.listen_addr", "API_LISTEN_ADDR")

	viper.BindEnv("websocket.listen_addr", "WEBSOCKET_LISTEN_ADDR")
	viper.BindEnv("websocket.url", "WEBSOCKET_API_URL")

	viper.BindEnv("aws.s3_access_key_id", "S3_USER_ACCESS_KEY_ID")
	viper.BindEnv("aws.s3_access_secret_key", "S3_USER_ACCESS_SECRET_KEY")

	viper.BindEnv("icons.s3_access_key_id", "ICONS_S3_ACCESS_KEY_ID")
	viper.BindEnv("icons.s3_access_secret_key", "ICONS_S3_ACCESS_SECRET_KEY")

	viper.BindEnv("security.rate_limit_ip", "SECURITY_RATE_LIMIT_IP")
	viper.BindEnv("security.rate_limit_be", "SECURITY_RATE_LIMIT_BE")
	viper.BindEnv("security.rate_limit_mobile", "SECURITY_RATE_LIMIT_MOBILE")

	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	}

	err := viper.ReadInConfig()
	if err != nil {
		logging.Fatal("failed to read the configuration file: %s", err)
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		logging.Fatal("Can not unmarshal configuration", err)
	}
}

func LoadConfiguration() {
	initViper("")
}
