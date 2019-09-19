// SPDX-License-Identifier: MIT

// Package input 用于处理输入的文件，从代码中提取基本的注释内容。
//
// 多行注释和单行注释在处理上会有一定区别：
//  - 单行注释，风格相同且相邻的注释会被合并成一个注释块；
//  - 单行注释，风格不相同且相邻的注释会被按注释风格多个注释块；
//  - 多行注释，即使两个注释释块相邻也会被分成两个注释块来处理。
package input

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"sync"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/caixw/apidoc/v5/doc"
	"github.com/caixw/apidoc/v5/internal/lang"
	"github.com/caixw/apidoc/v5/internal/locale"
	"github.com/caixw/apidoc/v5/message"
)

// 解析出来的注释块
type block struct {
	File string
	Line int
	Data []byte
}

// Parse 分析从 input 中获取的代码块
//
// 所有与解析有关的错误均通过 h 输出。
// 如果 input 参数有误，会通过 error 参数返回。
func Parse(h *message.Handler, opt ...*Options) (*doc.Doc, error) {
	blocks, err := buildBlock(h, opt...)
	if err != nil {
		return nil, err
	}

	doc := doc.New()
	wg := sync.WaitGroup{}

	for blk := range blocks {
		wg.Add(1)
		go func(b block) {
			parseBlock(doc, b, h)
			wg.Done()
		}(blk)
	}

	wg.Wait()

	if err := doc.Sanitize(); err != nil {
		h.Error(message.Erro, err)
	}

	return doc, nil
}

var (
	apidocBegin = []byte("<apidoc")
	apiBegin    = []byte("<api")
)

func parseBlock(d *doc.Doc, block block, h *message.Handler) {
	var err error
	switch {
	case bytes.HasPrefix(block.Data, apidocBegin):
		err = d.FromXML(block.Data)
	case bytes.HasPrefix(block.Data, apiBegin):
		err = d.NewAPI(block.File, block.Line).FromXML(block.Data)
	}

	h.Error(message.Erro, message.WithError(block.File, "", block.Line, err))
}

// 分析源代码，获取注释块。
//
// 当所有的代码块已经放入 Block 之后，Block 会被关闭。
func buildBlock(h *message.Handler, opt ...*Options) (chan block, *message.SyntaxError) {
	if len(opt) == 0 {
		return nil, message.NewError("", "opt", 0, locale.ErrRequired)
	}

	for index, item := range opt {
		field := "opt[" + strconv.Itoa(index) + "]"
		if item == nil {
			return nil, message.NewError("", field, 0, locale.ErrRequired)
		}

		if err := item.Sanitize(); err != nil {
			err.Field = field + "." + err.Field
			return nil, err
		}
	}

	data := make(chan block, 500)

	go func() {
		wg := &sync.WaitGroup{}
		for _, o := range opt {
			parseOptions(data, h, wg, o)
		}
		wg.Wait()

		close(data)
	}()

	return data, nil
}

// 分析每个配置项对应的内容
func parseOptions(data chan block, h *message.Handler, wg *sync.WaitGroup, o *Options) {
	for _, path := range o.paths {
		wg.Add(1)
		go func(path string) {
			parseFile(data, h, path, o)
			wg.Done()
		}(path)
	}
}

// 分析 path 指向的文件。
//
// NOTE: parseFile 内部不能有协程处理代码。
func parseFile(channel chan block, h *message.Handler, path string, o *Options) {
	data, err := readFile(path, o.encoding)
	if err != nil {
		h.Error(message.Erro, message.WithError(path, "", 0, err))
		return
	}

	ret := lang.Parse(path, data, o.blocks, h)
	for line, data := range ret {
		channel <- block{
			File: path,
			Line: line,
			Data: data,
		}
	}
}

// 以指定的编码方式读取内容。
func readFile(path string, encoding encoding.Encoding) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if encoding == nil {
		return data, nil
	}

	reader := transform.NewReader(bytes.NewReader(data), encoding.NewDecoder())
	return ioutil.ReadAll(reader)
}
