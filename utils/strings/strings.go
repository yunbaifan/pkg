package stringutil

import "fmt"

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
