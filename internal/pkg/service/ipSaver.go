package service

import "github.com/airenas/api-doorman/internal/pkg/handler"

type ipSaver struct {
	saver IPManager
	limit float64
}

func (is *ipSaver) Save(ip string) error {
	return is.saver.CheckCreate(ip, is.limit)
}

func newIPSaver(data *Data) handler.IPSaver {
	res := &ipSaver{}
	res.saver = data.IPSaver
	res.limit = data.Proxy.DefaultLimit
	return res
}
