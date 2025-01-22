package tag

import "testing"

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
