package tag

import (
	"fmt"
	"strings"
)

func Parse(tag string) (string /*key*/, string /*value*/, error) {
	if idx := strings.IndexByte(tag, ':'); idx >= 0 {
		return strings.ToLower(strings.TrimSpace(tag[:idx])), strings.TrimSpace(tag[idx+1:]), nil
	}
	if (strings.TrimSpace(tag)) == "" {
		return "", "", nil
	}
	return "", "", fmt.Errorf("wrong tag value, no ':' in '%s'", tag)
}
