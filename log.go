package core

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

type LogConfig struct {
	Level         int8   `json:"level" mapstructure:"level"`
	Debug         bool   `json:"debug" mapstructure:"debug"`
	OutputConsole bool   `json:"output_console" mapstructure:"output_console"` // 是否输出到控制台
	OutputFile    bool   `json:"output_file" mapstructure:"output_file"`       // 是否输出到文件
	Filename      string `json:"filename" mapstructure:"filename"`             // 日志文件路径
	MaxSize       int    `json:"max_size" mapstructure:"max_size"`             // 每个日志文件保存的最大尺寸  单位：M
	MaxBackups    int    `json:"max_backups" mapstructure:"max_backups"`       // 日志文件最多保存多少个备份
	MaxAge        int    `json:"max_age" mapstructure:"max_age"`               // 文件最多保存多少天
	Compress      bool   `json:"compress" mapstructure:"compress"`             // 是否压缩
}

func parseLogConf(v *viper.Viper) LogConfig {
	conf := LogConfig{}
	if err := v.UnmarshalKey("log", &conf); err != nil {
		panic(err)
	}
	// 两个选项必须开启一个
	if !conf.OutputConsole && !conf.OutputFile {
		conf.OutputConsole = true
	}
	//当开启了输出到文件
	if !conf.OutputFile {
		return conf
	}

	if conf.Filename == "" {
		conf.Filename = "/logs/log.log"
	}
	if conf.MaxAge == 0 {
		conf.MaxAge = 7
	}
	if conf.MaxBackups == 0 {
		conf.MaxBackups = 3
	}
	if conf.MaxSize == 0 {
		conf.MaxSize = 10
	}
	return conf
}

// Log 链路日志
func newLog(id string) *zap.Logger {
	return globalLog.With(zap.Any(TraceID, id))
}

func initLogKey(v *viper.Viper) {
	trace := v.GetString("trace_key")
	if trace != "" {
		TraceID = trace
	}
}

func initLog(v *viper.Viper, srvName string) *zap.Logger {
	conf := parseLogConf(v)
	hook := lumberjack.Logger{
		Filename:   conf.Filename,   // 日志文件路径
		MaxSize:    conf.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: conf.MaxBackups, // 日志文件最多保存多少个备份
		MaxAge:     conf.MaxAge,     // 文件最多保存多少天
		Compress:   conf.Compress,   // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,                          // 小写编码器
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.Level(conf.Level))

	// 设置输出方式
	var syncOps []zapcore.WriteSyncer
	if conf.OutputFile {
		syncOps = append(syncOps, zapcore.AddSync(&hook))
	}
	if conf.OutputConsole {
		syncOps = append(syncOps, zapcore.AddSync(os.Stdout))
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),   // 编码器配置
		zapcore.NewMultiWriteSyncer(syncOps...), // 输出方式
		atomicLevel,                             // 日志级别
	)

	var ops []zap.Option
	// 开启开发模式，堆栈跟踪
	if conf.Debug {
		ops = append(ops, zap.AddCaller())
	}
	// 开启文件及行号,设置初始化字段
	ops = append(ops, zap.Development(), zap.Fields(zap.String("service", srvName)))
	// 构造日志
	return zap.New(core, ops...)
}
