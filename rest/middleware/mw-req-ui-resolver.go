package middleware

import (
	"context"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/linkedservices"
	"github.com/mario-imperato/r3ds9-apigtw/linkedservices/mongodb"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-core/commons"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-core/domain"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-core/site"
	"github.com/rs/zerolog/log"
)

func RequestUiEnvResolver(reqCategory rest.ReqCategory) httpsrv.H {

	const semLogContext = "middleware/request-ui-env-resolver"
	return func(c *gin.Context) {
		env, ok := extractUiEnvFromContext(c, reqCategory)
		if !ok {
			redirectTo(c, rest.DefaultRouteToRoot)
			return
		}

		lks, err := linkedservices.GetMongoDbService(context.Background(), "r3ds9")
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			systemError(c, err)
			return
		}

		var redirectUrl string
		env, ok, redirectUrl = validateUiEnvFromStore(lks, env)
		if !ok {
			log.Error().Err(err).Msg(semLogContext)
			redirectTo(c, redirectUrl)
			return
		}

		SetRequestEnvironmentInContext(c, env)
	}
}

func extractUiEnvFromContext(c *gin.Context, reqCategory rest.ReqCategory) (ReqEnv, bool) {

	env := ReqEnv{
		Category:      reqCategory,
		Domain:        c.Param("domain"),
		Site:          c.Param("site"),
		Lang:          c.Param("lang"),
		App:           commons.App{Id: c.Param("appId")},
		ExtraPathInfo: c.Param("exPathInfo"),
	}

	env.ReqType = resolveUiRequestType(reqCategory, env.Domain, env.Site, env.Lang, env.App.Id)
	return env, env.IsValid()
}

// resolveRequestType syntax level resolution. doesn't check if domain or site do really exist.
func resolveUiRequestType(mountPoint rest.ReqCategory, domain, site, lang string, appName string) rest.ReqType {

	at := rest.ReqTypeDomains
	if domain != rest.UnassignedTargetDomain {
		if site != rest.UnassignedTargetSite {
			at = rest.ReqTypeSiteApp
		} else {
			at = rest.ReqTypeDomain
		}
	} else if site != rest.UnassignedTargetSite {
		at = rest.ReqTypeMalformed
	}

	// This check is only to check against the developed apps... Is quickier in attack case than looking up somethin in the Db.
	if !commons.IsAppIdInCatalog(appName) {
		at = rest.ReqTypeMalformed
	}

	return at
}

func validateUiEnvFromStore(lks *mongodb.MDbLinkedService, env ReqEnv) (ReqEnv, bool, string) {

	const semLogContext = "middleware/request-ui-env-resolver/validate-env-from-store"
	var ok bool
	redirectPath := rest.DefaultRouteToRoot
	switch env.ReqType {
	case rest.ReqTypeDomains:
		// In this case is the 'root' domain.
		var d *domain.Domain
		coll := lks.GetCollection("domain", "")
		d, ok = domain.GetFromCache(domain.NewCacheResolver(coll), env.Domain)
		if ok {
			if env.App, ok = d.GetAppByObjTypeAndId(rest.AppObjTypeFromReqType(env.Category), env.App.Id); !ok {
				if env.App.ObjType != "" {
					// Found a matching item by id...
					redirectPath = RequestPath4(rest.ReqCategoryFromAppObjType(commons.AppObjType(env.App.ObjType)), env.Domain, env.Site, env.Lang, env.App.Id, "")
				}
				log.Warn().Str("domain", env.Domain).Str("redirect", redirectPath).Str("app-id", env.App.Id).Msg(semLogContext + " invalid app id")
			}
		}
	case rest.ReqTypeDomain:
		var d *domain.Domain
		coll := lks.GetCollection("domain", "")
		d, ok = domain.GetFromCache(domain.NewCacheResolver(coll), env.Domain)
		if ok {
			if env.App, ok = d.GetAppByObjTypeAndId(rest.AppObjTypeFromReqType(env.Category), env.App.Id); !ok {
				if env.App.ObjType != "" {
					// Found a matching item by id...
					redirectPath = RequestPath4(rest.ReqCategoryFromAppObjType(commons.AppObjType(env.App.ObjType)), env.Domain, env.Site, env.Lang, env.App.Id, "")
				}
				log.Warn().Str("domain", env.Domain).Str("redirect", redirectPath).Str("app-id", env.App.Id).Msg(semLogContext + " invalid app id")
			}
		}
	case rest.ReqTypeSiteApp:
		var s *site.Site
		coll := lks.GetCollection("site", "")
		s, ok = site.GetFromCache(site.NewCacheResolver(coll), env.Domain, env.Site)
		if ok {
			if env.App, ok = s.GetAppByObjTypeAndId(rest.AppObjTypeFromReqType(env.Category), env.App.Id); !ok {
				if env.App.ObjType != "" {
					// Found a matching item by id...
					redirectPath = RequestPath4(rest.ReqCategoryFromAppObjType(commons.AppObjType(env.App.ObjType)), env.Domain, env.Site, env.Lang, env.App.Id, "")
				}
				log.Warn().Str("domain", env.Domain).Str("site", env.Site).Str("redirect", redirectPath).Str("app-id", env.App.Id).Msg(semLogContext + " invalid app id")
			}
		} else {
			coll = lks.GetCollection("domain", "")
			_, ok2 := domain.GetFromCache(domain.NewCacheResolver(coll), env.Domain)
			if ok2 {
				// Don't have an app to look into... redirect to home of domain.
				redirectPath = RequestPath4(rest.ReqTypeCategoryUi, env.Domain, rest.UnassignedTargetSite, env.Lang, string(commons.AppIdHome), "")
			}
		}
	default:
		ok = true
	}

	return env, ok, redirectPath
}
