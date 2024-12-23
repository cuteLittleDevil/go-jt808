package conf

import (
	"github.com/spf13/viper"
)

type (
	GlobalConfig struct {
		ServerConfig ServerConfig `mapstructure:"server" json:"server"`
		NatsConfig   NatsConfig   `mapstructure:"nats" json:"nats"`
	}

	ServerConfig struct {
		Address   string `mapstructure:"addr" json:"addr"`
		LogDir    string `mapstructure:"logDir" json:"logDir"`
		OnFileApi string `mapstructure:"onFileApi" json:"onFileApi"`
	}

	NatsConfig struct {
		Open    bool   `mapstructure:"open" json:"open"`
		Address string `mapstructure:"addr" json:"addr"`
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
	SetData(&globalConfig)
	globalConfig.Show()
	return nil
}
