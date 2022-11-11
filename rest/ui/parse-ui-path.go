package ui

import (
	"github.com/mario-imperato/r3ng-apigtw/constants"
	"github.com/rs/zerolog/log"
	"regexp"
	"strings"
)

type PathType int

const (
	ParsePathZero PathType = iota
	ParsePathLang
	ParsePathDomain
	ParsePathDomainLang
	ParsePathDomainSite
	ParsePathDomainSiteLang
	ParsePathDomainSiteLangAppName
)

type UiPathInfo struct {
	Console      bool
	PathType     PathType
	PathSegments []string
}

func (p UiPathInfo) Domain() string {
	if p.PathType == ParsePathZero || p.PathType == ParsePathLang {
		return ""
	}

	return p.PathSegments[0]
}

func (p UiPathInfo) Site() string {
	if p.PathType == ParsePathDomainSite || p.PathType == ParsePathDomainSiteLang || p.PathType == ParsePathDomainSiteLangAppName {
		return p.PathSegments[1]
	}

	return ""
}

func (p UiPathInfo) Lang() string {

	l := ""
	switch p.PathType {
	case ParsePathLang:
		l = p.PathSegments[0]
	case ParsePathDomainLang:
		l = p.PathSegments[1]
	case ParsePathDomainSiteLang:
		fallthrough
	case ParsePathDomainSiteLangAppName:
		l = p.PathSegments[2]
	}

	return l
}

func (p UiPathInfo) AppName() string {
	if p.PathType == ParsePathDomainSiteLangAppName {
		return p.PathSegments[3]
	}

	return ""
}

func (p UiPathInfo) AppPath() string {
	return p.PathSegments[len(p.PathSegments)-1]
}

func (p UiPathInfo) ShouldRedirect() (string, bool) {

	redirFlag := false

	var sb strings.Builder
	if p.Console {
		sb.WriteString("/ui-console")
	} else {
		sb.WriteString("/ui")
	}

	switch p.PathType {
	case ParsePathZero:
		sb.WriteString("/it")
		sb.WriteString(p.PathSegments[0])
		redirFlag = true
	case ParsePathLang:
	case ParsePathDomain:
		sb.WriteString("/")
		sb.WriteString(p.PathSegments[0])
		sb.WriteString("/it")
		sb.WriteString(p.PathSegments[1])
		redirFlag = true
	case ParsePathDomainLang:
	case ParsePathDomainSite:
		sb.WriteString("/")
		sb.WriteString(p.PathSegments[0])
		sb.WriteString("/")
		sb.WriteString(p.PathSegments[1])
		sb.WriteString("/it")
		sb.WriteString(p.PathSegments[2])
		redirFlag = true
	case ParsePathDomainSiteLang:
		sb.WriteString("/")
		sb.WriteString(p.PathSegments[0])
		sb.WriteString("/")
		sb.WriteString(p.PathSegments[1])
		sb.WriteString("/")
		sb.WriteString(p.PathSegments[2])
		sb.WriteString("/")
		sb.WriteString(constants.AppHome)
		sb.WriteString(p.PathSegments[3])
		redirFlag = true
	case ParsePathDomainSiteLangAppName:
	}

	if redirFlag {
		return sb.String(), redirFlag
	}

	return "", false
}

func ParseUiPathInfo(console bool, p string) UiPathInfo {

	tp := strings.TrimPrefix(p, "/")
	tp = strings.TrimSuffix(tp, "/")

	segments := strings.Split(tp, "/")
	log.Info().Int("segments", len(segments)).Str("path-info", tp).Msg("parse path-info")

	pi := UiPathInfo{PathType: ParsePathZero, Console: console, PathSegments: make([]string, 0)}

	for _, s := range segments {

		var isPartial, ok, isFinal bool
		var ident string
		if s == "" {
			break
		}

		switch pi.PathType {
		case ParsePathZero:
			ident, isPartial, ok = acceptIdentifier(s)
			if ok {
				pi.PathSegments = append(pi.PathSegments, ident)
				if acceptLanguage(ident) {
					pi.PathType = ParsePathLang
					isFinal = true
				} else {
					pi.PathType = ParsePathDomain
				}
			}

		case ParsePathDomain:
			ident, isPartial, ok = acceptIdentifier(s)
			if ok {
				pi.PathSegments = append(pi.PathSegments, ident)
				if acceptLanguage(ident) {
					pi.PathType = ParsePathDomainLang
					isFinal = true
				} else {
					pi.PathType = ParsePathDomainSite
				}
			}

		case ParsePathDomainSite:
			ident, isPartial, ok = acceptIdentifier(s)
			if ok {
				if acceptLanguage(ident) {
					pi.PathSegments = append(pi.PathSegments, ident)
					pi.PathType = ParsePathDomainSiteLang
				} else {
					isFinal = true
				}
			}
		case ParsePathDomainSiteLang:
			ident, isPartial, ok = acceptIdentifier(s)
			if ok {
				pi.PathSegments = append(pi.PathSegments, ident)
				pi.PathType = ParsePathDomainSiteLangAppName
				isFinal = true
			}
		}

		if isPartial || !ok || isFinal {
			break
		}
	}

	pfix := "/" + strings.Join(pi.PathSegments, "/")
	if pi.PathType == ParsePathZero {
		pi.PathSegments = append(pi.PathSegments, p)
	} else {
		pi.PathSegments = append(pi.PathSegments, strings.TrimPrefix(p, pfix))
	}

	return pi
}

func acceptLanguage(n string) bool {
	return n == "it"
}

func acceptIdentifier(n string) (string, bool, bool) {

	nn := n
	if ndx := strings.Index(n, "?"); ndx > 0 {
		n = n[0:ndx]
	}

	if ndx := strings.Index(n, "#"); ndx > 0 {
		n = n[0:ndx]
	}

	return n, nn != n, isIdentifier(n)
}

var IdentifierStringRegexp = regexp.MustCompile("^[a-zA-Z0-9_-]*$")

func isIdentifier(inputData string) bool {
	return IdentifierStringRegexp.Match([]byte(inputData))
}
