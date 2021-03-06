// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/caixw/apidoc/v6/internal/locale"
)

func initLocale() {
	command.New("locale", doLocale, buildUsage(locale.CmdLocaleUsage))
}

func doLocale(w io.Writer) error {
	locales := make(map[string]string, len(locale.DisplayNames()))

	// 计算各列的最大长度值
	var maxID int
	for k, v := range locale.DisplayNames() {
		id := k.String()
		calcMaxWidth(id, &maxID)
		locales[id] = v
	}
	maxID += tail

	for k, v := range locales {
		id := k + strings.Repeat(" ", maxID-len(k))
		if _, err := fmt.Fprintln(w, id, v); err != nil {
			return err
		}
	}

	return nil
}
