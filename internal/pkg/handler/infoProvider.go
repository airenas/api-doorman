package handler

import "strings"

type infoProvider interface {
	Info(prefix string) string
}

//GetInfo return info if handler has one
func GetInfo(pr string, h interface{}) string {
	pr = pr + "  "
	if ip, ok := h.(infoProvider); ok {
		return addNewLine(pr + ip.Info(pr))
	}
	return addNewLine(pr + "Handler does not provide info")
}

func addNewLine(pr string) string {
	if !strings.HasSuffix(pr, "\n") {
		pr = pr + "\n"
	}
	return pr
}
