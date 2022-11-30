package middleware

import (
	"context"
	"errors"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apicommon/definitions"
	"github.com/mario-imperato/r3ds9-apicommon/linkedservices"
	"github.com/mario-imperato/r3ds9-apicommon/linkedservices/mongodb"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw/domain"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw/site"
	"github.com/rs/zerolog/log"
	"net/http"
)

func RequestApiEnvResolver(reqCategory rest.ReqCategory) httpsrv.H {

	const semLogContext = "middleware/request-api-env-resolver"
	return func(c *gin.Context) {
		env, ok := extractApiEnvFromContext(c, reqCategory)
		if !ok {
			c.AbortWithError(http.StatusBadRequest, errors.New("bad request"))
			return
		}

		lks, err := linkedservices.GetMongoDbService(context.Background(), "r3ds9")
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			systemError(c, err)
			return
		}

		env, ok = validateApiRequestFromStore(lks, env)
		if !ok {
			log.Error().Err(err).Msg(semLogContext)
			c.AbortWithError(http.StatusBadRequest, errors.New("bad request"))
			return
		}

		SetRequestEnvironmentInContext(c, env)
	}
}

func extractApiEnvFromContext(c *gin.Context, reqCategory rest.ReqCategory) (ReqEnv, bool) {

	env := ReqEnv{
		Category:      reqCategory,
		Domain:        c.Param(definitions.HostDomainUrlParam),
		Site:          c.Param(definitions.HostSiteUrlParam),
		Lang:          c.Param(definitions.HostLangUrlParam),
		ExtraPathInfo: c.Param(definitions.ApiGtwExtraPathInfoUrlParam),
	}

	env.ReqType = resolveApiRequestType(reqCategory, env.Domain, env.Site, env.Lang)
	return env, env.IsValid()
}

// resolveRequestType syntax level resolution. doesn't check if domain or site do really exist.
func resolveApiRequestType(mountPoint rest.ReqCategory, domain, site, lang string) rest.ReqType {

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

	return at
}

func validateApiRequestFromStore(lks *mongodb.MDbLinkedService, env ReqEnv) (ReqEnv, bool) {

	const semLogContext = "middleware/request-api-env-resolver/validate-env-from-store"
	var ok bool
	switch env.ReqType {
	case rest.ReqTypeDomains:
		// In this case is the 'root' domain.
		coll := lks.GetCollection("domain", "")
		_, ok = domain.GetFromCache(domain.NewCacheResolver(coll), env.Domain)

	case rest.ReqTypeDomain:
		coll := lks.GetCollection("domain", "")
		_, ok = domain.GetFromCache(domain.NewCacheResolver(coll), env.Domain)

	case rest.ReqTypeSiteApp:
		coll := lks.GetCollection("site", "")
		_, ok = site.GetFromCache(site.NewCacheResolver(coll), env.Domain, env.Site)

	default:
		ok = true
	}

	return env, ok
}
