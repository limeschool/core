package core

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

type redisConfig struct {
	Enable   bool   `json:"enable" mapstructure:"enable"`     //是否启用redis
	Host     string `json:"host" mapstructure:"host"`         //redis的连接地址
	Password string `json:"password" mapstructure:"password"` //redis的密码
	DB       int    `json:"db" mapstructure:"db"`
	PoolSize int    `json:"pool_size" mapstructure:"pool_size"`
}

func parseRedisConfig(v *viper.Viper) (conf redisConfig) {
	if v == nil {
		return
	}
	if err := v.UnmarshalKey("redis", &conf); err != nil {
		panic("redis 配置解析错误" + err.Error())
	}
	return
}

func InitRedis() {
	conf := parseRedisConfig(globalConfig)
	if !conf.Enable {
		return
	}
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Host,
		Password: conf.Password,
		PoolSize: conf.PoolSize,
		DB:       conf.DB,
	})
	if err := client.Ping(context.TODO()).Err(); err != nil {
		panic("redis 连接失败" + err.Error())
	}
	globalRedisConnect = client
}
