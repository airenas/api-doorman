package mongodb

import (
	"context"
	"sync"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/utils"
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

//IndexData keeps index creation data
type IndexData struct {
	Table  string
	Fields []string
	Unique bool
}

//NewIndexData creates index data
func newIndexData(table string, fields []string, unique bool) IndexData {
	return IndexData{Table: table, Fields: fields, Unique: unique}
}

//SessionProvider connects and provides session for mongo DB
type SessionProvider struct {
	client *mongo.Client
	URL    string
	m      sync.Mutex // struct field mutex
}

//NewSessionProvider creates Mongo session provider
func NewSessionProvider(url string) (*SessionProvider, error) {
	if url == "" {
		return nil, errors.New("No Mongo url provided")
	}
	res := &SessionProvider{URL: url}
	return res, nil
}

//Close closes mongo session
func (sp *SessionProvider) Close() {
	if sp.client != nil {
		ctx, cancel := mongoContext()
		defer cancel()
		sp.client.Disconnect(ctx)
	}
}

//NewSession creates mongo session
func (sp *SessionProvider) NewSession() (mongo.Session, error) {
	sp.m.Lock()
	defer sp.m.Unlock()

	if sp.client == nil {
		goapp.Log.Info("Dial mongo: " + utils.HidePass(sp.URL))
		ctx, cancel := mongoContext()
		defer cancel()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(sp.URL))
		if err != nil {
			return nil, errors.Wrap(err, "can't dial to mongo")
		}
		sp.client = client
	}
	return sp.client.StartSession()
}

func checkIndexes(session mongo.Session, indexes []IndexData, database string) error {
	for _, index := range indexes {
		err := checkIndex(session, index, database)
		if err != nil {
			return errors.Wrapf(err, "Can't create index: %s:%v", index.Table, index.Fields)
		}
	}
	return nil
}

func checkIndex(s mongo.Session, indexData IndexData, database string) error {
	c := s.Client().Database(database).Collection(indexData.Table)
	keys := bsonx.Doc{}
	for _, f := range indexData.Fields {
		keys = keys.Append(f, bsonx.Int32(int32(1)))
	}
	index := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(indexData.Unique).SetBackground(true).SetSparse(true),
	}
	_, err := c.Indexes().CreateOne(context.Background(), index)
	return err
}

// Healthy checks if mongo DB is up
func (sp *SessionProvider) CheckIndexes(dbs []string) error {
	session, err := sp.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	if err := checkIndexes(session, keyMapIndexData, keyMapDB); err != nil {
		return errors.Wrapf(err, "index fail for collection %s", keyMapDB)
	}
	for _, db := range dbs {
		if err := checkIndexes(session, indexData, db); err != nil {
			return errors.Wrapf(err, "index fail for collection %s", db)
		}
	}
	return nil
}

// Healthy checks if mongo DB is up
func (sp *SessionProvider) Healthy() error {
	session, err := sp.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())
	return session.Client().Ping(context.Background(), nil)
}

func mongoContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
