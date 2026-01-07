package httpx

import "github.com/gin-gonic/gin"

type Router struct {
	engine *gin.Engine
	group  *gin.RouterGroup
}

// Root router
func NewRouter(engine *gin.Engine) *Router {
	return &Router{
		engine: engine,
	}
}

// Group router
func newGroup(group *gin.RouterGroup) *Router {
	return &Router{
		group: group,
	}
}

// ---------- Routing ----------

func (r *Router) GET(path string, h AppHandler) {
	r.handle("GET", path, h)
}

func (r *Router) POST(path string, h AppHandler) {
	r.handle("POST", path, h)
}

func (r *Router) PUT(path string, h AppHandler) {
	r.handle("PUT", path, h)
}

func (r *Router) DELETE(path string, h AppHandler) {
	r.handle("DELETE", path, h)
}

func (r *Router) handle(method, path string, h AppHandler) {
	handler := Adapt(h)

	if r.group != nil {
		r.group.Handle(method, path, handler)
		return
	}

	r.engine.Handle(method, path, handler)
}

// ---------- Groups ----------

func (r *Router) Group(relativePath string, middleware ...gin.HandlerFunc) *Router {
	var g *gin.RouterGroup

	if r.group != nil {
		g = r.group.Group(relativePath, middleware...)
	} else {
		g = r.engine.Group(relativePath, middleware...)
	}

	return newGroup(g)
}
