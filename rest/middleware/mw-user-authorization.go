package middleware

import (
	"context"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/linkedservices"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw/session"
	"github.com/rs/zerolog/log"
)

func RequestUserAuthorizazion() httpsrv.H {
	const semLogContext = "middleware/request-user-authorization"
	return func(c *gin.Context) {

		lks, err := linkedservices.GetMongoDbService(context.Background(), "r3ds9")
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			systemError(c, err)
			return
		}

		env := GetRequestEnvironmentFromContext(c)

		ok := true
		if env.IsUi() && env.App.RoleRequired {
			ok = verifyUserAppAuthorization(env)
			if !ok {
				log.Warn().Str("nickname", env.AuthInfo.User.Nickname).Str("domain", env.Domain).Str("site", env.Site).Str("app-id", env.App.Id).Msg(semLogContext + " unauthorized")
				redirectTo(c, rest.DefaultRouteToRoot)
				return
			}
		}

		if env.AuthInfo.NewSid {
			s := session.Session{
				Nickname:   env.AuthInfo.User.Nickname,
				Remoteaddr: env.AuthInfo.Remoteaddr,
			}

			coll := lks.GetCollection("session", "")
			sid, err := session.Insert(context.TODO(), coll, &s)
			if err != nil {
				log.Error().Err(err).Msg(semLogContext)
				redirectTo(c, rest.DefaultRouteToRoot)
			}

			env.AuthInfo.Sid = sid
			SetRequestEnvironmentInContext(c, env)
			c.SetCookie(rest.IdentTypeSessionCookieName, sid, 0, "/", "", false, false)
		}
	}
}

func verifyUserAppAuthorization(env ReqEnv) bool {
	u := env.AuthInfo.User
	return u.HasRole4DomainSiteAppId(env.Domain, env.Site, env.App.Id)
}
