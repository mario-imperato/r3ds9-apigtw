package middleware

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-gin/httpsrv"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ng-apigtw/constants"
	"net/http"
	"net/url"
)

func RequestEnvResolver(mountpPoint string) httpsrv.H {
	return func(c *gin.Context) {
		env, ok := extractEnvFromContext(c, mountpPoint)
		if !ok {
			location := url.URL{Path: constants.DefaultRouteToRoot}
			c.Redirect(http.StatusFound, location.RequestURI())
			c.Abort()
			return
		}

		c.Set("env", env)
	}
}
