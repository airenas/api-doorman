package mongodb

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

//IndexData keeps index creation data
type IndexData struct {
	Table  string
	Field  string
	Unique bool
}

//NewIndexData creates index data
func newIndexData(table string, field string, unique bool) IndexData {
	return IndexData{Table: table, Field: field, Unique: unique}
}

//SessionProvider connects and provides session for mongo DB
type SessionProvider struct {
	client  *mongo.Client
	URL     string
	indexes []IndexData
	m       sync.Mutex // struct field mutex
}

//NewSessionProvider creates Mongo session provider
func NewSessionProvider(url string) (*SessionProvider, error) {
	if url == "" {
		return nil, errors.New("No Mongo url provided")
	}
	return &SessionProvider{URL: url, indexes: indexData}, nil
}

//Close closes mongo session
func (sp *SessionProvider) Close() {
	if sp.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		sp.client.Disconnect(ctx)
	}
}

//NewSession creates mongo session
func (sp *SessionProvider) NewSession() (mongo.Session, error) {
	sp.m.Lock()
	defer sp.m.Unlock()

	if sp.client == nil {
		//cmdapp.Log.Info("Dial mongo: " + utils.HidePass(sp.URL))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(sp.URL))
		if err != nil {
			return nil, errors.Wrap(err, "Can't dial to mongo")
		}
		err = checkIndexes(client, sp.indexes)
		if err != nil {
			return nil, errors.Wrap(err, "Can't create indexes")
		}
		sp.client = client
	}
	return sp.client.StartSession()
}

func checkIndexes(s *mongo.Client, indexes []IndexData) error {
	session, err := s.StartSession()
	if err != nil {
		return errors.Wrap(err, "Can't cinit session")
	}
	defer session.EndSession(context.Background())
	for _, index := range indexes {
		err := checkIndex(session, index)
		if err != nil {
			return errors.Wrap(err, "Can't create index: "+index.Table+":"+index.Field)
		}
	}
	return nil
}

func checkIndex(s mongo.Session, indexData IndexData) error {
	c := s.Client().Database(store).Collection(indexData.Table)
	keys := bsonx.Doc{{Key: indexData.Field, Value: bsonx.Int32(int32(1))}}
	tv := true
	index := mongo.IndexModel{
		Keys: keys,
		Options: &options.IndexOptions{Unique: &indexData.Unique,
			Background: &tv,
			Sparse:     &tv,
		}}
	_, err := c.Indexes().CreateOne(context.Background(), index)
	return err
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
