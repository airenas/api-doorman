package tag

import (
	"testing"
)

func TestParseTag(t *testing.T) {
	type args struct {
		hs string
	}
	tests := []struct {
		name    string
		args    args
		wantK   string
		wantV   string
		wantErr bool
	}{
		{name: "Empty", args: args{hs: ""}, wantK: "", wantV: "", wantErr: false},
		{name: "Parse", args: args{hs: "aaa:oooo"}, wantK: "aaa", wantV: "oooo", wantErr: false},
		{name: "Parse upper", args: args{hs: "olia-aAa:oOOo"}, wantK: "olia-aaa", wantV: "oOOo", wantErr: false},
		{name: "Parse Value", args: args{hs: ":aaa:oooo"}, wantK: "", wantV: "aaa:oooo", wantErr: false},
		{name: "Parse Key", args: args{hs: "aaa:"}, wantK: "aaa", wantV: "", wantErr: false},
		{name: "Fail", args: args{hs: "aaa"}, wantK: "", wantV: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotK, gotV, err := Parse(tt.args.hs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotK != tt.wantK {
				t.Errorf("ParseTag() got = %v, want %v", gotK, tt.wantK)
			}
			if gotV != tt.wantV {
				t.Errorf("ParseTag() gotValue = %v, want %v", gotV, tt.wantV)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	type args struct {
		valueCondition string
		value          string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Empty", args: args{valueCondition: "", value: ""}, wantErr: false},
		{name: "Empty value", args: args{valueCondition: "in[1,2,3]", value: ""}, wantErr: false},
		{name: "Any", args: args{valueCondition: "", value: "aaaaaaa"}, wantErr: false},
		{name: "In", args: args{valueCondition: "in[1,2,3]", value: "2"}, wantErr: false},
		{name: "In One", args: args{valueCondition: "in[3]", value: "3"}, wantErr: false},
		{name: "In Last", args: args{valueCondition: "in[1,2,3]", value: "3"}, wantErr: false},
		{name: "In Fail", args: args{valueCondition: "in[1,2,3]", value: "4"}, wantErr: true},
		{name: "In Str", args: args{valueCondition: "in[1,aaaa,3]", value: "aaaa"}, wantErr: false},

		{name: "Between", args: args{valueCondition: "between[1,10000]", value: "1"}, wantErr: false},
		{name: "Between", args: args{valueCondition: "between[1,10000]", value: "1"}, wantErr: false},
		{name: "Between", args: args{valueCondition: "between[1,10000]", value: "10000"}, wantErr: false},
		{name: "Between One", args: args{valueCondition: "between[100,100]", value: "100"}, wantErr: false},
		{name: "Between Fail", args: args{valueCondition: "between[1,3]", value: "4"}, wantErr: true},
		{name: "Between Fail Less", args: args{valueCondition: "between[1,3]", value: "0"}, wantErr: true},
		{name: "Between Fail Str from", args: args{valueCondition: "between[aa,3]", value: "1"}, wantErr: true},
		{name: "Between Fail Str to", args: args{valueCondition: "between[1,bb]", value: "1"}, wantErr: true},
		{name: "Between Fail Str val", args: args{valueCondition: "between[1,3]", value: "vv"}, wantErr: true},
		{name: "Between Fail condition", args: args{valueCondition: "between[1]", value: "1"}, wantErr: true},

		{name: "Other", args: args{valueCondition: "ops[1,10000]", value: "1"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateValue(tt.args.valueCondition, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
