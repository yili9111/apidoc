// SPDX-License-Identifier: MIT

// Package path 提供一些文件相关的操作
package path

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/caixw/apidoc/v6/internal/locale"
)

// Abs 获取 path 的绝对路径
//
// 如果 path 是相对路径的，则将其设置为相对于 wd 的路径
func Abs(path, wd string) (p string, err error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	if !isBeginHome(path) {
		path = filepath.Join(wd, path)
	}

	// ~ 路开头的相对路径，需要将其定位到 HOME 目录之下
	if isBeginHome(path) {
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(dir, path[2:])
	}

	if !filepath.IsAbs(path) {
		if path, err = filepath.Abs(path); err != nil {
			return "", err
		}
	}

	return filepath.Clean(path), nil
}

func isBeginHome(path string) bool {
	return strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\")
}

// Rel 尽可能地返回 path 相对于 wd 的路径，如果不存在相对关系，则原因返回 path。
func Rel(path, wd string) string {
	p, err := filepath.Rel(wd, path)
	if err != nil { // 不能转换不算错误，直接返回原值
		return path
	}
	return p
}

// CurrPath 获取相当于调用者所在目录的路径列表，相当于 PHP 的 __DIR__ + "/" + path
func CurrPath(path string) string {
	_, fi, _, _ := runtime.Caller(1)
	return filepath.Join(filepath.Dir(fi), path)
}

// ReadFile 读取本地或是远程的文件内容
//
// 根据 path 是否以 http:// 和 https:// 开头判断是否为远程文件
func ReadFile(path string, enc encoding.Encoding) ([]byte, error) {
	if !strings.HasPrefix(path, "https://") && !strings.HasPrefix(path, "http://") {
		return readLocalFile(path, enc)
	}
	return readRemoteFile(path, enc)
}

// 以指定的编码方式读取本地文件内容
func readLocalFile(path string, enc encoding.Encoding) ([]byte, error) {
	if enc == nil || enc == encoding.Nop {
		return ioutil.ReadFile(path)
	}

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	reader := transform.NewReader(r, enc.NewDecoder())
	return ioutil.ReadAll(reader)
}

// 以指定的编码方式读取远程文件内容
func readRemoteFile(url string, enc encoding.Encoding) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 300 {
		return nil, locale.Errorf(locale.ErrReadRemoteFile, url, resp.StatusCode)
	}

	if enc == nil || enc == encoding.Nop {
		return ioutil.ReadAll(resp.Body)
	}
	reader := transform.NewReader(resp.Body, enc.NewDecoder())
	return ioutil.ReadAll(reader)
}
