package conf

import (
	"github.com/spf13/viper"
)

type (
	GlobalConfig struct {
		ServerConfig   ServerConfig   `mapstructure:"server" json:"server"`
		NatsConfig     NatsConfig     `mapstructure:"nats" json:"nats"`
		AlarmConfig    AlarmConfig    `mapstructure:"alarm" json:"alarm"`
		TdengineConfig TdengineConfig `mapstructure:"tdengine" json:"tdengine"`
		MongodbConfig  MongodbConfig  `mapstructure:"mongodb" json:"mongodb"`
	}

	ServerConfig struct {
		Address string `mapstructure:"addr" json:"addr"`
		LogDir  string `mapstructure:"logDir" json:"logDir"`
	}

	NatsConfig struct {
		Open    bool   `mapstructure:"open" json:"open"`
		Address string `mapstructure:"addr" json:"addr"`
	}

	AlarmConfig struct {
		Enable    bool   `mapstructure:"enable" json:"enable"`
		OnFileApi string `mapstructure:"onFileApi" json:"onFileApi"`
	}

	TdengineConfig struct {
		Enable bool   `mapstructure:"enable" json:"enable"`
		Dsn    string `mapstructure:"dsn" json:"dsn"`
	}

	MongodbConfig struct {
		Enable bool   `mapstructure:"enable" json:"enable"`
		Dsn    string `mapstructure:"dsn" json:"dsn"`
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
