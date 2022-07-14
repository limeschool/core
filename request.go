package core

import (
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"time"
)

type httpToolConfig struct {
	EnableLog        bool          `json:"enable_log" mapstructure:"enable_log"`
	RetryCount       int           `json:"retry_count" mapstructure:"retry_count"`
	RetryWaitTime    time.Duration `json:"retry_wait_time" mapstructure:"retry_wait_time"`
	MaxRetryWaitTime time.Duration `json:"max_retry_wait_time" mapstructure:"max_retry_wait_time"`
	Timeout          time.Duration `json:"timeout" mapstructure:"timeout"`
	RequestMsg       string        `json:"request_msg" mapstructure:"request_msg"`
	ResponseMsg      string        `json:"response_msg" mapstructure:"response_msg"`
}

func initHttpToolConfig() *httpToolConfig {
	conf := &httpToolConfig{}
	if err := globalConfig.UnmarshalKey("http_tool", conf); err != nil {
		panic(err)
	}
	return conf
}

type httpTool struct {
	ctx     *Context
	request *resty.Request
}

type HttpToolFunc func(*resty.Request) *resty.Request

func (h *httpTool) Raw(fn HttpToolFunc) *httpTool {
	h.request = fn(h.request)
	return h
}

func (h *httpTool) log() {
	if !globalRequestConfig.EnableLog {
		return
	}
	logs := []zap.Field{
		zap.Any("method", h.request.Method),
		zap.Any("url", h.request.URL),
		zap.Any("header", h.request.Header),
		zap.Any("body", h.request.Body),
	}
	if len(h.request.FormData) != 0 {
		logs = append(logs, zap.Any("form-data", h.request.FormData))
	}
	if len(h.request.QueryParam) != 0 {
		logs = append(logs, zap.Any("query-data", h.request.QueryParam))
	}
	h.ctx.Log.Info(globalRequestConfig.RequestMsg, logs...)
}

func (h *httpTool) Get(url string) *httpResult {
	defer h.log()
	response := &httpResult{ctx: h.ctx}
	response.response, response.err = h.request.Get(url)
	return response
}

func (h *httpTool) Post(url string, data interface{}) *httpResult {
	defer h.log()
	response := &httpResult{ctx: h.ctx}
	response.response, response.err = h.request.SetBody(data).Post(url)
	return response
}

func (h *httpTool) PostJson(url string, data interface{}) *httpResult {
	defer h.log()
	response := &httpResult{ctx: h.ctx}
	response.response, response.err = h.request.ForceContentType("application/json").SetBody(data).Post(url)
	return response
}

func (h *httpTool) Put(url string, data interface{}) *httpResult {
	defer h.log()
	response := &httpResult{ctx: h.ctx}
	response.response, response.err = h.request.SetBody(data).Put(url)
	return response
}

func (h *httpTool) PutJson(url string, data interface{}) *httpResult {
	defer h.log()
	response := &httpResult{ctx: h.ctx}
	response.response, response.err = h.request.ForceContentType("application/json").SetBody(data).Put(url)
	return response
}

func (h *httpTool) Delete(url string) *httpResult {
	defer h.log()
	response := &httpResult{ctx: h.ctx}
	response.response, response.err = h.request.Delete(url)
	return response
}

type httpResult struct {
	ctx      *Context
	err      error
	response *resty.Response
}

func (h *httpResult) log() {
	if !globalRequestConfig.EnableLog {
		return
	}

	logs := []zap.Field{
		zap.Any("status", h.response.Status()),
		zap.Any("time", h.response.Time()),
		zap.Any("body", string(h.response.Body())),
		zap.Any("error", h.err),
	}
	h.ctx.Log.Info(globalRequestConfig.ResponseMsg, logs...)
}

func (r *httpResult) Body() ([]byte, error) {
	defer r.log()
	return r.response.Body(), r.err
}

func (r *httpResult) Result(val interface{}) error {
	defer r.log()
	if r.err != nil {
		return r.err
	}
	return json.Unmarshal(r.response.Body(), val)
}
