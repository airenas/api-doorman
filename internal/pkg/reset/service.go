package reset

import (
	"context"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
)

// Cleaner interface for one Clean job
type Reseter interface {
	Reset(ctx context.Context, project string, since time.Time, limit float64) error
}

// TimerData keeps clean timer info
type TimerData struct {
	Reseter  Reseter
	Projects map[string]float64
}

// StartTimer starts timer in loop for doing reset quota tasks
func StartTimer(ctx context.Context, data *TimerData) (<-chan struct{}, error) {
	if data.Reseter == nil {
		return nil, errors.Errorf("no Reseter")
	}
	return startLoop(ctx, data), nil
}

func startLoop(ctx context.Context, data *TimerData) <-chan struct{} {
	goapp.Log.Infof("Starting reset timer")
	res := make(chan struct{}, 2)
	go func() {
		defer close(res)
		serviceLoop(ctx, data)
	}()
	return res
}

func serviceLoop(ctx context.Context, data *TimerData) {
	// run on startup
	now := time.Now()
	nextRun, err := utils.StartOfMonth(now, 1), doReset(ctx, now, data)
	if err != nil {
		goapp.Log.Error(err)
		nextRun = now.Add(time.Hour)
	}

	for {
		goapp.Log.Infof("next reset run at %s", nextRun.Format(time.RFC3339))
		select {
		case <-time.After(time.Until(nextRun)):
			now := time.Now()
			nextRun = utils.StartOfMonth(now, 1)
			err = doReset(ctx, now, data)
			if err != nil {
				goapp.Log.Error(err)
				nextRun = now.Add(time.Hour)
			}
		case <-ctx.Done():
			goapp.Log.Info("Stopped reset service")
			return
		}
	}
}

func doReset(ctx context.Context, now time.Time, data *TimerData) error {
	goapp.Log.Info("Running reset")
	for pr, value := range data.Projects {
		if err := data.Reseter.Reset(ctx, pr, now, value); err != nil {
			return err
		}
	}
	return nil
}
