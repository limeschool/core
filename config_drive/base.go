package config_drive

import "github.com/spf13/viper"

type Config struct {
	Drive    string //中间件 etcd/consul/zk
	Host     string //连接地址
	Type     string //配置数据格式 json\yaml...
	Username string //连接用户名
	Password string //连接密码
	Path     string //配置存储目录
}

type ConfigService interface {
	Init() *viper.Viper
	Get(v *viper.Viper) error
	Watch(v *viper.Viper)
}

type CallFunc func(v *viper.Viper)

var CallBack func(v *viper.Viper)

func Init(conf *Config) *viper.Viper {
	switch conf.Drive {
	case "etcd":
		return NewEtcd(conf).Init()
	case "zk":
		return NewZK(conf).Init()
	case "consul":
		return NewConsul(conf).Init()
	case "local":
		return NewLocal(conf).Init()
	default:
		panic("config type is fail")
	}
	return nil
}
