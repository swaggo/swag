package swagger

import "github.com/gin-gonic/gin"

type Engine struct {
	routes gin.RoutesInfo

	basePath string
}

func New(routes gin.RoutesInfo) *Engine {
	engine := &Engine{
		basePath: "/swagger-ui",
		routes:   routes,
	}
	return engine
}

// @WTF
func (s *Engine) Routes() gin.RoutesInfo {
	return s.routes
}

func (s *Engine) Build() *Engine {
	//TODO: 1.parsing annotate to swagger doc
	// TODO : generate  router for swagger
	s.parseApiSpec()
	return s
}

func (s *Engine) parseApiSpec() *Engine {

	return s
}
