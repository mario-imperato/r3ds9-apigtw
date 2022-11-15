package ui

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-apigtw/rest/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
	"os"
)

func init() {
	log.Info().Msg("ui endpoints init function")
	ra := httpsrv.GetApp()
	ra.RegisterGFactory(registerGroups)
}

func registerGroups(srvCtx httpsrv.ServerContext) []httpsrv.G {

	ctxParam, ok := srvCtx.GetConfig(rest.AppsRootFolderContextParam)
	if !ok {
		log.Error().Msgf("cannot find context param %s... skipping ui handler s config", rest.AppsRootFolderContextParam)
		return nil
	}

	appsRootFolder := ctxParam.(string)
	if _, err := os.Stat(appsRootFolder); err != nil {
		log.Error().Str(rest.AppsRootFolderContextParam, appsRootFolder).Msg("context param found but directory doesn't exists")
		return nil
	}

	gs := make([]httpsrv.G, 0, 2)

	gs = append(gs, httpsrv.G{
		Name:    "Ui Home",
		Path:    "/",
		AbsPath: true,
		Resources: []httpsrv.R{
			{
				Name:          "home",
				Path:          "",
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiRootHandler(appsRootFolder)},
			},
		},
	})

	gs = append(gs, httpsrv.G{
		Name:        "Ui Group",
		Path:        "/ui",
		AbsPath:     true,
		Middlewares: []httpsrv.H{middleware.RequestUiEnvResolver(rest.ReqTypeCategoryUi), middleware.RequestUserResolver(), middleware.RequestUserAuthorizazion()},
		Resources: []httpsrv.R{
			{
				Name:          "home-ui",
				Path:          ":domain/:site/:lang/:appId",
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler(appsRootFolder)},
			},
			{
				/*  :domain/*uiPath - Should present a selection of links to sites
				 *  :domain/:site/:lang/*uiPath - Is the website of :site
				 */
				Name:          "app path",
				Path:          ":domain/:site/:lang/:appId/*exPathInfo",
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler(appsRootFolder)},
			},
		},
	})

	gs = append(gs, httpsrv.G{
		Name:        "Ui Group",
		Path:        "/ui-console",
		AbsPath:     true,
		Middlewares: []httpsrv.H{middleware.RequestUiEnvResolver(rest.ReqTypeCategoryUiConsole), middleware.RequestUserResolver(), middleware.RequestUserAuthorizazion()},
		Resources: []httpsrv.R{
			{
				Name:          "console-domain_or_site",
				Path:          ":domain/:site/:lang/:appId",
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler(appsRootFolder)},
			},
			{
				/*
				 * :domain/*uiPath - Should present a selection of links to sites console and functions to create/delete a site.
				 * :domain/:site/:lang/:appName/*uiPath Is the console for the site...
				 */
				Name:          "console app path",
				Path:          ":domain/:site/:lang/:appId/*exPathInfo",
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler(appsRootFolder)},
			},
		},
	})

	return gs
}

func uiRootHandler(appsRootFolder string) httpsrv.H {
	return func(c *gin.Context) {
		redirectToPath(c, rest.DefaultRouteToRoot)
	}
}

func uiHandler(appsRootFolder string) httpsrv.H {
	return func(c *gin.Context) {
		reqEnv := middleware.GetRequestEnvironmentFromContext(c)
		c.String(http.StatusOK, reqEnv.String())
	}
}

func redirectToPath(c *gin.Context, p string) {
	location := url.URL{Path: p}
	c.Redirect(http.StatusFound, location.RequestURI())
}
