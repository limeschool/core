package core

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"time"
)

// zap 适配gorm 日志
type log struct {
	logger        *zap.Logger
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func newMysqlLog(conf mysqlConfig) logger.Interface {
	return &log{
		logger:        globalLog,
		LogLevel:      logger.LogLevel(conf.Level),
		SlowThreshold: time.Duration(conf.SlowThreshold),
	}
}

func (l *log) Log(ctx context.Context) *zap.Logger {
	id, _ := ctx.Value(TraceID).(string)
	return l.logger.With(zap.Any(TraceID, id))
}

// LogMode log mode
func (l *log) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

// Info print info
func (l log) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Log(ctx).Info("SQL信息", getSqlInfo("", fmt.Sprintf(msg, data...), 0, 0, false)...)
	}
}

// Warn print warn messages
func (l log) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Log(ctx).Info("SQL告警", getSqlInfo("", fmt.Sprintf(msg, data...), 0, 0, false)...)
	}
}

// Error print error messages
func (l log) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Log(ctx).Info("SQL错误", getSqlInfo("", fmt.Sprintf(msg, data...), 0, 0, false)...)
	}
}

// Trace print sql message
func (l log) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	costTime := float64(elapsed.Nanoseconds()) / 1e6
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound)):
		sql, rows := fc()
		l.Log(ctx).Info("SQL错误", getSqlInfo(err.Error(), sql, rows, costTime, false)...)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		l.Log(ctx).Info("SQL告警", getSqlInfo("", sql, rows, costTime, true)...)
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		l.Log(ctx).Info("SQL信息", getSqlInfo("", sql, rows, costTime, false)...)
	}
}

// 获取sql的执行信息
func getSqlInfo(err, sql string, rows int64, costTime float64, slow bool) []zap.Field {
	return []zap.Field{
		zap.String("err", err),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Float64("time", costTime),
		zap.Bool("slow", slow),
	}
}
