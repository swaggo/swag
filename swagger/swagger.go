//hahahaha
package swagger

import "github.com/gin-gonic/gin"

type Engine struct {
	routes gin.RoutesInfo

	basePath string
}

func New(routes gin.RoutesInfo) *Engine {
	engine := &Engine{
		basePath: "/swagger",
		routes:   routes,
	}
	return engine
}

// @WTF
func (s *Engine) Routes() gin.RoutesInfo {
	return s.routes
}

// hello
func (s *Engine) Build() *Engine {
	//TODO: 1.parsing annotate to swagger doc
	// TODO : generate  router for swagger
	//HH
	return s
}

func (s *Engine) parseApiSpec() *Engine {

	return s
}
