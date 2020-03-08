// SPDX-License-Identifier: MIT

// Package spec 对文档规则的定义
package spec

import (
	xmessage "golang.org/x/text/message"

	"github.com/caixw/apidoc/v6/message"
)

const (
	// Version 文档规范的版本
	Version = "6.0.0"

	// MajorVersion 文档规范的主版本信息
	MajorVersion = "v6"
)

// Block 表示原始的注释代码块
type Block struct {
	File string
	Line int
	Data []byte // 整理之后的数据
	Raw  []byte // 原始数据
}

func (b *Block) localeError(field string, key xmessage.Reference, v ...interface{}) error {
	return message.NewLocaleError(b.File, field, b.Line, key, v...)
}