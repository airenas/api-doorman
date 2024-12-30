package mongodb

import (
	"strings"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// SProvider session provider wrapper
type SProvider interface {
	NewSession() (mongo.Session, error)
	CheckIndexes(dbs []string) error
}

// DBProvider keeps SessionProvider and database for mongo DB
type DBProvider struct {
	sessionP SProvider
	db       string
}

// NewDBProvider creates Mongo session provider and opens client with selected db
func NewDBProvider(sessionP SProvider, db string) (*DBProvider, error) {
	if sessionP == nil {
		return nil, errors.New("no SessionProvider provided")
	}
	db = strings.TrimSpace(db)
	if db == "" {
		return nil, errors.New("no DB provided")
	}
	if err := sessionP.CheckIndexes([]string{db}); err != nil {
		return nil, errors.Wrapf(err, "fail index check")
	}
	return &DBProvider{sessionP: sessionP, db: db}, nil
}

// NewSesionDatabase creates mongo session and databse
func (sdp *DBProvider) NewSesionDatabase() (mongo.Session, *mongo.Database, error) {
	session, err := sdp.sessionP.NewSession()
	if err != nil {
		return nil, nil, errors.Wrap(err, "can't open session")
	}
	return session, session.Client().Database(sdp.db), nil
}
