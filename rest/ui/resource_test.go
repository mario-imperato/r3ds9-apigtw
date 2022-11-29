package ui_test

import (
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestGlobPattern(t *testing.T) {
	//p := "/Users/marioa.imperato/projects/r3ds9-2/r3ds9-apps/ng-hello-world/dist/ng-hello-world/index.tmpl"
	p := "/Users/marioa.imperato/projects/r3ds9-2/r3ds9-apps/*/dist/*/*tmpl"

	files, err := filepath.Glob(p)
	require.NoError(t, err)

	for _, f := range files {
		t.Log(f)
	}
}
