package core

import (
	"context"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoConfig struct {
	Enable      bool   `json:"enable" mapstructure:"enable"`
	Name        string `json:"name" mapstructure:"name"`
	Dsn         string `json:"dsn" mapstructure:"dsn"`
	Database    string `json:"database" mapstructure:"database"`
	MinPoolSize uint64 `json:"min_pool_size" mapstructure:"min_pool_size"`
	MaxPoolSize uint64 `json:"max_pool_size" mapstructure:"max_pool_size"`
}

func parseMongoConfig(v *viper.Viper) (conf []mongoConfig) {
	if v == nil {
		return nil
	}
	if err := v.UnmarshalKey("database", &conf); err != nil {
		panic("log 配置解析错误" + err.Error())
	}
	for key, item := range conf {
		if item.MaxPoolSize == 0 {
			conf[key].MaxPoolSize = 100
		}
	}
	return
}

func initMongo() {
	confList := parseMongoConfig(globalConfig)
	clients := make(map[string]*mongo.Client)
	for _, conf := range confList {
		if !conf.Enable {
			return
		}
		client, err := mongo.Connect(context.Background(),
			options.Client().ApplyURI(conf.Dsn).
				SetMinPoolSize(conf.MinPoolSize).
				SetMaxPoolSize(conf.MaxPoolSize),
		)
		if err != nil {
			panic(err)
		}
		clients[conf.Name] = client
	}
	globalMongoConnects = clients
}
