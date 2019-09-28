// SPDX-License-Identifier: MIT

// apidoc 是一个 RESTful API 文档生成工具
//
// 大致的使用方法为：
//  apidoc [options] [path]
// 具体的参数说明，可以使用 h 参数查看：
//  apidoc -h
// path 表示的路径列表，多个使用空格分隔。
// 如果带了参数 d，则表示的需要检测的目录列表，
// 否则为 apidoc 的配置文件地址列表。为空表示当前目录。
package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/issue9/term/colors"
	"golang.org/x/text/language"

	"github.com/caixw/apidoc/v5"
	"github.com/caixw/apidoc/v5/input"
	"github.com/caixw/apidoc/v5/internal/cmd/config"
	"github.com/caixw/apidoc/v5/internal/cmd/term"
	"github.com/caixw/apidoc/v5/internal/locale"
	"github.com/caixw/apidoc/v5/internal/vars"
	"github.com/caixw/apidoc/v5/message"
)

const (
	succColor = colors.Green
	infoColor = colors.Default
	warnColor = colors.Cyan
	erroColor = colors.Red
)

var (
	succOut = os.Stdout
	infoOut = os.Stdout
	warnOut = os.Stderr
	erroOut = os.Stderr
)

func init() {
	if err := apidoc.Init(language.Und); err != nil {
		term.Line(warnOut, warnColor, err)
	}
}

func main() {
	h := flag.Bool("h", false, locale.Sprintf(locale.FlagHUsage))
	v := flag.Bool("v", false, locale.Sprintf(locale.FlagVUsage))
	d := flag.Bool("d", false, locale.Sprintf(locale.FlagDUsage))
	t := flag.Bool("t", false, locale.Sprintf(locale.FlagTUsage))
	l := flag.Bool("l", false, locale.Sprintf(locale.FlagLUsage))
	flag.Usage = usage
	flag.Parse()

	switch {
	case *h:
		flag.Usage()
		return
	case *v:
		term.Locale(infoOut, infoColor, locale.FlagVersionBuildWith, vars.Name, apidoc.Version(), runtime.Version())
		term.Locale(infoOut, infoColor, locale.FlagVersionCommitHash, vars.CommitHash())
		return
	case *t:
		parse(true)
		return
	case *l:
		for _, l := range term.Langs(3) {
			term.Line(infoOut, infoColor, l)
		}
		return
	case *d:
		detect()
		return
	}

	parse(false)
}

func detect() {
	paths, err := getPaths()
	if err != nil {
		term.Line(erroOut, erroColor, err)
		return
	}

	for _, dir := range paths {
		if err := config.Write(dir); err != nil {
			term.Line(erroOut, erroColor, err)
			return
		}
		term.Locale(succOut, succColor, locale.ConfigWriteSuccess, dir)
	}
}

// 参数 test 表示是否只作语法检测，不输出内容。
func parse(test bool) {
	h := message.NewHandler(term.NewHandlerFunc(erroOut, warnOut, infoOut, succOut,
		erroColor, warnColor, infoColor, succColor))

	paths, err := getPaths()
	if err != nil {
		h.Error(message.Erro, err)
		return
	}
	for _, path := range paths {
		now := time.Now()

		cfg, err := config.Load(path)
		if err != nil {
			h.Error(message.Erro, err)
			continue
		}

		if test {
			if err := apidoc.Do(h, cfg.Output, cfg.Inputs...); err != nil {
				h.Error(message.Erro, err)
				continue
			}
			h.Message(message.Succ, locale.Complete, cfg.Output.Path, time.Now().Sub(now))
		} else {
			if _, err := input.Parse(h, cfg.Inputs...); err != nil {
				h.Error(message.Erro, err)
				continue
			}
			h.Message(message.Succ, locale.TestSuccess)
		}
	} // end for paths

	h.Stop()
}

func usage() {
	buf := new(bytes.Buffer)
	flag.CommandLine.SetOutput(buf)
	flag.PrintDefaults()

	term.Locale(infoOut, infoColor, locale.FlagUsage, vars.Name, buf.String(), vars.RepoURL, vars.OfficialURL)
}

func getPaths() ([]string, error) {
	paths := flag.Args()
	if len(paths) == 0 {
		paths = append(paths, "./")
	}

	for index, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}

		paths[index] = path
	}

	return paths, nil
}
