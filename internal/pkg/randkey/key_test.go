package randkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name    string
		args    args
		wantLen    int
		wantErr bool
	}{
		{name: "Simple", args: args{n: 10}, wantErr: false, wantLen: 10},
		{name: "30", args: args{n: 30}, wantErr: false, wantLen: 30},
		{name: "Long", args: args{n: 1000}, wantErr: false, wantLen: 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Generate(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("Generate() = %v, want %v", got, tt.wantLen)
			}
		})
	}
}

func TestDiffers(t *testing.T) {
	got, _ := Generate(5)
	got1, _ := Generate(5)
	assert.NotEqual(t, got, got1)
}
