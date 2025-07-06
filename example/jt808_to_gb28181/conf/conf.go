package conf

import (
	"github.com/cuteLittleDevil/go-jt808/shared/consts"
	"github.com/spf13/viper"
)

type (
	GlobalConfig struct {
		Adapter   AdapterConfig   `mapstructure:"adapter" json:"adapter"`
		JT808     JT808Config     `mapstructure:"jt808" json:"jt808"`
		Simulator SimulatorConfig `mapstructure:"simulator" json:"simulator"`
	}

	AdapterConfig struct {
		Enable      bool           `mapstructure:"enable" json:"enable"`
		Address     string         `mapstructure:"address" json:"address"`
		RetrySecond int            `mapstructure:"retrySecond" json:"retrySecond"`
		Leader      string         `mapstructure:"leader" json:"leader"`
		Followers   []FollowerInfo `mapstructure:"followers" json:"followers"`
	}

	FollowerInfo struct {
		Address       string                    `mapstructure:"address" json:"address"`
		AllowCommands []consts.JT808CommandType `mapstructure:"allowCommands" json:"allowCommands"`
	}

	JT808Config struct {
		ApiAddress string        `mapstructure:"apiAddress" json:"apiAddress"`
		Address    string        `mapstructure:"address" json:"address"`
		HasDetails bool          `mapstructure:"hasDetails" json:"hasDetails"`
		JT1078     JT1078Config  `mapstructure:"jt1078" json:"jt1078"`
		GB28181    GB28181Config `mapstructure:"gb28181" json:"gb28181"`
	}

	JT1078Config struct {
		IP        string `mapstructure:"ip" json:"ip"`
		OnPlayURL string `mapstructure:"onPlayURL" json:"onPlayURL"`
	}

	GB28181Config struct {
		Transport string `mapstructure:"transport" json:"transport"`
		Platform  struct {
			Domain   string `mapstructure:"domain" json:"domain"`
			ID       string `mapstructure:"id" json:"id"`
			Password string `mapstructure:"password" json:"password"`
			IP       string `mapstructure:"ip" json:"ip"`
			Port     int    `mapstructure:"port" json:"port"`
		} `mapstructure:"platform" json:"platform"`
		Device struct {
			Type      string `mapstructure:"type" json:"type"`
			Path      string `mapstructure:"path" json:"path"`
			SheetName string `mapstructure:"sheetName" json:"sheetName"`
		} `mapstructure:"device" json:"device"`
	}

	SimulatorConfig struct {
		Enable   bool   `mapstructure:"enable" json:"enable"`
		Address  string `mapstructure:"address" json:"address"`
		Sim      string `mapstructure:"sim" json:"sim"`
		FilePath string `mapstructure:"filePath" json:"filePath"`
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
