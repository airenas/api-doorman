package handler

import "testing"

func Test_idOrHash(t *testing.T) {
	type args struct {
		ctx *customData
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	{name: "id", args: args{ctx: &customData{KeyID: "id", Key: "olia"}}, want: "id"},
	{name: "key", args: args{ctx: &customData{KeyID: "", Key: "olia"}}, want: "d57329cf35"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := idOrHash(tt.args.ctx); got != tt.want {
				t.Errorf("idOrHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
