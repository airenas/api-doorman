package test

import (
	"context"
	"log"
	"net"
	"net/url"
	"time"
)

func WaitForOpenOrFail(ctx context.Context, urlWait string) {
	u, err := url.Parse(urlWait)
	if err != nil {
		log.Fatalf("FAIL: can't parse %s", urlWait)
	}
	for {
		if err := listen(net.JoinHostPort(u.Hostname(), u.Port())); err != nil {
			log.Printf("waiting for %s ...", urlWait)
		} else {
			return
		}
		select {
		case <-ctx.Done():
			log.Fatalf("FAIL: can't access %s", urlWait)
			break
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func listen(urlStr string) error {
	log.Printf("dial %s", urlStr)
	conn, err := net.DialTimeout("tcp", urlStr, time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return err
}
