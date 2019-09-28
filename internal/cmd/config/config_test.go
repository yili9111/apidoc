// SPDX-License-Identifier: MIT

package config

import (
	"path/filepath"
	"testing"

	"github.com/issue9/assert"

	"github.com/caixw/apidoc/v5"
	"github.com/caixw/apidoc/v5/input"
)

func TestLoad(t *testing.T) {
	a := assert.New(t)

	cfg, err := Load("./not-exists-file")
	a.Error(err).Nil(cfg)

	cfg, err = Load("./failed.yaml")
	a.Error(err).Nil(cfg)
}

func TestWrite_Load(t *testing.T) {
	a := assert.New(t)

	wd, err := filepath.Abs("./")
	a.NotError(err).NotEmpty(wd)
	a.NotError(Write(wd))

	path := filepath.Join(wd, configFilename)
	cfg, err := Load(path)
	a.NotError(err).
		NotNil(cfg)

	a.Equal(cfg.Version, apidoc.Version()).
		Equal(cfg.Inputs[0].Lang, "go")
}

func TestConfig_sanitize(t *testing.T) {
	a := assert.New(t)

	// 错误的版本号格式
	conf := &Config{}
	err := conf.sanitize("./apidoc.yaml")
	a.Error(err).
		Equal(err.Field, "version")

	// 与当前程序的版本号不兼容
	conf.Version = "1.0"
	err = conf.sanitize("./apidoc.yaml")
	a.Error(err).
		Equal(err.Field, "version")

	// 未声明 inputs
	conf.Version = "5.0.1"
	err = conf.sanitize("./apidoc.yaml")
	a.Error(err).
		Equal(err.Field, "inputs")

	// 未声明 output
	conf.Inputs = []*input.Options{{}}
	err = conf.sanitize("./apidoc.yaml")
	a.Error(err).
		Equal(err.Field, "output")
}