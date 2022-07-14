package core

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	globalConfig        *viper.Viper
	globalLog           *zap.Logger
	globalServiceName   string //服务名
	globalRequestConfig httpToolConfig
	globalSystemConfig  systemConfig
	globalRedisConnect  *redis.Client
	globalMysqlConnects map[string]*gorm.DB
	globalMongoConnects map[string]*mongo.Client
)
