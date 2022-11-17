package rest

import "github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw/commons"

type ReqCategory string

const (
	// ReqTypeCategoryUi the url prefix for accessing Angular apps devoted to public part of the interaction. The
	// url are of type /ui/<domain-name | root>/<site-name | _>/<it>/<app-home | ...>/extra path for angular router
	ReqTypeCategoryUi ReqCategory = "ui"

	// ReqTypeCategoryUiConsole the url prefix for accessing Angular console apps. The
	// url are of type /ui-console/<domain-name | root>/<site-name | _>/<it>/<app-home | ...>/extra path for angular router
	ReqTypeCategoryUiConsole ReqCategory = "ui-console"

	// ReqTypeCategoryApi the prefix for the api calls that are proxied to actual microservices.
	ReqTypeCategoryApi ReqCategory = "api"
)

// Adapters from commons.AppObjType and ReqCategory

func ReqCategoryFromAppObjType(appObjType commons.AppObjType) ReqCategory {
	at := ReqTypeCategoryUi
	if appObjType == commons.AppObjTypeConsole {
		at = ReqTypeCategoryUiConsole
	}

	return at
}

func AppObjTypeFromReqType(category ReqCategory) commons.AppObjType {
	at := commons.AppObjTypeWWW
	if category == ReqTypeCategoryUiConsole {
		at = commons.AppObjTypeConsole
	}

	return at
}

type ReqType string

const (
	ReqTypeApi       ReqType = "api"
	ReqTypeDomains   ReqType = "ui-domains"
	ReqTypeDomain    ReqType = "ui-domain"
	ReqTypeSiteApp   ReqType = "ui-domain-site-app"
	ReqTypeMalformed ReqType = "malformed"
	ReqTypeInvalid   ReqType = "invalid"
)

const (
	UnassignedTargetDomain = "root"
	UnassignedTargetSite   = "_"

	ReqTypeUiSiteAppHome        = "req-type-ui-site-app-home"
	ReqTypeUiConsoleSiteAppHome = "req-type-ui-console-site-app-home"

	// DefaultRouteToRoot it's the hard coded re-route in case of an invalid ui or ui-console url, or the access to the home.
	// can be overridden in config
	DefaultRouteToRoot = "/ui/" + UnassignedTargetDomain + "/" + UnassignedTargetSite + "/it/app-home"

	// AppsRootFolderContextParam the key name of the config param to hold the path of the Angular apps.
	// should be synched with the static route in the config.
	AppsRootFolderContextParam = "apps-root-folder"

	// RequestEnvironmentGinContextVarName the name used to store ReqEnv object extracted from mw-domain-site
	RequestEnvironmentGinContextVarName = "env"
)

type IdentType string

const (
	IdentTypeUnknown           IdentType = "unknown"
	IdentTypeSessionCookie     IdentType = "r3ds9-sid"
	IdentTypeSessionCookieName           = "r3ds9-sid"
)
