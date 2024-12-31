package service

import (
	"context"

	"github.com/airenas/api-doorman/internal/pkg/handler"
)

type ipSaver struct {
	saver IPManager
	limit float64
}

func (is *ipSaver) Save(ctx context.Context, ip string) (string, error) {
	return is.saver.CheckCreateIPKey(ctx, ip, is.limit)
}

func newIPSaver(saver IPManager, limit float64) handler.IPSaver {
	res := &ipSaver{}
	res.saver = saver
	res.limit = limit
	return res
}
