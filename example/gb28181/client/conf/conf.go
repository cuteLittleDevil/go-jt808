package conf

import (
	"github.com/spf13/viper"
)

type (
	GlobalConfig struct {
		GB28181 GB28181Config `mapstructure:"gb28181" json:"gb28181"`
		JT1078  JT1078Config  `mapstructure:"jt1078" json:"jt1078"`
	}

	GB28181Config struct {
		Transport string         `mapstructure:"transport" json:"transport"`
		Platform  PlatformConfig `mapstructure:"platform" json:"platform"`
		Device    DeviceConfig   `mapstructure:"device" json:"device"`
	}

	PlatformConfig struct {
		Domain   string `mapstructure:"domain" json:"domain"`
		ID       string `mapstructure:"id" json:"id"`
		Password string `mapstructure:"password" json:"password"`
		IP       string `mapstructure:"ip" json:"ip"`
		Port     int    `mapstructure:"port" json:"port"`
	}

	DeviceConfig struct {
		ID   string `mapstructure:"id" json:"id"`
		IP   string `mapstructure:"ip" json:"ip"`
		Port int    `mapstructure:"port" json:"port"`
	}

	JT1078Config struct {
		File string `mapstructure:"file" json:"file"`
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
