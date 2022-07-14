package core

import (
	"github.com/spf13/viper"
	"time"
)

type systemConfig struct {
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

func initSystemConfig(v *viper.Viper) {
	conf := systemConfig{}
	if err := v.UnmarshalKey("system", &conf); err != nil {
		panic(err)
	}
	globalSystemConfig = conf
}
