package utils

import (
	"strings"
	"testing"
)

func TestHashKey(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{k: ""}, want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{name: "some", args: args{k: "olia"}, want: "d57329cf35760377655bf8417b666cd1b4028878276d8684b9f571746e908996"},
		{name: "long", args: args{k: strings.Repeat("olia", 20)}, want: "9740a2df2322efaa6e4e6005bf4466255232c08b3d3afe5855c80ce947e9be7d"},
		{name: "long", args: args{k: strings.Repeat("olia", 40)}, want: "7e74eea301942d81a8b123c6c4a48568c6332bbf3b7a8962e26ef54d0e10533c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashKey(tt.args.k); got != tt.want {
				t.Errorf("HashKey() = %v, want %v", got, tt.want)
			}
		})
	}

}
