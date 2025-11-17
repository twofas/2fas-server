package domain

import (
	"time"

	"github.com/spf13/viper"
)

type DebugLogsConfig struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsRegion          string        `mapstructure:"aws_region" json:"aws_region"`
	AwsProfile         string        `mapstructure:"aws_profile" json:"aws_profile"`
	DebugLogsDirectory string        `mapstructure:"directory" json:"debug_logs_directory"`
	ExpireAt           time.Duration `mapstructure:"expire_at" json:"expire_at"`
}

func LoadDebugLogsConfig() DebugLogsConfig {
	viper.BindEnv("mobile.debug.aws_access_key_id", "MOBILE_DEBUG_AWS_ACCESS_KEY_ID")
	viper.BindEnv("mobile.debug.aws_secret_access_key", "MOBILE_DEBUG_AWS_SECRET_ACCESS_KEY")

	return DebugLogsConfig{
		AwsRegion:          viper.GetString("mobile.debug.aws_region"),
		AwsProfile:         viper.GetString("mobile.debug.aws_profile"),
		AwsAccessKeyId:     viper.GetString("mobile.debug.aws_access_key_id"),
		AwsSecretAccessKey: viper.GetString("mobile.debug.aws_secret_access_key"),
		DebugLogsDirectory: viper.GetString("mobile.debug.directory"),
		ExpireAt:           time.Hour * 24 * 7,
	}
}
