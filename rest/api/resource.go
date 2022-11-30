package api

import (
	"context"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apicommon/definitions"
	"github.com/mario-imperato/r3ds9-apicommon/linkedservices"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-apigtw/rest/middleware"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw/session"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func init() {
	log.Info().Msg("api endpoints init function")
	ra := httpsrv.GetApp()
	ra.RegisterGFactory(registerGroups)
}

type ProxyConfig struct {
	Hostname string `yaml:"hostname" mapstructure:"hostname"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Url      string `yaml:"url" mapstructure:"url"`
	Scheme   string `yaml:"scheme" mapstructure:"scheme"`
	ApiKey   string `yaml:"api-key" mapstructure:"api-key"`
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
					Url:      "/api/:apiCtx/:hostDomain/:hostSite/:hostLang*exPathInfo",
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
		Name:        "Api Home",
		Path:        definitions.ApiGtwApiGroupPattern,
		Middlewares: []httpsrv.H{middleware.RequestApiEnvResolver(rest.ReqTypeCategoryApi), middleware.RequestUserResolver(true)},
		Resources: []httpsrv.R{
			{
				Name:          "proxy",
				Path:          "*" + definitions.ApiGtwExtraPathInfoUrlParam,
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

		wsCtx := c.Param(definitions.ApiContextUrlParam)
		cfg, ok := proxyConfig[wsCtx]
		if !ok {
			log.Error().Str("ws-ctx", wsCtx).Msg(semLogContext + " ws context not found")
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

			/*
			 * Set up the headers.
			 */
			req.Header = c.Request.Header
			if cfg.ApiKey != "" {
				req.Header.Set(definitions.ApiKeyHeaderName, cfg.ApiKey)
			}

			if sid != "" {
				req.Header.Set(definitions.SidHeaderName, sid)
				req.Header.Set(definitions.UserHeaderName, reqEnv.AuthInfo.User.Nickname)
			}

			req.Host = cfg.Host()
			path := resolveRemotePath(cfg.Url, wsCtx, reqEnv.Domain, reqEnv.Site, reqEnv.Lang, reqEnv.ExtraPathInfo)
			log.Trace().Str("remote-path", path).Msg(semLogContext)
			req.URL.Path = path
			//req.URL.RawQuery = c.Request.URL.RawQuery

			log.Trace().Interface("raw-query", req.URL).Send()
		}

		proxy.ModifyResponse = func(r *http.Response) error {
			hu := r.Header.Get(definitions.UserHeaderName)
			if hu != "" {
				if hu != reqEnv.AuthInfo.User.Nickname {
					// The service returned a different nickname. Should promote the session on the new nickname
					if err := switchSessionNickname(sid, reqEnv.AuthInfo.User.Nickname, hu); err != nil {
						log.Error().Err(err).Msg(semLogContext)
					}
				}
				r.Header.Del(definitions.UserHeaderName)
			}
			return nil
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func redirectToPath(c *gin.Context, p string) {
	location := url.URL{Path: p}
	c.Redirect(http.StatusFound, location.RequestURI())
}

func resolveRemotePath(remotePath string, wsCtx, domain, site, lang, exPathInfo string) string {
	remotePath = strings.ReplaceAll(remotePath, ":"+definitions.ApiContextUrlParam, wsCtx)
	remotePath = strings.ReplaceAll(remotePath, ":"+definitions.HostDomainUrlParam, domain)
	remotePath = strings.ReplaceAll(remotePath, ":"+definitions.HostSiteUrlParam, site)
	remotePath = strings.ReplaceAll(remotePath, ":"+definitions.HostLangUrlParam, lang)
	remotePath = strings.ReplaceAll(remotePath, ":"+definitions.ApiGtwExtraPathInfoUrlParam, exPathInfo)
	return remotePath
}

func switchSessionNickname(sid string, oldNickname, newNickname string) error {
	const semLogContext = "/api/resource/apiHandler/switchSessionNickname"
	log.Info().Str("sid", sid).Str("old-user", oldNickname).Str("new-user", newNickname).Msg(semLogContext)

	lks, err := linkedservices.GetMongoDbService(context.Background(), "r3ds9")
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return err
	}

	coll := lks.GetCollection("session", "")
	_, err = session.UpdateBySid(context.TODO(), coll, sid, true, session.UpdateWithNickname(newNickname))
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return err
	}

	/*
	 * TODO: should propagate an event of session update to all api-gtw. For now simply invalidate...
	 */
	session.InvalidateSession(sid)
	return nil
}
