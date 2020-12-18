package mongodb

import (
	"strings"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

//DBProvider keeps SessionProvider and database for mongo DB
type DBProvider struct {
	sessionP *SessionProvider
	db       string
}

//NewDBProvider creates Mongo session provider and opens client with selected db
func NewDBProvider(sessionP *SessionProvider, db string) (*DBProvider, error) {
	if sessionP == nil {
		return nil, errors.New("No SessionProvider provided")
	}
	db = strings.TrimSpace(db)
	if db == "" {
		return nil, errors.New("No DB provided")
	}
	return &DBProvider{sessionP: sessionP, db: db}, nil
}

//NewSesionDatabase creates mongo session and databse
func (sdp *DBProvider) NewSesionDatabase() (mongo.Session, *mongo.Database, error) {
	session, err := sdp.sessionP.NewSession(sdp.db)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Can't open session")
	}
	return session, session.Client().Database(sdp.db), nil
}
