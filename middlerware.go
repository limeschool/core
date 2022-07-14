package core

import (
	"context"
	"fmt"
	"github.com/didip/tollbooth"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/load"
	"go.uber.org/zap"
	"net/http"
	"runtime"
)

func recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				c.Log.Error(message, zap.Any("trace", tracePanicErr()))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}

func tracePanicErr() []string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller
	var arr []string
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		arr = append(arr, fmt.Sprintf("%s:%d", file, line))
	}
	return arr
}

func traceLog() HandlerFunc {
	return func(ctx *Context) {
		trace := ctx.GetHeader(TraceID)
		if trace == "" {
			trace = uuid.New().String()
		}
		ctx.TraceID = trace
		ctx.SetValue(TraceID, trace)
		ctx.Log = newLog(trace)
		ctx.Config = newConfig(ctx.Log)
	}
}

// ipLimit ip限流
func ipLimit() HandlerFunc {
	max := globalConfig.GetFloat64("ip_limit.max")
	limit := tollbooth.NewLimiter(max, nil)
	return func(ctx *Context) {
		if httpError := tollbooth.LimitByRequest(limit, ctx.Writer, ctx.Request); httpError != nil {
			ctx.Fail(400, "ip request fail")
		}
	}
}

// CpuLoad 自适应降载
func cpuLoad() HandlerFunc {
	sd := load.NewAdaptiveShedder(load.WithCpuThreshold(0))
	return func(ctx *Context) {
		promise, err := sd.Allow()
		if err != nil {
			ctx.Fail(500, "系统繁忙，请稍后再试")
		}
		promise.Pass()
	}
}

// CpuLoad 自适应降载
func timeout() HandlerFunc {
	return func(ctx *Context) {
		called := make(chan bool)
		uctx, cancel := context.WithTimeout(ctx.Context, globalSystemConfig.Timeout)
		defer cancel()

		go func() {
			ctx.Next()
			<-called
		}()

		if globalSystemConfig.Timeout == 0 {
			called <- true
			return
		}

		select {
		case <-uctx.Done():
			close(called)
			ctx.Fail(500, "request timeout")
		case called <- true:
		}
	}
}
