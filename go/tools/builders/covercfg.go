package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type coverPkgConfig struct {
	OutConfig   string `json:"OutConfig"`
	PkgPath     string `json:"PkgPath"`
	PkgName     string `json:"PkgName"`
	Granularity string `json:"Granularity"`
	Local       bool   `json:"Local"`
}

type encoder func(*bytes.Buffer, any) error

// buildCoveragecfgFileForCompile generates a package configuration file that
// is given as value to the "pkgcfg" flag of "go tool cover".
func buildCoveragecfgFileForCompile(e encoder, pkgPath, pkgName, outDir string) (string, error) {
	buf := &bytes.Buffer{}
	outConfig := filepath.Join(outDir, "coveragecfg")
	covConfig := coverPkgConfig{
		OutConfig:   outConfig,
		PkgPath:     pkgPath,
		PkgName:     pkgName,
		Granularity: "perblock",
		Local:       false,
	}
	if err := e(buf, covConfig); err != nil {
		return "", err
	}

	f, err := ioutil.TempFile(outDir, "pkgcfg")
	if err != nil {
		return "", err
	}

	filename := f.Name()

	if _, err := io.Copy(f, buf); err != nil {
		f.Close()
		os.Remove(filename)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(filename)
		return "", err
	}

	return filename, nil
}

// buildOutFileList creates a file with names for all the instrumented code of
// a package, this file will be given as value of the "outfilelist" from
// "go tool cover"
func buildOutFileList(srcs []string, outDir string) (string, error) {
	buf := &bytes.Buffer{}
	// Starting in go1.21, go tool cover dumps the cover variables in a
	// separate file
	coverVars := filepath.Join(outDir, "covervars.go")
	fmt.Fprintf(buf, "%s\n", coverVars)
	for i, filename := range srcs {
		base := filepath.Base(filepath.Base(filename))
		name := strings.TrimSuffix(base, filepath.Ext(base))
		coverSrc := filepath.Join(outDir, fmt.Sprintf("%s.cover.go", filepath.Base(name)))
		if i+1 < len(srcs) {
			fmt.Fprintf(buf, "%s\n", coverSrc)
		} else {
			fmt.Fprintf(buf, "%s", coverSrc)
		}
	}

	f, err := ioutil.TempFile(outDir, "outfilelist")
	if err != nil {
		return "", err
	}

	filename := f.Name()

	if _, err := io.Copy(f, buf); err != nil {
		f.Close()
		os.Remove(filename)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(filename)
		return "", err
	}
	return filename, nil
}
