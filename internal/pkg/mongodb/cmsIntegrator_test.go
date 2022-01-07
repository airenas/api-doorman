package mongodb

import (
	"reflect"
	"testing"
	"time"
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
