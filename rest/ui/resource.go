package ui

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-common/util"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apicommon/definitions"
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

	/* Not used any more
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
	*/

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
				RouteHandlers: []httpsrv.H{uiRootHandler()},
			},
		},
	})

	gs = append(gs, httpsrv.G{
		Name:        "Ui Group",
		Path:        definitions.ApiGtwUiGroup,
		AbsPath:     true,
		Middlewares: []httpsrv.H{middleware.RequestUiEnvResolver(rest.ReqTypeCategoryUi), middleware.RequestUserResolver(false), middleware.RequestUserAuthorizazion()},
		Resources: []httpsrv.R{
			{
				Name:          "home-ui",
				Path:          definitions.ApiGtwUiPathPatternHome,
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler()},
			},
			{
				/*  :domain/*uiPath - Should present a selection of links to sites
				 *  :domain/:site/:lang/*uiPath - Is the website of :site
				 */
				Name:          "app path",
				Path:          definitions.ApiGtwUiPathPatternNested,
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler()},
			},
		},
	})

	gs = append(gs, httpsrv.G{
		Name:        "Ui Group",
		Path:        definitions.ApiGtwUiConsoleGroup,
		AbsPath:     true,
		Middlewares: []httpsrv.H{middleware.RequestUiEnvResolver(rest.ReqTypeCategoryUiConsole), middleware.RequestUserResolver(false), middleware.RequestUserAuthorizazion()},
		Resources: []httpsrv.R{
			{
				Name:          "console-domain_or_site",
				Path:          definitions.ApiGtwUiPathPatternHome,
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler()},
			},
			{
				/*
				 * :domain/*uiPath - Should present a selection of links to sites console and functions to create/delete a site.
				 * :domain/:site/:lang/:appName/*uiPath Is the console for the site...
				 */
				Name:          "console app path",
				Path:          definitions.ApiGtwUiPathPatternNested,
				Method:        http.MethodGet,
				RouteHandlers: []httpsrv.H{uiHandler()},
			},
		},
	})

	return gs
}

func uiRootHandler() httpsrv.H {
	return func(c *gin.Context) {
		redirectToPath(c, rest.DefaultRouteToRoot)
	}
}

func uiHandler() httpsrv.H {

	const semLogContext = "/ui/resource/uiHandler"
	return func(c *gin.Context) {
		reqEnv := middleware.GetRequestEnvironmentFromContext(c)

		//appPath := filepath.Join(reqEnv.App.Path, "index.tmpl")
		appPath := reqEnv.App.Path
		log.Trace().Str("app-index", appPath).Msg(semLogContext + " - fetching app index file")
		c.HTML(http.StatusOK, appPath, reqEnv)
		/*
			appIndex, err := resolveAppIndex(reqEnv, reqEnv.App.Id, appPath)
			if err != nil {
				log.Error().Err(err).Str("app-index", appPath).Msg(semLogContext)
				c.String(http.StatusOK, reqEnv.String())
			} else {
				c.HTML(http.StatusOK, appPath, reqEnv)
			}
		*/
	}
}

func resolveAppIndex(env middleware.ReqEnv, appName, appPath string) (string, error) {

	const semLogContext = "/ui/resource/resolve-app-index"
	if _, err := os.Stat(appPath); err != nil {
		log.Error().Err(err).Str("path", appPath).Msg(semLogContext)
		return "", err
	}

	indexTmpl, err := os.ReadFile(appPath)
	if err != nil {
		log.Error().Err(err).Str("path", appPath).Msg(semLogContext)
		return "", err
	}

	tmpl, err := util.ParseTemplates([]util.TemplateInfo{
		{Name: appName, Content: string(indexTmpl)},
	}, nil)

	if err != nil {
		log.Error().Err(err).Str("path", appPath).Msg(semLogContext)
		return "", err
	}

	appIndex, err := util.ProcessTemplate(tmpl, env, false)
	if err != nil {
		log.Error().Err(err).Str("path", appPath).Msg(semLogContext)
		return "", err
	}

	return string(appIndex), nil
}

func redirectToPath(c *gin.Context, p string) {
	location := url.URL{Path: p}
	c.Redirect(http.StatusFound, location.RequestURI())
}
