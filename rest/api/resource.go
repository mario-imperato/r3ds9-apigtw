package api

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ng-apigtw/rest/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func init() {
	log.Info().Msg("api endpoints init function")
	ra := httpsrv.GetApp()
	ra.RegisterGFactory(registerGroups)
}

func registerGroups(srvCtx httpsrv.ServerContext) []httpsrv.G {

	gs := make([]httpsrv.G, 0, 2)

	gs = append(gs, httpsrv.G{
		Name:        "Ui Home",
		Path:        ":domain/:site/:lang",
		Middlewares: []httpsrv.H{middleware.RequestEnvResolver("api")},
		Resources: []httpsrv.R{
			{
				Name:          "proxy",
				Path:          "*exPathInfo",
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{apiHandler()},
			},
		},
	})

	return gs
}

func apiHandler() httpsrv.H {
	return func(c *gin.Context) {

		uiPath := c.Param("exPathInfo")
		log.Info().Str("exPathInfo", uiPath).Send()

		remote, err := url.Parse("http://localhost:3000")
		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = "/r3ds9-auth/user"
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
