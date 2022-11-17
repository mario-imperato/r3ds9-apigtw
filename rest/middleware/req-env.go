package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mario-imperato/r3ds9-apigtw/rest"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw/commons"
	"github.com/mario-imperato/r3ds9-mongodb/model/r3ds9-apigtw/user"
	"net/http"
	"net/url"
)

type AuthInfo struct {
	Sid        string    `json:"sid,omitempty" yaml:"sid,omitempty"`
	User       user.User `json:"user,omitempty" yaml:"user,omitempty"`
	Remoteaddr string    `json:"remoteaddr,omitempty" yaml:"remoteaddr,omitempty"`
	Flags      string    `json:"flags,omitempty" yaml:"flags,omitempty"`
	NewSid     bool      `json:"-" yaml:"-"`
}

type ReqEnv struct {
	Category      rest.ReqCategory
	ReqType       rest.ReqType
	Domain        string
	Site          string
	Lang          string
	App           commons.App
	ExtraPathInfo string
	AuthInfo      AuthInfo
}

func (re ReqEnv) IsMalformed() bool {
	return re.ReqType == rest.ReqTypeMalformed
}

func (re ReqEnv) IsInvalid() bool {
	return re.ReqType == rest.ReqTypeInvalid
}

func (re ReqEnv) IsValid() bool {
	return re.ReqType != rest.ReqTypeInvalid && re.ReqType != rest.ReqTypeMalformed
}

func (re ReqEnv) IsUi() bool {
	return re.Category != rest.ReqTypeCategoryApi
}

func (re ReqEnv) IsApi() bool {
	return re.Category == rest.ReqTypeCategoryApi
}

func (re ReqEnv) String() string {
	return fmt.Sprintf("[%s] category: %s, domain: %s, site: %s, lang: %s, appId: %s, extra-path: %s", re.ReqType, re.Category, re.Domain, re.Site, re.Lang, re.App.Id, re.ExtraPathInfo)
}

func (re ReqEnv) Path() string {
	return RequestPath4(re.Category, re.Domain, re.Site, re.Lang, re.App.Id, re.ExtraPathInfo)
}

func (re ReqEnv) AbortRequest(c *gin.Context, redirectUrl string) {
	if re.Category != rest.ReqTypeCategoryApi {
		location := url.URL{Path: redirectUrl}
		c.Redirect(http.StatusFound, location.RequestURI())
	} else {
		c.String(http.StatusBadRequest, "Bad Request")
	}
	c.Abort()
}

func RequestPath4(category rest.ReqCategory, domain, site, lang, appId, extraPath string) string {
	return fmt.Sprintf("/%s/%s/%s/%s/%s%s", category, domain, site, lang, appId, extraPath)
}
