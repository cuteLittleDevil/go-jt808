package conf

import (
	"github.com/spf13/viper"
)

type (
	GlobalConfig struct {
		GB28181 GB28181Config `mapstructure:"gb28181" json:"gb28181"`
		ZLM     ZLMConfig     `mapstructure:"zlm" json:"zlm"`
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

	ZLMConfig struct {
		Secret         string `mapstructure:"secret" json:"secret"`
		Vhost          string `mapstructure:"vhost" json:"vhost"`
		App            string `mapstructure:"app" json:"app"`
		Stream         string `mapstructure:"stream" json:"stream"`
		Rtsp           string `mapstructure:"rtsp" json:"rtsp"`
		AddStreamProxy string `mapstructure:"addStreamProxy" json:"addStreamProxy"`
		StartSendRtp   string `mapstructure:"startSendRtp" json:"startSendRtp"`
		StopSendRtp    string `mapstructure:"stopSendRtp" json:"stopSendRtp"`
		CloseStreams   string `mapstructure:"closeStreams" json:"closeStreams"`
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
