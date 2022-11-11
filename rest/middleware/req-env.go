package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ng-apigtw/constants"
)

type ReqEnv struct {
	Category      string
	ReqType       string
	Domain        string
	Site          string
	Lang          string
	AppName       string
	ExtraPathInfo string
}

func extractEnvFromContext(c *gin.Context, reqCategory string) (ReqEnv, bool) {

	env := ReqEnv{
		Category:      reqCategory,
		Domain:        c.Param("domain"),
		Site:          c.Param("site"),
		Lang:          c.Param("lang"),
		AppName:       c.Param("appName"),
		ExtraPathInfo: c.Param("exPathInfo"),
	}

	env.ReqType = resolveRequestType(reqCategory, env.Domain, env.Site, env.Lang, env.AppName)
	return env, true
}

func resolveRequestType(mountPoint string, domain, site, lang, appName string) string {

	at := constants.AppTypeUiDomains
	if domain != "domains" {
		if site != "sites" {
			at = constants.AppTypeUiSiteAppHome
			switch appName {
			case constants.AppHome:
				at = constants.AppTypeUiSiteAppHome
			default:
				at = constants.AppTypeUiSiteAppHome
			}
		} else {
			at = constants.AppTypeUiDomain
		}
	}

	switch mountPoint {
	case "ui-console":
		switch at {
		case constants.AppTypeUiDomains:
			at = constants.AppTypeUiConsoleDomains
		case constants.AppTypeUiDomain:
			at = constants.AppTypeUiConsoleDomain
		case constants.AppTypeUiSiteAppHome:
			at = constants.AppTypeUiConsoleSiteAppHome
		}
	case "api":
		at = "To-be-defined"
	}

	return at
}
