package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Context struct {
	context.Context
	Writer     http.ResponseWriter
	Request    *http.Request
	Path       string
	Method     string
	StatusCode int
	Params     map[string]string
	handlers   []HandlerFunc
	index      int
	engine     *engine
	TraceID    string      //链路ID
	Log        *zap.Logger //链路日志
	Config     *config     //配置中心
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Path:    r.URL.Path,
		Method:  r.Method,
		index:   -1,
		Context: context.Background(),
	}
}

func (c *Context) SrvName() string {
	return globalServiceName
}

func (c *Context) Mysql(db string) *gorm.DB {
	return globalMysqlConnects[db]
}

func (c *Context) Mongo(db string) *mongo.Client {
	return globalMongoConnects[db]
}

func (c *Context) Redis() *redis.Client {
	return globalRedisConnect
}

func (c *Context) SetValue(key string, value interface{}) {
	c.Context = context.WithValue(c.Context, key, value)
}

func (c *Context) GetString(key string) string {
	val, _ := c.Context.Value(key).(string)
	return val
}

func (c *Context) GetBool(key string) bool {
	val, _ := c.Context.Value(key).(bool)
	return val
}

func (c *Context) GetInt(key string) int {
	val, _ := c.Context.Value(key).(int)
	return val
}

func (c *Context) GetFloat64(key string) float64 {
	val, _ := c.Context.Value(key).(float64)
	return val
}

func (c *Context) GetStringSlice(key string) []string {
	val, _ := c.Context.Value(key).([]string)
	return val
}

func (c *Context) GetIntSlice(key string) []int {
	val, _ := c.Context.Value(key).([]int)
	return val
}

func (c *Context) GetFloat64Slice(key string) []float64 {
	val, _ := c.Context.Value(key).([]float64)
	return val
}

func (c *Context) GetMapString(key string) map[string]interface{} {
	val, _ := c.Context.Value(key).(map[string]interface{})
	return val
}

func (c *Context) UnmarshalKey(key string, val interface{}) error {
	data := c.Context.Value(key)
	byteData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(byteData, val)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key, val string) {
	c.Writer.Header().Set(key, val)
}

func (c *Context) AddHeader(key, val string) {
	c.Writer.Header().Add(key, val)
}

func (c *Context) GetHeader(key string) string {
	return c.Writer.Header().Get(key)
}

func (c *Context) GetHeaders(key string) []string {
	return c.Writer.Header().Values(key)
}

func (c *Context) DelHeader(key string) {
	c.Writer.Header().Del(key)
}

func (c *Context) JSON(code int, data interface{}) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(data); err != nil {
		panic(err)
	}
}

func (c *Context) String(code int, format string, arg ...interface{}) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, arg...)))
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

func (c *Context) XML(code int, html string) {
	c.Writer.Header().Set("Content-Type", "text/xml")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) Param(key string) string {
	return c.Params[key]
}

func (c *Context) Next() {
	c.index++
	length := len(c.handlers)
	for ; c.index < length; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Abort() {
	c.index = len(c.handlers)
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}

func (c *Context) HttpTool() *httpTool {
	client := resty.New()

	client.SetRetryCount(globalRequestConfig.RetryCount).
		SetRetryWaitTime(time.Duration(globalRequestConfig.RetryWaitTime*1000)*time.Millisecond).
		SetRetryMaxWaitTime(time.Duration((globalRequestConfig.RetryWaitTime+2)*1000)*time.Millisecond).
		SetTimeout(globalRequestConfig.Timeout).
		SetHeader(TraceID, c.TraceID).
		SetHeader("User-Agent", c.SrvName()).
		SetHeader("Remote-Service", c.SrvName())

	return &httpTool{ctx: c, request: client.R()}
}
