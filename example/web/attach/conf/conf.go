package conf

import (
	"github.com/spf13/viper"
)

type (
	GlobalConfig struct {
		Addr         string       `mapstructure:"addr" json:"addr"`
		AttachConfig AttachConfig `mapstructure:"attach" json:"attach"`
	}

	AttachConfig struct {
		Enable      bool        `mapstructure:"enable" json:"enable"`
		Dir         string      `mapstructure:"dir" json:"dir"`
		LogFile     string      `mapstructure:"logFile" json:"logFile"`
		MinioConfig MinioConfig `mapstructure:"minio" json:"minio"`
	}

	MinioConfig struct {
		Enable    bool   `mapstructure:"enable" json:"enable"`
		Endpoint  string `mapstructure:"endpoint" json:"endpoint"`
		AppKey    string `mapstructure:"appKey" json:"appKey"`
		AppSecret string `mapstructure:"appSecret" json:"appSecret"`
		Bucket    string `mapstructure:"bucket" json:"bucket"`
	}
)

func InitConfig(path string) error {
	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	var globalConfig GlobalConfig
	if err := v.Unmarshal(&globalConfig); err != nil {
		return err
	}
	setData(&globalConfig)
	globalConfig.Show()
	return nil
}
