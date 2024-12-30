package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
)

// LogSaver saves log info to mongo db
type LogSaver struct {
	SessionProvider *DBProvider
}

// NewLogSaver creates LogSaver instance
func NewLogSaver(sessionProvider *DBProvider) (*LogSaver, error) {
	f := LogSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save log key to DB
func (ss *LogSaver) Save(logr *adminapi.Log) error {
	log.Info().Msgf("Saving log - %s, ip: %s, response: %d", logr.URL, logr.IP, logr.ResponseCode)
	ctx, cancel := mongoContext()
	defer cancel()

	session, db, err := ss.SessionProvider.NewSesionDatabase()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	c := db.Collection(logTable)

	_, err = c.InsertOne(ctx, mapFromLog(logr))
	return err
}

// LogProvider retrieves the log
type LogProvider struct {
	SessionProvider *SessionProvider
}

// NewLogProvider creates LogProvider instance
func NewLogProvider(sessionProvider *SessionProvider) (*LogProvider, error) {
	f := LogProvider{SessionProvider: sessionProvider}
	return &f, nil
}

// Get return all logs for key
func (ss *LogProvider) Get(project, key string) ([]*adminapi.Log, error) {
	log.Info().Msgf("getting log list for key")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(logTable)
	cursor, err := c.Find(ctx, bson.M{getLogKeyField(key): Sanitize(key)})
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
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("can't get logs: %w", err)
	}
	return res, nil
}

func getLogKeyField(key string) string {
	if strings.Contains(key, "-") { // UUID?
		return "keyID"
	}
	return "key"
}

func (ss *LogProvider) List(project string, to time.Time) ([]*adminapi.Log, error) {
	log.Info().Msgf("getting log list up to date")
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(logTable)
	cursor, err := c.Find(ctx, bson.M{"date": bson.M{"$lt": to}})
	if err != nil {
		return nil, fmt.Errorf("can't get logs: %w", err)
	}
	defer cursor.Close(ctx)
	res := make([]*adminapi.Log, 0)
	for cursor.Next(ctx) {
		var logR logRecord
		if err = cursor.Decode(&logR); err != nil {
			return nil, fmt.Errorf("can't get logs: %w", err)
		}
		res = append(res, mapToLog(&logR))
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("can't get logs: %w", err)
	}
	return res, nil
}

func (ss *LogProvider) Delete(project string, to time.Time) (int, error) {
	log.Info().Msgf("deleting log list")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return 0, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(project).Collection(logTable)
	res, err := c.DeleteMany(ctx, bson.M{"date": bson.M{"$lt": to}})
	if err != nil {
		return 0, fmt.Errorf("can't delete logs: %w", err)
	}
	return int(res.DeletedCount), nil
}

func mapFromLog(v *adminapi.Log) *logRecord {
	res := &logRecord{}
	if getLogKeyField(v.Key) == "key" {
		res.Key = v.Key
	} else {
		res.KeyID = v.Key
	}
	res.Date = v.Date
	res.Fail = v.Fail
	res.IP = v.IP
	res.URL = v.URL
	res.QuotaValue = v.QuotaValue
	res.Value = v.Value
	res.ResponseCode = v.ResponseCode
	res.RequestID = v.RequestID
	res.ErrorMsg = v.ErrorMsg
	return res
}

func mapToLog(v *logRecord) *adminapi.Log {
	res := &adminapi.Log{}
	res.Key = getKey(v.KeyID, v.Key)
	res.Date = v.Date
	res.Fail = v.Fail
	res.IP = v.IP
	res.URL = v.URL
	res.QuotaValue = v.QuotaValue
	res.Value = v.Value
	res.ResponseCode = v.ResponseCode
	res.RequestID = v.RequestID
	res.ErrorMsg = v.ErrorMsg
	return res
}

func getKey(id, key string) string {
	if id != "" {
		return id
	}
	return utils.HashKey(key)
}
