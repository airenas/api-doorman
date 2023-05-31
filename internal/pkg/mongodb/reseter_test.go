package mongodb

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func Test_keyFilterIP(t *testing.T) {
	type args struct {
		keyID string
	}
	tests := []struct {
		name string
		args args
		want bson.M
	}{
		{name: "uuid", args: args{keyID: "olia-olia"}, want: bson.M{"keyID": "olia-olia", "manual": false}},
		{name: "pure key", args: args{keyID: "olia"}, want: bson.M{"key": "olia", "manual": false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyFilterIP(tt.args.keyID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("keyFilterIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKeyNoHash(t *testing.T) {
	type args struct {
		id  string
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "id", args: args{id: "id", key: "key"}, want: "id"},
		{name: "key", args: args{id: "", key: "key"}, want: "key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getKeyNoHash(tt.args.id, tt.args.key); got != tt.want {
				t.Errorf("getKeyNoHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
