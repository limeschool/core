package core

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"testing"
)

func TestLog(t *testing.T) {
	v := viper.New()
	v.SetConfigType("json")
	v.Set("log", H{
		"OutputFile":    true,
		"OutputConsole": true,
		"Filename":      "./logs/log.log",
	})
	log := initLog(v, "test-service")
	log.Error("info msg", zap.Any("test", "ss"))
}
