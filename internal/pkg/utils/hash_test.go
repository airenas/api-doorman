package utils

import (
	"strings"
	"testing"
)

// func TestHashKey(t *testing.T) {
// 	type args struct {
// 		k string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		{name: "empty", args: args{k: ""}, want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
// 		{name: "some", args: args{k: "olia"}, want: "d57329cf35760377655bf8417b666cd1b4028878276d8684b9f571746e908996"},
// 		{name: "long", args: args{k: strings.Repeat("olia", 20)}, want: "9740a2df2322efaa6e4e6005bf4466255232c08b3d3afe5855c80ce947e9be7d"},
// 		{name: "long", args: args{k: strings.Repeat("olia", 40)}, want: "7e74eea301942d81a8b123c6c4a48568c6332bbf3b7a8962e26ef54d0e10533c"},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := HashKey(tt.args.k); got != tt.want {
// 				t.Errorf("HashKey() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}

// }

func TestHashWithHmac(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{k: ""}, want: "6d8196decc3a41883a7627066c9ec660195ba8d4104769fe70f405af218866d0"},
		{name: "some", args: args{k: "olia"}, want: "4514d417c68a5b3adbf1a263e4999b14081ab165a5ebdc46aeb89df1210ae091"},
		{name: "long", args: args{k: strings.Repeat("olia", 20)}, want: "79c8e6d696bf531be598a0b413a80d35b69fcd8643ea54f6972043c4a7e7a1a9"},
		{name: "long", args: args{k: strings.Repeat("olia", 40)}, want: "59baa4eee4b3d31b2ffc5d508d86340a5db018dd017395ab998789704e8c2c3b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashKeyWithHMAC(tt.args.k, "olia"); got != tt.want {
				t.Errorf("HashKey() = %v, want %v", got, tt.want)
			}
		})
	}

}
