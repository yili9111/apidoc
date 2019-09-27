// SPDX-License-Identifier: MIT

package doc

import (
	"encoding/xml"

	"github.com/caixw/apidoc/v5/internal/locale"
)

// Request 请求内容
type Request struct {
	Type        Type       `xml:"type,attr"`
	Deprecated  Version    `xml:"deprecated,attr,omitempty"`
	Enums       []*Enum    `xml:"enum,omitempty"`
	Array       bool       `xml:"array,attr,omitempty"`
	Items       []*Param   `xml:"param,omitempty"`
	Reference   string     `xml:"ref,attr,omitempty"`
	Summary     string     `xml:"summary,attr,omitempty"`
	Status      Status     `xml:"status,attr,omitempty"`
	Mimetype    string     `xml:"mimetype,attr"`
	Examples    []*Example `xml:"example,omitempty"`
	Headers     []*Header  `xml:"header,omitempty"`
	Description string     `xml:"description,omitempty"`
}

// IsEnum 是否为枚举值
func (r *Request) IsEnum() bool {
	return len(r.Enums) > 0
}

type shadowRequest Request

// UnmarshalXML xml.Unmarshaler
func (r *Request) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	field := "/" + start.Name.Local
	var shadow shadowRequest
	if err := d.DecodeElement(&shadow, &start); err != nil {
		return fixedSyntaxError(err, "", field, 0)
	}

	if shadow.Type == None {
		return newSyntaxError(field+"#type", locale.ErrRequired)
	}
	if shadow.Type == Object && len(shadow.Items) == 0 {
		return newSyntaxError(field+"/item", locale.ErrRequired)
	}

	if shadow.Mimetype == "" {
		return newSyntaxError(field+"#mimetype", locale.ErrRequired)
	}

	// 判断 enums 的值是否相同
	if key := getDuplicateEnum(shadow.Enums); key != "" {
		return newSyntaxError(field+"/enum", locale.ErrDuplicateValue)
	}

	// 判断 items 的值是否相同
	if key := getDuplicateItems(shadow.Items); key != "" {
		return newSyntaxError(field+"/item", locale.ErrDuplicateValue)
	}

	*r = Request(shadow)
	return nil
}
