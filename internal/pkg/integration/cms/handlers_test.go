package cms

import (
	"testing"
	"time"
)

func Test_parseDateParam(t *testing.T) {
	type args struct {
		s string
	}
	wanted, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{name: "Empty", args: args{s: ""}, wantErr: false},
		{name: "Error", args: args{s: "err"}, wantErr: true},
		{name: "Error", args: args{s: "2006-13-02T15:04:05Z"}, wantErr: true},
		{name: "Error", args: args{s: "2006-11-31T15:04:05Z"}, wantErr: true},
		{name: "Parse", args: args{s: "2006-01-02T15:04:05Z"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T16:04:05+01:00"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T17:04:05+02:00"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T12:04:05-03:00"}, want: wanted, wantErr: false},
		{name: "Parse TZ", args: args{s: "2006-01-02T11:34:05-03:30"}, want: wanted, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDateParam(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.want.IsZero() {
				if got == nil || !(got.Before(tt.want.Add(time.Millisecond)) &&
					got.After(tt.want.Add(-time.Millisecond))) {
					t.Errorf("parseDateParam() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
