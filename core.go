package core

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type routerGroup struct {
	prefix      string
	middlewares []HandlerFunc
	parent      *routerGroup
	engine      *engine
}

type engine struct {
	*routerGroup
	router        *router
	groups        []*routerGroup
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
}

func (e *engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

func New(srvName string) *engine {
	globalServiceName = srvName
	// 初始化配置信息
	initConfig()
	// 初始化请求配置信息
	initHttpToolConfig()
	// 初始化日志信息
	globalLog = initLog(globalConfig, srvName)

	// 初始化路由
	e := &engine{router: newRouter()}
	e.routerGroup = &routerGroup{engine: e}
	e.groups = []*routerGroup{e.routerGroup}
	//注册中间件
	e.Use(timeout(), cpuLoad(), recovery(), traceLog(), ipLimit())
	return e
}

func (group *routerGroup) Group(prefix string) *routerGroup {
	e := group.engine
	newGroup := &routerGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: e,
	}
	e.groups = append(e.groups, newGroup)
	return newGroup
}

func (group *routerGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *routerGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *routerGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// PUT defines the method to add PUT request
func (group *routerGroup) PUT(pattern string, handler HandlerFunc) {
	group.addRoute("PUT", pattern, handler)
}

// DELETE defines the method to add DELETE request
func (group *routerGroup) DELETE(pattern string, handler HandlerFunc) {
	group.addRoute("DELETE", pattern, handler)
}

func (e *engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, r)
	c.engine = e
	c.handlers = middlewares
	e.router.handler(c)
}

func (e *engine) Run(port string) error {
	return http.ListenAndServe(port, e)
}

func (group *routerGroup) Use(middlewares ...HandlerFunc) *engine {
	group.middlewares = append(group.middlewares, middlewares...)
	return group.engine
}

func (group *routerGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

func (group *routerGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}
