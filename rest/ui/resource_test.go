package ui_test

import (
	"fmt"
	"github.com/mario-imperato/r3ng-apigtw/rest/ui"
	"github.com/stretchr/testify/require"
	"testing"
)

type inputWanted struct {
	input          string
	wantedPathType ui.PathType
	wantedPathInfo string
	wantedRedirect bool
}

func TestParseUrl(t *testing.T) {

	arr := []inputWanted{
		{"", ui.ParsePathZero, "", true},
		{"/", ui.ParsePathZero, "/", true},
		{"/cvf", ui.ParsePathDomain, "", true},
		{"/cvf/", ui.ParsePathDomain, "/", true},
		{"/cvf/champ42", ui.ParsePathDomainSite, "", true},
		{"/cvf/champ42/", ui.ParsePathDomainSite, "/", true},
		{"/cvf/champ42/it", ui.ParsePathDomainSiteLang, "", true},
		{"/cvf/champ42/it/", ui.ParsePathDomainSiteLang, "/", true},
		{"/cvf/champ42/it/?param=23", ui.ParsePathDomainSiteLang, "/?param=23", true},
		{"/cvf/champ42/it/cms/", ui.ParsePathDomainSiteLangAppName, "/", false},
		{"/cvf/champ42/it/cms", ui.ParsePathDomainSiteLangAppName, "", false},
		{"/cvf/champ42/it/cms?param=23", ui.ParsePathDomainSiteLangAppName, "?param=23", false},
		{"/cvf/champ42/it/cms/?param=23", ui.ParsePathDomainSiteLangAppName, "/?param=23", false},
	}

	for i := range arr {
		pathInfo := ui.ParseUiPathInfo(false, arr[i].input)
		require.Equal(t, arr[i].wantedPathType, pathInfo.PathType, fmt.Sprintf("got %s but expected %d for input %s", pathInfo.PathType.String(), arr[i].wantedPathType, arr[i].input))
		require.Equal(t, arr[i].wantedPathInfo, pathInfo.AppPath(), fmt.Sprintf("got %s but expected %s for input %s", pathInfo.AppPath(), arr[i].wantedPathInfo, arr[i].input))

		redirUrl, shouldRedirect := pathInfo.ShouldRedirect()
		require.Equal(t, arr[i].wantedRedirect, shouldRedirect, fmt.Sprintf("got unexpected redirection to %s for input %s", redirUrl, arr[i].input))
	}

}
