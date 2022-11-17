package api

import (
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-apigtw/rest/middleware"
	"github.com/mitchellh/mapstructure"
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

type ProxyConfig struct {
	Hostname string `yaml:"hostname" mapstructure:"hostname"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Prefix   string `yaml:"prefix" mapstructure:"prefix"`
	Scheme   string `yaml:"scheme" mapstructure:"scheme"`
}

func (cfg ProxyConfig) Host() string {
	return fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port)
}

func (cfg ProxyConfig) Remote() string {
	return fmt.Sprintf("%s://%s:%d", cfg.Scheme, cfg.Hostname, cfg.Port)
}

var proxyConfig map[string]ProxyConfig

func registerGroups(srvCtx httpsrv.ServerContext) []httpsrv.G {

	const semLogContext = "/api/resource/register-groups"
	if m, ok := srvCtx.GetConfig("proxy-mappings"); ok {
		pm, ok := m.(map[string]interface{})
		if ok {
			proxyConfig = make(map[string]ProxyConfig)
			for k, v := range pm {
				pxcfg := ProxyConfig{
					Hostname: "localhost",
					Port:     80,
					Scheme:   "http",
					Prefix:   "/api",
				}
				if err := mapstructure.Decode(v, &pxcfg); err != nil {
					log.Error().Err(err).Str("config-key", k).Msg(semLogContext + " - decoding proxy config")
				} else {
					proxyConfig[k] = pxcfg
				}
			}
		} else {
			log.Error().Msg(semLogContext + " - proxy-mappings is not a map...")
		}
	} else {
		log.Error().Msg(semLogContext + " - proxy config not found")
	}

	gs := make([]httpsrv.G, 0, 2)

	gs = append(gs, httpsrv.G{
		Name:        "Ui Home",
		Path:        ":domain/:site/:lang",
		Middlewares: []httpsrv.H{middleware.RequestApiEnvResolver(rest.ReqTypeCategoryApi), middleware.RequestUserResolver(), middleware.RequestUserAuthorizazion()},
		Resources: []httpsrv.R{
			{
				Name:          "proxy",
				Path:          ":wsCtx/*exPathInfo",
				Method:        httpsrv.MethodAny,
				RouteHandlers: []httpsrv.H{apiHandler()},
			},
		},
	})

	return gs
}

func apiHandler() httpsrv.H {
	return func(c *gin.Context) {

		const semLogContext = "/api/resource/apiHandler"

		reqEnv := middleware.GetRequestEnvironmentFromContext(c)
		if reqEnv.IsMalformed() {
			c.String(http.StatusBadRequest, reqEnv.String())
			return
		}

		cfg, ok := proxyConfig[c.Param("wsCtx")]
		if !ok {
			log.Error().Str("ws-ctx", c.Param("wsCtx")).Msg(semLogContext + " ws context not found")
			c.String(http.StatusBadRequest, reqEnv.String())
			return
		}

		remote, err := url.Parse(cfg.Remote())
		if err != nil {
			log.Error().Str("remote", cfg.Remote()).Msg(semLogContext + " error parsing remote")
			c.String(http.StatusBadRequest, reqEnv.String())
			return
		}

		sid := reqEnv.AuthInfo.Sid

		proxy := httputil.NewSingleHostReverseProxy(remote)
		director := proxy.Director
		proxy.Director = func(req *http.Request) {
			director(req)

			req.Header = c.Request.Header
			req.Header.Set("X-R3ds9-Api-Key", "pippo")
			req.Header.Set("X-R3ds9-Sid", sid)

			req.Host = cfg.Host()
			//req.URL.Scheme = cfg.Scheme
			//req.URL.Host = cfg.Host()

			path := cfg.Prefix + c.Param("exPathInfo")
			log.Trace().Str("remote-path", path).Msg(semLogContext)
			req.URL.Path = path
			//req.URL.RawQuery = c.Request.URL.RawQuery

			log.Trace().Interface("raw-query", req.URL).Send()
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
