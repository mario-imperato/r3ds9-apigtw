package api

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-apigtw/rest/middleware"
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
		Middlewares: []httpsrv.H{middleware.RequestApiEnvResolver(rest.ReqTypeCategoryApi), middleware.RequestUserResolver(), middleware.RequestUserAuthorizazion()},
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

		reqEnv := middleware.GetRequestEnvironmentFromContext(c)
		if reqEnv.IsMalformed() {
			c.String(http.StatusBadRequest, reqEnv.String())
			return
		}

		remote, err := url.Parse("http://localhost:3000")
		if err != nil {
			panic(err)
		}

		sid := reqEnv.AuthInfo.Sid

		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.URL.Path = "/r3ds9-auth/user"

			req.Header.Set("X-R3ds9-Api-Key", "pippo")
			req.Header.Set("X-R3ds9-Sid", sid)
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			hu := r.Header.Get("X-R3ds9-User")
			if hu != "" {
				log.Info().Str("sid", sid).Str("user", hu).Send()
				r.Header.Del("X-R3ds9-User")
			}
			r.Header.Set("x-R3ds9-Apigtw", "new header")
			return nil
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func redirectToPath(c *gin.Context, p string) {
	location := url.URL{Path: p}
	c.Redirect(http.StatusFound, location.RequestURI())
}
