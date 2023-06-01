package utils

import "testing"

func TestFloat64Equal(t *testing.T) {
	type args struct {
		f1 float64
		f2 float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name:"eq", args: args{f1: 0, f2:0}, want: true},
		{name:"eq", args: args{f1: 0.2, f2:0.2}, want: true},
		{name:"eq", args: args{f1: 10.000002, f2:10.00000200003}, want: true},
		{name:"eq", args: args{f1: -50.00007, f2:-50.00007}, want: true},
		{name:"eq", args: args{f1: 3.33333333, f2:10.0/3}, want: true},
		{name:"neq", args: args{f1: 0, f2:0.1}, want: false},
		{name:"neq", args: args{f1: 0, f2:-0.1}, want: false},
		{name:"neq", args: args{f1: 3.33, f2:10.0/3}, want: false},
		{name:"neq", args: args{f1: 3333, f2:1111.00001*3}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Float64Equal(tt.args.f1, tt.args.f2); got != tt.want {
				t.Errorf("Float64Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
