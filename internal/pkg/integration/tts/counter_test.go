package tts

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit_FailOnWronURL(t *testing.T) {
	_, err := NewCounter("http://service")
	assert.NotNil(t, err)
	_, err = NewCounter("")
	assert.NotNil(t, err)
}

func TestInit(t *testing.T) {
	d, err := NewCounter("http://localhost:8000/{{requestID}}")
	assert.Nil(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, "requestID", d.GetParamName())
}

func TestCounter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/olia", req.URL.Path)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{"count":123}`))
	}))
	defer server.Close()
	d, _ := NewCounter(server.URL + "/{{reqID}}")
	d.httpclient = server.Client()

	r, err := d.Get("olia")

	assert.Nil(t, err)
	assert.Equal(t, int64(123), r)
}

func TestCounter_Fail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	d, _ := NewCounter(server.URL + "/{{reqID}}")
	d.httpclient = server.Client()

	_, err := d.Get("111")
	assert.NotNil(t, err)
}

func TestCounter_FailJson(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("{"))
	}))
	defer server.Close()

	d, _ := NewCounter(server.URL + "/{{reqID}}")
	d.httpclient = server.Client()

	_, err := d.Get("111")
	assert.NotNil(t, err)
}

func TestCounter_FailTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Millisecond * 10)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{"count":123}`))
	}))
	defer server.Close()

	d, _ := NewCounter(server.URL + "/{{reqID}}")
	d.httpclient = server.Client()
	d.timeOut = time.Millisecond * 5

	_, err := d.Get("111")
	assert.NotNil(t, err)
}

func Test_extractParam(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{name: "extracts", args: "http://tts:8000/check/{{requestID}}", want: "requestID", wantErr: false},
		{name: "fails", args: "", want: "", wantErr: true},
		{name: "fails", args: "http://tts:8000/check/", want: "", wantErr: true},
		{name: "fails", args: "http://tts:8000/check/{requestID}", want: "", wantErr: true},
		{name: "fails", args: "http://tts:8000/check/{{}}", want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractParam(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractParam() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepareURL(t *testing.T) {
	type args struct {
		s     string
		pName string
		param string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "returns", args: args{s: "http://tts:8000/tts/{{reqID}}", pName: "reqID", param: "123-4"},
			want: "http://tts:8000/tts/123-4"},
		{name: "no change", args: args{s: "http://tts:8000/tts/{{reqID}}", pName: "req", param: "123-4"},
			want: "http://tts:8000/tts/{{reqID}}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareURL(tt.args.s, tt.args.pName, tt.args.param); got != tt.want {
				t.Errorf("prepareURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
