package api

import "testing"

func TestErrField_Error(t *testing.T) {
	type fields struct {
		Field string
		Msg   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "Simple", fields: fields{Field: "olia", Msg: "bad"}, want: "wrong field 'olia' - bad"},
		{name: "No field", fields: fields{Field: "", Msg: "bad"}, want: "wrong field '' - bad"},
		{name: "No msg", fields: fields{Field: "f1", Msg: ""}, want: "wrong field 'f1' - "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ErrField{
				Field: tt.fields.Field,
				Msg:   tt.fields.Msg,
			}
			if got := r.Error(); got != tt.want {
				t.Errorf("ErrField.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
