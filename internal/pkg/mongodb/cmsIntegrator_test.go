package mongodb

import (
	"reflect"
	"testing"
	"time"

	"github.com/airenas/api-doorman/internal/pkg/integration/cms/api"
	"go.mongodb.org/mongo-driver/bson"
)

func Test_toTime(t *testing.T) {
	type args struct {
		time *time.Time
	}
	tn := time.Now()
	tests := []struct {
		name string
		args args
		want *time.Time
	}{
		{name: "nil", args: args{time: nil}, want: nil},
		{name: "empty", args: args{time: &time.Time{}}, want: nil},
		{name: "nil", args: args{time: &tn}, want: &tn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toTime(tt.args.time); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateCreditsInput(t *testing.T) {
	type args struct {
		input *api.CreditsInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "OK", args: args{input: &api.CreditsInput{OperationID: "1", Credits: 10}}, wantErr: false},
		{name: "Fail", args: args{input: &api.CreditsInput{OperationID: "1", Credits: 0}}, wantErr: true},
		{name: "Fail", args: args{input: &api.CreditsInput{OperationID: "1", Credits: 0.01}}, wantErr: true},
		{name: "Fail", args: args{input: &api.CreditsInput{OperationID: "", Credits: 10}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateCreditsInput(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("validateCreditsInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateInput(t *testing.T) {
	type args struct {
		input *api.CreateInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "OK", args: args{input: &api.CreateInput{OperationID: "1",
			Credits: 10, ID: "1", Service: "srv", SaveRequests: false,
			ValidTo: &[]time.Time{time.Now().Add(time.Minute)}[0]}}, wantErr: false},
		{name: "Fail", args: args{input: &api.CreateInput{OperationID: "",
			Credits: 10, ID: "1", Service: "srv", SaveRequests: false,
			ValidTo: &[]time.Time{time.Now().Add(time.Minute)}[0]}}, wantErr: true},
		{name: "Fail", args: args{input: &api.CreateInput{OperationID: "1",
			Credits: 0, ID: "1", Service: "srv", SaveRequests: false,
			ValidTo: &[]time.Time{time.Now().Add(time.Minute)}[0]}}, wantErr: true},
		{name: "Fail", args: args{input: &api.CreateInput{OperationID: "1",
			Credits: 0.09, ID: "1", Service: "srv", SaveRequests: false,
			ValidTo: &[]time.Time{time.Now().Add(time.Minute)}[0]}}, wantErr: true},
		{name: "Fail", args: args{input: &api.CreateInput{OperationID: "1",
			Credits: 10, ID: "", Service: "srv", SaveRequests: false,
			ValidTo: &[]time.Time{time.Now().Add(time.Minute)}[0]}}, wantErr: true},
		{name: "Fail", args: args{input: &api.CreateInput{OperationID: "1",
			Credits: 10, ID: "1", Service: "", SaveRequests: false,
			ValidTo: &[]time.Time{time.Now().Add(time.Minute)}[0]}}, wantErr: true},
		{name: "Fail", args: args{input: &api.CreateInput{OperationID: "1",
			Credits: 10, ID: "1", Service: "srv", SaveRequests: false,
			ValidTo: &[]time.Time{time.Now().Add(-time.Minute)}[0]}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateInput(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_makeDateFilter(t *testing.T) {
	type args struct {
		key  string
		from *time.Time
		to   *time.Time
	}
	from := time.Now()
	to := from.Add(time.Minute)
	tests := []struct {
		name string
		args args
		want bson.M
	}{
		{name: "Key", args: args{key: "id"}, want: bson.M{"key": "id"}},
		{name: "Dates", args: args{key: "id", from: &from, to: &to},
			want: bson.M{"key": "id", "date": bson.M{"$gte": from, "$lt": to}}},
		{name: "From", args: args{key: "id", from: &from},
			want: bson.M{"key": "id", "date": bson.M{"$gte": from}}},
		{name: "To", args: args{key: "id", to: &to},
			want: bson.M{"key": "id", "date": bson.M{"$lt": to}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeDateFilter(tt.args.key, tt.args.from, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeDateFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapLogRecord(t *testing.T) {
	type args struct {
		log *logRecord
	}
	now := time.Now()
	tests := []struct {
		name string
		args args
		want *api.Log
	}{
		{name: "All", args: args{log: &logRecord{URL: "url", QuotaValue: 10, IP: "ip",
			ResponseCode: 400, Fail: true, Date: now}},
			want: &api.Log{UsedCredits: 10, Date: &now, IP: "ip", Fail: true, Response: 400}},
		{name: "No date", args: args{log: &logRecord{URL: "url", QuotaValue: 10, IP: "ip",
			ResponseCode: 400, Fail: true, Date: time.Time{}}},
			want: &api.Log{UsedCredits: 10, Date: nil, IP: "ip", Fail: true, Response: 400}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapLogRecord(tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapLogRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapToKey(t *testing.T) {
	type args struct {
		keyMapR *keyMapRecord
		keyR    *keyRecord
	}
	now := time.Now()
	tests := []struct {
		name string
		args args
		want *api.Key
	}{
		{name: "All", args: args{keyMapR: &keyMapRecord{Key: "key", Project: "pro", ExternalID: "ID"},
			keyR: &keyRecord{Key: "key", Limit: 100, QuotaValue: 10, QuotaValueFailed: 20,
				Manual: true, ValidTo: now, Created: now, Updated: now, LastUsed: now, LastIP: "ip", 
				IPWhiteList: "ipw",
				Disabled: true}},
			want: &api.Key{UsedCredits: 10, Key: "key", Service: "pro", TotalCredits: 100,
				FailedCredits: 20, Disabled: true, IPWhiteList: "ipw", LastIP: "ip",
				LastUsed: &now, ValidTo: &now, Created: &now, Updated: &now}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapToKey(tt.args.keyMapR, tt.args.keyR); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapToKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_initNewKey(t *testing.T) {
	type args struct {
		input       *api.CreateInput
		defDuration time.Duration
		now         time.Time
	}
	now := time.Now()
	tests := []struct {
		name string
		args args
		want *keyRecord
	}{
		{name: "All", args: args{input: &api.CreateInput{ID: "1", Credits: 100}, defDuration: time.Minute, now: now},
			want: &keyRecord{Manual: true, ValidTo: now.Add(time.Minute), Created: now, Limit: 100}},
		{name: "ValidTo", args: args{input: &api.CreateInput{ID: "1", Credits: 100,
			ValidTo: &[]time.Time{now.Add(time.Hour)}[0]},
			defDuration: time.Minute, now: now},
			want: &keyRecord{Manual: true, ValidTo: now.Add(time.Hour), Created: now, Limit: 100}},
		{name: "SaveRequest", args: args{input: &api.CreateInput{ID: "1", SaveRequests: true},
			defDuration: time.Minute, now: now},
			want: &keyRecord{Manual: true, ValidTo: now.Add(time.Minute), Created: now,
				Tags: []string{"x-tts-collect-data:always"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initNewKey(tt.args.input, tt.args.defDuration, tt.args.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initNewKey() = %v, want %v", got, tt.want)
			}
		})
	}
}