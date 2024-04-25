package stringutil

import (
	"bytes"
	"fmt"
)

func FormatString(text string, length int, alignLeft bool) string {
	if len(text) > length {
		return text[:length]
	}
	var formatStr string
	if alignLeft {
		formatStr = fmt.Sprintf("%%-%ds", length)
	} else {
		formatStr = fmt.Sprintf("%%%ds", length)
	}
	return fmt.Sprintf(formatStr, text)
}

func ToString(s string, kv []any) string {
	var buf bytes.Buffer
	switch {
	case len(kv) == 0:
	default:
		buf.WriteString(s)
		for i := 0; i < len(kv); i += 2 {
			if buf.Len() > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%v", kv[i]))
			buf.WriteString("=")
			if j := i + 1; j < len(kv) {
				buf.WriteString(fmt.Sprintf("%v", kv[j]))
			} else {
				buf.WriteString("Missing value")
			}
		}
	}
	return buf.String()
}
