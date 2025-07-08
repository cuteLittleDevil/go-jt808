package conf

import (
	"encoding/json"
	"fmt"
)

var (
	_global *GlobalConfig
)

func GetData() *GlobalConfig {
	return _global
}

func setData(globalConfig *GlobalConfig) {
	_global = globalConfig
}

func (g *GlobalConfig) Show() {
	b, _ := json.MarshalIndent(g, " ", "\t")
	fmt.Println(string(b))
}
