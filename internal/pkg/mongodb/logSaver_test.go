package mongodb

import (
	"reflect"
	"testing"

	adminapi "github.com/airenas/api-doorman/internal/pkg/admin/api"
)

func Test_getLogKeyField(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "simple", args: args{key: "70fb0d8c23784848d019e07544be5f0"}, want: "key"},
		{name: "UUID", args: args{key: "70fb0d8c-2378-4848-ad01-9e07544be5f0"}, want: "keyID"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLogKeyField(tt.args.key); got != tt.want {
				t.Errorf("getLogKeyField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKey(t *testing.T) {
	type args struct {
		id  string
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "id", args: args{id: "id", key: "olia"}, want: "id"},
		{name: "key", args: args{id: "", key: "olia"}, want: "d57329cf35760377655bf8417b666cd1b4028878276d8684b9f571746e908996"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getKey(tt.args.id, tt.args.key); got != tt.want {
				t.Errorf("getKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapFromLog(t *testing.T) {
	type args struct {
		v *adminapi.Log
	}
	tests := []struct {
		name string
		args args
		want *logRecord
	}{
		{name: "key", args: args{v: &adminapi.Log{Key: "olia"}}, want: &logRecord{Key: "olia"}},
		{name: "ID as UUID", args: args{v: &adminapi.Log{Key: "uuid-olia"}}, want: &logRecord{KeyID: "uuid-olia"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapFromLog(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapFromLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mapToLog(t *testing.T) {
	type args struct {
		v *logRecord
	}
	tests := []struct {
		name string
		args args
		want *adminapi.Log
	}{
		{name: "key", args: args{v: &logRecord{Key: "olia"}}, want: &adminapi.Log{Key: "d57329cf35760377655bf8417b666cd1b4028878276d8684b9f571746e908996"}},
		{name: "ID", args: args{v: &logRecord{KeyID: "olia"}}, want: &adminapi.Log{Key: "olia"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapToLog(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapToLog() = %v, want %v", got, tt.want)
			}
		})
	}
}
