package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"github.com/airenas/api-doorman/internal/pkg/randkey"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CmsIntegrator struct {
	sessionProvider        *SessionProvider
	newKeySize             int
	defaultValidToDuration time.Duration
}

//CmsIntegrator creates CmsIntegrator instance
func NewCmsIntegrator(sessionProvider *SessionProvider, keySize int) (*CmsIntegrator, error) {
	f := CmsIntegrator{sessionProvider: sessionProvider}
	if keySize < 10 || keySize > 100 {
		return nil, errors.New("wrong keySize")
	}
	f.newKeySize = keySize
	f.defaultValidToDuration = time.Hour * 24 * 365 * 10 //aprox 10 years
	return &f, nil
}

func (ss *CmsIntegrator) Create(input *api.CreateInput) (*api.Key, bool, error) {
	if err := validateInput(input); err != nil {
		return nil, false, err
	}

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.sessionProvider.NewSession()
	if err != nil {
		return nil, false, err
	}
	defer session.EndSession(context.Background())

	inserted := false

	resInt, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		c := session.Client().Database(keyMapDB).Collection(keyMapTable)
		var keyMap keyMapRecord
		err = c.FindOne(sessCtx, bson.M{"externalID": sanitize(input.ID)}).Decode(&keyMap)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				key, err := ss.createKeyWithQuota(sessCtx, session, input)
				if err == nil {
					inserted = true
					return &api.Key{Key: key.Key}, nil
				}
				return nil, err
			}
			return nil, err
		}
		if keyMap.Project != input.Service {
			return nil, &api.ErrField{Field: "service", Msg: "exists for other service"}
		}
		return &api.Key{Key: keyMap.Key}, nil
	})

	if err != nil {
		return nil, false, err
	}
	res := resInt.(*api.Key)
	return res, inserted, err
}

func (ss *CmsIntegrator) GetKey(keyID string) (*api.Key, error) {
	if keyID == "" {
		return nil, api.ErrNoRecord
	}

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.sessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(keyMapDB).Collection(keyMapTable)
	var keyMapR keyMapRecord
	err = c.FindOne(ctx, bson.M{"externalID": sanitize(keyID)}).Decode(&keyMapR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't load from keymap")
	}
	c = session.Client().Database(keyMapR.Project).Collection(keyTable)
	keyR := &keyRecord{}
	err = c.FindOne(ctx, bson.M{"key": sanitize(keyMapR.Key)}).Decode(&keyR)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, api.ErrNoRecord
		}
		return nil, errors.Wrapf(err, "can't load from %s.key", keyMapR.Project)
	}
	return mapToKey(keyMapR, keyR), nil
}

func mapToKey(keyMapR keyMapRecord, keyR *keyRecord) *api.Key {
	return &api.Key{Key: keyMapR.Key, Service: keyMapR.Project, ValidTo: toTime(&keyR.ValidTo),
		LastUsed: toTime(&keyR.LastUsed), LastIP: keyR.LastIP,
		TotalCredits: keyR.Limit, UsedCredits: keyR.QuotaValue, FailedCredits: keyR.QuotaValueFailed,
		Disabled: keyR.Disabled,
	}
}

func toTime(time *time.Time) *time.Time {
	if (time == nil || time.IsZero()) {
		return nil
	}
	return time
}

func (ss *CmsIntegrator) createKeyWithQuota(ctx context.Context, session mongo.Session, input *api.CreateInput) (*keyRecord, error) {
	// create map
	c := session.Client().Database(keyMapDB).Collection(keyMapTable)
	var keyMap keyMapRecord
	keyMap.Created = time.Now()
	keyMap.ExternalID = input.ID
	keyMap.Key = randkey.Generate(ss.newKeySize)
	keyMap.Project = input.Service
	_, err := c.InsertOne(ctx, keyMap)
	if err != nil {
		if IsDuplicate(err) {
			return nil, errors.New("can't insert keymap - duplicate")
		}
		return nil, errors.Wrap(err, "can't insert keymap")
	}

	c = session.Client().Database(input.Service).Collection(operationTable)
	var operation operationRecord
	operation.Date = time.Now()
	operation.Key = keyMap.Key
	operation.OperationID = input.OperationID
	operation.QuotaValue = input.Credits
	_, err = c.InsertOne(ctx, operation)
	if err != nil {
		if IsDuplicate(err) {
			return nil, &api.ErrField{Field: "operationID", Msg: "duplicate"}
		}
		return nil, errors.Wrap(err, "can't insert operation")
	}
	c = session.Client().Database(input.Service).Collection(keyTable)
	res := &keyRecord{}
	res.Key = keyMap.Key
	res.Limit = input.Credits

	if input.ValidTo != nil {
		res.ValidTo = *input.ValidTo
	} else {
		res.ValidTo = time.Now().Add(ss.defaultValidToDuration)
	}
	res.Created = time.Now()
	res.Manual = true
	if input.SaveRequests {
		res.Tags = []string{"x-tts-collect-data:always"}
	}
	_, err = c.InsertOne(ctx, res)
	if err != nil {
		if IsDuplicate(err) {
			return nil, errors.New("can't insert key - duplicate")
		}
		return nil, errors.Wrap(err, "can't insert key")
	}
	return res, err
}

func validateInput(input *api.CreateInput) error {
	if input == nil {
		return &api.ErrField{Field: "id", Msg: "missing"}
	}
	if strings.TrimSpace(input.ID) == "" {
		return &api.ErrField{Field: "id", Msg: "missing"}
	}
	if strings.TrimSpace(input.OperationID) == "" {
		return &api.ErrField{Field: "operationID", Msg: "missing"}
	}
	if strings.TrimSpace(input.Service) == "" {
		return &api.ErrField{Field: "service", Msg: "missing"}
	}
	if input.ValidTo != nil && input.ValidTo.Before(time.Now()) {
		return &api.ErrField{Field: "validTo", Msg: "past date"}
	}
	if input.Credits <= 0.1 {
		return &api.ErrField{Field: "credits", Msg: "less than 0.1"}
	}
	return nil
}
