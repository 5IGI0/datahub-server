package main

import (
	"bytes"
	"strconv"
	"unicode/utf16"
)

type UnicodeEscape string

// complete version of:
// https://github.com/golang/go/issues/39137#issuecomment-630987977
// https://go.dev/play/p/YVSQzad2Z2r
func (ue UnicodeEscape) MarshalJSON() ([]byte, error) {
	var result = bytes.NewBufferString(`"`)
	for _, r := range ue {
		if r == '\\' || r == '"' {
			result.WriteByte('\\')
			result.WriteRune(r)
			continue
		}
		if r <= '~' && r >= ' ' {
			result.WriteRune(r)
			continue
		}
		if r < 0x10000 {
			result.WriteString("\\u")
			tmp := strconv.FormatInt(int64(r), 16)
			result.WriteString("0000"[:4-len(tmp)])
			result.WriteString(tmp)
			continue
		}

		r1, r2 := utf16.EncodeRune(r)
		result.WriteString("\\u")
		tmp := strconv.FormatInt(int64(r1), 16)
		result.WriteString("0000"[:4-len(tmp)])
		result.WriteString(tmp)

		result.WriteString("\\u")
		tmp = strconv.FormatInt(int64(r2), 16)
		result.WriteString("0000"[:4-len(tmp)])
		result.WriteString(tmp)
	}

	result.WriteByte('"')
	return result.Bytes(), nil
}
