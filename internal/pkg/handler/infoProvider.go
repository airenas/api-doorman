package handler

import "strings"

type infoProvider interface {
	Info(prefix string) string
}

//GetInfo return info if handler has one
func GetInfo(pr string, h interface{}) string {
	if ip, ok := h.(infoProvider); ok {
		return checkNewLine(ip.Info(pr))
	}
	return checkNewLine(pr + "Handler does not provide info")
}

//LogShitf appends spaces to strin
func LogShitf(str string) string {
	return str + "  "
}

func checkNewLine(pr string) string {
	if !strings.HasSuffix(pr, "\n") {
		pr = pr + "\n"
	}
	return pr
}
