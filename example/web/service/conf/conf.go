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
		Address      string       `mapstructure:"addr" json:"addr"`
		AttachConfig AttachConfig `mapstructure:"attach" json:"attach"`
		CameraConfig CameraConfig `mapstructure:"camera" json:"camera"`
	}

	AttachConfig struct {
		IP      string `mapstructure:"ip" json:"ip"`
		Port    int    `mapstructure:"port" json:"port"`
		Enable  bool   `mapstructure:"enable" json:"enable"`
		Dir     string `mapstructure:"dir" json:"dir"`
		LogFile string `mapstructure:"logFile" json:"logFile"`
	}

	CameraConfig struct {
		Enable      bool        `mapstructure:"enable" json:"enable"`
		Dir         string      `mapstructure:"dir" json:"dir"`
		URLPrefix   string      `mapstructure:"urlPrefix" json:"urlPrefix"`
		MinioConfig MinioConfig `mapstructure:"minio" json:"minio"`
	}

	MinioConfig struct {
		Enable    bool   `mapstructure:"enable" json:"enable"`
		Endpoint  string `mapstructure:"endpoint" json:"endpoint"`
		AppKey    string `mapstructure:"appKey" json:"appKey"`
		AppSecret string `mapstructure:"appSecret" json:"appSecret"`
		Bucket    string `mapstructure:"bucket" json:"bucket"`
	}

	JTConfig struct {
		Address    string `mapstructure:"addr" json:"addr"`
		ID         string `mapstructure:"id" json:"id"`
		Verify     bool   `mapstructure:"verify" json:"verify"`
		HTTPPrefix string `mapstructure:"httpPrefix" json:"httpPrefix"`
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
