package mongodb

import (
	"context"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// LogSaver saves log info to mongo db
type LogSaver struct {
	SessionProvider *SessionProvider
}

//NewLogSaver creates LogSaver instance
func NewLogSaver(sessionProvider *SessionProvider) (*LogSaver, error) {
	f := LogSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save log key to DB
func (ss *LogSaver) Save(log *adminapi.Log) error {
	goapp.Log.Infof("Saving log - %s, ip: %s, response: %d", log.URL, log.IP, log.ResponseCode)
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(logTable)

	_, err = c.InsertOne(ctx, mapFromLog(log))
	return err
}

// Get return all logs for key
func (ss *LogSaver) Get(key string) ([]*adminapi.Log, error) {
	goapp.Log.Infof("getting log list")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(logTable)
	cursor, err := c.Find(ctx, bson.M{"key": sanitize(key)})
	if err != nil {
		return nil, errors.Wrap(err, "Can't get logs")
	}
	defer cursor.Close(ctx)
	res := make([]*adminapi.Log, 0)
	for cursor.Next(ctx) {
		var logR logRecord
		if err = cursor.Decode(&logR); err != nil {
			return nil, errors.Wrap(err, "Can't get log record")
		}
		res = append(res, mapToLog(&logR))
	}
	return res, nil
}

func mapFromLog(v *adminapi.Log) *logRecord {
	res := &logRecord{}
	res.Key = v.Key
	res.Date = v.Date
	res.Fail = v.Fail
	res.IP = v.IP
	res.URL = v.URL
	res.QuotaValue = v.QuotaValue
	res.Value = v.Value
	res.ResponseCode = v.ResponseCode
	return res
}

func mapToLog(v *logRecord) *adminapi.Log {
	res := &adminapi.Log{}
	res.Key = v.Key
	res.Date = v.Date
	res.Fail = v.Fail
	res.IP = v.IP
	res.URL = v.URL
	res.QuotaValue = v.QuotaValue
	res.Value = v.Value
	res.ResponseCode = v.ResponseCode
	return res
}
