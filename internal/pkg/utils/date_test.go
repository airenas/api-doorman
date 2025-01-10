package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestParseDateParam(t *testing.T) {
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
		{name: "Millis", args: args{s: "2006-01-02T11:34:05.123-03:30"}, want: wanted.Add(time.Millisecond * 123),
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateParam(tt.args.s)
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

func TestStartOfMonth(t *testing.T) {
	date := time.Now()
	type args struct {
		now  time.Time
		next int
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{name: "Test", args: args{now: date}, want: time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())},
		{name: "Test next", args: args{now: time.Date(2022, 4, 1, 0, 0, 0, 0, date.Location()), next: 1}, want: time.Date(2022, 5, 1, 0, 0, 0, 0, date.Location())},
		{name: "Same", args: args{now: time.Date(2022, 4, 1, 0, 0, 0, 0, date.Location()), next: 0}, want: time.Date(2022, 4, 1, 0, 0, 0, 0, date.Location())},
		{name: "Test next year", args: args{now: time.Date(2022, 12, 1, 0, 0, 0, 0, date.Location()), next: 1}, want: time.Date(2023, 1, 1, 0, 0, 0, 0, date.Location())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StartOfMonth(tt.args.now, tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StartOfMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
