package decorator

import (
	"net/http"
	"regexp"
	"strings"

	restful "github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	"gomod.alauda.cn/log"
)

// NewRewriteRouter initiates a router for rewritting requests
// using a prefix
func NewRewriteRouter(path string, logger *zap.Logger) RewriteRouter {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	logg := logger.With(log.String("prefix", path))
	regex := strings.ReplaceAll(path, "/", `\/`) + `\/`
	pathRegex := regexp.MustCompile(regex + ".*")
	return RewriteRouter{
		PrefixPath: path,
		log:        logg,
		regex:      pathRegex,
	}
}

// RewriteRouter implementation of a rewrite router
type RewriteRouter struct {
	restful.CurlyRouter
	PrefixPath string
	log        *zap.Logger
	regex      *regexp.Regexp
}

// SelectRoute implements the restful.Router
func (r RewriteRouter) SelectRoute(
	webServices []*restful.WebService,
	req *http.Request) (selectedService *restful.WebService, selected *restful.Route, err error) {
	if r.regex.MatchString(req.URL.Path) {
		newPath := strings.Replace(req.URL.Path, r.PrefixPath, "", 1)
		r.log.Debug("::: rewriting path", log.String("old", req.URL.Path), log.String("new", newPath))
		req.URL.Path = newPath
	}
	return r.CurlyRouter.SelectRoute(webServices, req)
}
