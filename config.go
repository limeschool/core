package core

import (
	"core/config_drive"
	"encoding/json"
	"flag"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

type Config interface {
	Get(key string) interface{}
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetInt32(key string) int32
	GetInt64(key string) int64
	GetUint(key string) uint
	GetUint32(key string) uint32
	GetUint64(key string) uint64
	GetFloat64(key string) float64
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	UnmarshalKey(key string, val interface{}) error
	Unmarshal(val interface{}) error
}

type config struct {
	logger *zap.Logger
}

func newConfig(log *zap.Logger) *config {
	return &config{
		logger: log,
	}
}

func WatchConfig(f config_drive.CallFunc) {
	config_drive.CallBack = f
}

var configFile = flag.String("c", "config/dev.json", "the config file path")

func initConfig() {
	flag.Parse()
	conf := config_drive.Config{}
	if configFile == nil {
		conf = config_drive.Config{
			Drive:    os.Getenv("Drive"),
			Host:     os.Getenv("Host"),
			Type:     os.Getenv("Type"),
			Username: os.Getenv("Username"),
			Password: os.Getenv("Password"),
			Path:     os.Getenv("Path"),
		}
	} else {
		temp := strings.Split(*configFile, ".")
		conf = config_drive.Config{
			Drive: "local",
			Type:  temp[len(temp)-1],
			Path:  *configFile,
		}
	}
	globalConfig = config_drive.Init(&conf)
	initLogKey(globalConfig)
	initSystemConfig(globalConfig)
}

func (c *config) Set(key string, value interface{}) {
	globalConfig.Set(key, value)
	c.logger.Info(SetConfigTip, zap.Any(key, value))
}

func (c *config) Get(key string) interface{} {
	res := globalConfig.Get(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetString(key string) string {
	res := globalConfig.GetString(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetBool(key string) bool {
	res := globalConfig.GetBool(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetInt(key string) int {
	res := globalConfig.GetInt(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetInt32(key string) int32 {
	res := globalConfig.GetInt32(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetInt64(key string) int64 {
	res := globalConfig.GetInt64(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetUint(key string) uint {
	res := globalConfig.GetUint(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetUint32(key string) uint32 {
	res := globalConfig.GetUint32(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetUint64(key string) uint64 {
	res := globalConfig.GetUint64(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetFloat64(key string) float64 {
	res := globalConfig.GetFloat64(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetTime(key string) time.Time {
	res := globalConfig.GetTime(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetDuration(key string) time.Duration {
	res := globalConfig.GetDuration(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetIntSlice(key string) []int {
	res := globalConfig.GetIntSlice(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetStringSlice(key string) []string {
	res := globalConfig.GetStringSlice(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetStringMap(key string) map[string]interface{} {
	res := globalConfig.GetStringMap(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetStringMapString(key string) map[string]string {
	res := globalConfig.GetStringMapString(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) GetStringMapStringSlice(key string) map[string][]string {
	res := globalConfig.GetStringMapStringSlice(key)
	c.logger.Info(GetConfigTip, zap.Any(key, res))
	return res
}

func (c *config) UnmarshalKey(key string, val interface{}) error {
	defer c.logger.Info(GetConfigTip, zap.Any(key, val))

	h := H{}
	if err := globalConfig.UnmarshalKey(key, val); err != nil {
		return err
	}
	byteData, _ := json.Marshal(h)
	return json.Unmarshal(byteData, val)
}

func (c *config) Unmarshal(val interface{}) error {
	h := H{}
	defer c.logger.Info(GetConfigTip, zap.Any("res", val))
	if err := globalConfig.Unmarshal(&h); err != nil {
		return err
	}
	byteData, _ := json.Marshal(h)
	return json.Unmarshal(byteData, val)
}
