package conf

import (
	"github.com/spf13/viper"
)

type (
	GlobalConfig struct {
		ServerConfig ServerConfig `mapstructure:"server" json:"server"`
		NatsConfig   NatsConfig   `mapstructure:"nats" json:"nats"`
		FileConfig   FileConfig   `mapstructure:"file" json:"file"`
		JTConfig     JTConfig     `mapstructure:"jt808" json:"jt808"`
	}

	ServerConfig struct {
		Address string `mapstructure:"addr" json:"addr"`
		LogDir  string `mapstructure:"logDir" json:"logDir"`
	}

	NatsConfig struct {
		Open    bool   `mapstructure:"open" json:"open"`
		Address string `mapstructure:"addr" json:"addr"`
	}

	FileConfig struct {
		Address    string `mapstructure:"addr" json:"addr"`
		Dir        string `mapstructure:"dir" json:"dir"`
		LogFile    string `mapstructure:"logFile" json:"logFile"`
		AttachIP   string `mapstructure:"attachIP" json:"attachIP"`
		AttachPort int    `mapstructure:"attachPort" json:"attachPort"`
	}

	JTConfig struct {
		Address         string `mapstructure:"addr" json:"addr"`
		ID              string `mapstructure:"id" json:"id"`
		CameraDir       string `mapstructure:"cameraDir" json:"cameraDir"`
		CameraURLPrefix string `mapstructure:"cameraURLPrefix" json:"cameraURLPrefix"`
		HttpPrefix      string `mapstructure:"httpPrefix" json:"httpPrefix"`
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
