package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCoveragecfgFileForCompile(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		desc    string
		e       encoder
		pkgPath string
		pkgName string
		outDir  string
		expErr  string
	}{
		{
			desc: "expected behavior",
			e: func(b *bytes.Buffer, v any) error {
				enc := json.NewEncoder(b)
				return enc.Encode(v)
			},
			pkgName: "something",
			pkgPath: "github.com/Someone/something",
			outDir:  "/tmp/",
			expErr:  "",
		},
		{
			desc: "failed to encode",
			e: func(b *bytes.Buffer, v any) error {
				return errors.New("great sadness")
			},
			pkgName: "something",
			pkgPath: "github.com/Someone/something",
			outDir:  "/tmp/",
			expErr:  "great sadness",
		},
	}
	for _, tt := range testcases {
		t.Run(tt.desc, func(t *testing.T) {
			path, err := buildCoveragecfgFileForCompile(tt.e, tt.pkgPath, tt.pkgName, tt.outDir)
			if len(tt.expErr) > 0 {
				assert.Equal(t, len(path), 0)
				assert.ErrorContains(t, err, tt.expErr)
				return
			}
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			assert.Contains(t, string(content), fmt.Sprintf(`"PkgPath":"%s"`, tt.pkgPath))
			assert.Contains(t, string(content), fmt.Sprintf(`"PkgName":"%s"`, tt.pkgName))
			assert.Contains(t, string(content), fmt.Sprintf(`"OutConfig":"%s"`, filepath.Join(tt.outDir, "coveragecfg")))
		})
	}
}

func TestBuildOutFileList(t *testing.T) {
	t.Parallel()
	t.Run("first element is covervars", func(t *testing.T) {
		path, err := buildOutFileList([]string{}, "/tmp")
		if err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		filenames := strings.Split(string(content), "\n")
		assert.Equal(t, "/tmp/covervars.go", filenames[0])
	})
	t.Run("output file contain cover files routes", func(t *testing.T) {
		files := []string{"main.go", "foo.go"}
		expFiles := []string{"/tmp/main.cover.go", "/tmp/foo.cover.go"}
		path, err := buildOutFileList(files, "/tmp")
		if err != nil {
			t.Fatal(err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		filenames := strings.Split(string(content), "\n")
		filenames = filenames[1:]
		for i, fileroute := range filenames {
			assert.Equal(t, fileroute, expFiles[i])
		}
	})
}
