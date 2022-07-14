package core

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type mysqlConfig struct {
	Enable                 bool   `json:"enable" mapstructure:"enable"`                                     //是否启用数据库
	Name                   string `json:"name" mapstructure:"name"`                                         //数据库名字【代称】
	Dsn                    string `json:"dsn" mapstructure:"dsn"`                                           //数据库dsn
	ConnMaxLifetime        int    `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`               //连接最大存活时长
	MaxOpenConn            int    `json:"max_open_conn" mapstructure:"max_open_conn"`                       //最大连接数
	MaxIdleConn            int    `json:"max_idle_conn" mapstructure:"max_idle_conn"`                       //最大空闲连接数
	Level                  int    `json:"level" mapstructure:"level"`                                       //日志打印等级
	SlowThreshold          int    `json:"slow_threshold" mapstructure:"slow_threshold"`                     //慢查询阈值
	TablePrefix            string `json:"table_prefix" mapstructure:"table_prefix"`                         //表前缀
	SkipDefaultTransaction *bool  `json:"skip_default_transaction" mapstructure:"skip_default_transaction"` //是否跳过默认事物
	SingularTable          *bool  `json:"singular_table" mapstructure:"singular_table"`                     //是否启用单数命名
	DryRun                 *bool  `json:"dry_run" mapstructure:"dry_run"`                                   //是否生成不执行
	PrepareStmt            *bool  `json:"prepare_stmt" mapstructure:"prepare_stmt"`                         //是否缓存
	DisableForeignKey      *bool  `json:"disable_foreign_key" mapstructure:"disable_foreign_key"`           //是否禁用外间约束
}

var (
	SkipDefaultTransaction = true
	SingularTable          = true
	DryRun                 = false
	PrepareStmt            = true
	DisableForeignKey      = false
)

func parseMysqlConfig(v *viper.Viper) (conf []mysqlConfig) {
	if v == nil {
		return
	}
	if err := v.UnmarshalKey("database", &conf); err != nil {
		panic("log 配置解析错误" + err.Error())
	}
	for key, item := range conf {
		if item.Level == 0 {
			conf[key].Level = 4
		}
		if item.ConnMaxLifetime == 0 {
			conf[key].ConnMaxLifetime = 120
		}
		if item.MaxOpenConn == 0 {
			conf[key].MaxOpenConn = 10
		}
		if item.MaxIdleConn == 0 {
			conf[key].MaxIdleConn = 5
		}
		if item.SlowThreshold == 0 {
			conf[key].SlowThreshold = 2
		}
		item.SlowThreshold = item.SlowThreshold * 1e6
		if item.SkipDefaultTransaction == nil {
			conf[key].SkipDefaultTransaction = &SkipDefaultTransaction
		}
		if item.SingularTable == nil {
			conf[key].SingularTable = &SingularTable
		}
		if item.DryRun == nil {
			conf[key].DryRun = &DryRun
		}
		if item.PrepareStmt == nil {
			conf[key].PrepareStmt = &PrepareStmt
		}
		if item.DisableForeignKey == nil {
			conf[key].DisableForeignKey = &DisableForeignKey
		}
	}
	return
}

func initMysql() {
	confList := parseMysqlConfig(globalConfig)
	clients := make(map[string]*gorm.DB)
	for _, conf := range confList {
		if !conf.Enable {
			continue
		}
		db, err := gorm.Open(mysql.Open(conf.Dsn), &gorm.Config{
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
			Logger:                 newMysqlLog(conf),
			SkipDefaultTransaction: *conf.SkipDefaultTransaction,
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   conf.TablePrefix,
				SingularTable: *conf.SingularTable,
			},
			DryRun:                                   *conf.DryRun,
			PrepareStmt:                              *conf.PrepareStmt,
			DisableForeignKeyConstraintWhenMigrating: *conf.DisableForeignKey,
		})
		if err != nil {
			panic(err)
		}
		clients[conf.Name] = db
		sdb, _ := db.DB()
		sdb.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second) //设置最大的链接时间
		sdb.SetMaxOpenConns(conf.MaxOpenConn)                                     //最大链接数量
		sdb.SetMaxIdleConns(conf.MaxIdleConn)                                     //最大闲置数量
	}
	globalMysqlConnects = clients
}
