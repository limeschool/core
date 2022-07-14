package core

import (
	"strings"
)

type HandlerFunc func(ctx *Context)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

func (e *router) addRoute(method string, pattern string, f HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := e.roots[method]
	if !ok {
		e.roots[method] = &node{}
	}
	e.roots[method].insert(pattern, parts, 0)
	e.handlers[key] = f
}

func (e *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := e.roots[method]
	if !ok {
		return nil, nil
	}
	n := root.search(searchParts, 0)
	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

func (e *router) handler(ctx *Context) {
	n, params := e.getRoute(ctx.Method, ctx.Path)
	if n != nil {
		ctx.Params = params
		key := ctx.Method + "-" + n.pattern
		ctx.handlers = append(ctx.handlers, e.handlers[key])
	} else {
		ctx.String(404, "NOT FOUND URL %v", ctx.Path)
	}
	ctx.Next()
}
