package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"net/url"
)

func redirectTo(c *gin.Context, redirectTo string) {
	location := url.URL{Path: redirectTo}
	c.Redirect(http.StatusFound, location.RequestURI())
	c.Abort()
}

func systemError(c *gin.Context, err error) {
	c.String(http.StatusInternalServerError, err.Error())
	c.Abort()
}

func GetRequestEnvironmentFromContext(c *gin.Context) ReqEnv {
	const semLogContext = "middleware/get-request-environment-from-context"
	ctxEnv, _ := c.Get(rest.RequestEnvironmentGinContextVarName)

	var zEvt *zerolog.Event
	if ctxEnv == nil {
		ctxEnv = ReqEnv{}
		zEvt = log.Warn()
	} else {
		zEvt = log.Trace()
	}
	zEvt.Interface(rest.RequestEnvironmentGinContextVarName, ctxEnv).Msg(semLogContext)
	return ctxEnv.(ReqEnv)
}

func SetRequestEnvironmentInContext(c *gin.Context, env ReqEnv) {
	c.Set(rest.RequestEnvironmentGinContextVarName, env)
}

func AddAuthInfo2RequestEnvironment(c *gin.Context, ai AuthInfo) ReqEnv {
	env := GetRequestEnvironmentFromContext(c)
	env.AuthInfo = ai
	c.Set(rest.RequestEnvironmentGinContextVarName, env)
	return env
}
