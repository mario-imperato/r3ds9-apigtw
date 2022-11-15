package middleware

import (
	"context"
	"errors"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/linkedservices"
	"github.com/mario-imperato/r3ds9-apigtw/linkedservices/mongodb"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-core/session"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-core/user"
	"github.com/rs/zerolog/log"
	"net/http"
)

func RequestUserResolver() httpsrv.H {
	const semLogContext = "middleware/request-user-resolver"
	return func(c *gin.Context) {

		lks, err := linkedservices.GetMongoDbService(context.Background(), "r3ds9")
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			systemError(c, err)
			return
		}

		var a AuthInfo
		identType, sid := resolveIdentificationType(c)
		switch identType {
		case rest.IdentTypeUnknown:
			a, err = newGuestSession(c, lks)
			if err != nil {
				log.Error().Err(err).Msg(semLogContext)
				systemError(c, err)
			}
		case rest.IdentTypeSessionCookie:
			a, err = retrieveUserSession(c, lks, sid)
			if err != nil {
				a, err = newGuestSession(c, lks)
				if err != nil {
					log.Error().Err(err).Msg(semLogContext)
					systemError(c, err)
				}
			}
		}

		AddAuthInfo2RequestEnvironment(c, a)
	}
}

func resolveIdentificationType(c *gin.Context) (rest.IdentType, string) {
	const semLogContext = "middleware/request-user-resolver/resolve-identification-type"
	cookie, err := c.Request.Cookie(rest.IdentTypeSessionCookieName)
	if err != nil && err != http.ErrNoCookie {
		log.Error().Err(err).Interface("identification-type", rest.IdentTypeUnknown).Msg(semLogContext)
		return rest.IdentTypeUnknown, ""
	}

	var tok string
	it := rest.IdentTypeUnknown
	if err == nil {
		it = rest.IdentTypeSessionCookie
		tok = cookie.Value
	}
	log.Trace().Interface("identification-type", it).Interface("cookie", cookie).Msg(semLogContext)
	return it, tok
}

func newGuestSession(c *gin.Context, lks *mongodb.MDbLinkedService) (AuthInfo, error) {
	const semLogContext = "middleware/request-user-resolver/new-guest-session"
	u, ok := user.GetFromCache(user.NewCacheResolver(lks.GetCollection("user", "")), "guest")
	if !ok {
		err := errors.New("user not found")
		log.Error().Err(err).Msg(semLogContext)
		return AuthInfo{}, err
	}

	return AuthInfo{
		User:       *u,
		Remoteaddr: c.Request.RemoteAddr,
		Flags:      "",
		NewSid:     true,
	}, nil
}

func retrieveUserSession(c *gin.Context, lks *mongodb.MDbLinkedService, sid string) (AuthInfo, error) {
	const semLogContext = "middleware/request-user-resolver/retrieve-user-session"
	s, ok := session.GetFromCache(session.NewCacheResolver(lks.GetCollection("session", "")), sid)
	if !ok {
		err := errors.New("session not found")
		log.Error().Err(err).Msg(semLogContext)
		return AuthInfo{}, err
	}

	u, ok := user.GetFromCache(user.NewCacheResolver(lks.GetCollection("user", "")), s.Nickname)
	if !ok {
		err := errors.New("session user not found")
		log.Error().Err(err).Msg(semLogContext)
		return AuthInfo{}, err
	}

	return AuthInfo{
		User:       *u,
		Remoteaddr: c.Request.RemoteAddr,
		Flags:      "",
		NewSid:     false,
	}, nil
}
