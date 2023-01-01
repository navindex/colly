package mem

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

var (
	emptyJar = func() *cookiejar.Jar { jar, _ := cookiejar.New(nil); return jar }()
	getURL   = func(str string) *url.URL { u, _ := url.Parse(str); return u }
)

// ------------------------------------------------------------------------

func TestNewCookieStorage(t *testing.T) {
	tests := []struct {
		name    string
		want    *stgCookie
		wantErr bool
	}{
		{
			name: "default",
			want: &stgCookie{
				lock: &sync.Mutex{},
				jar:  emptyJar,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCookieStorage()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewstgCookie() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewstgCookie() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgCookie_Close(t *testing.T) {
	type fields struct {
		lock *sync.Mutex
		jar  *cookiejar.Jar
	}
	tests := []struct {
		name    string
		fields  fields
		want    *stgCookie
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				lock: &sync.Mutex{},
				jar:  emptyJar,
			},
			want: &stgCookie{
				lock: &sync.Mutex{},
				jar:  nil,
			},
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock: &sync.Mutex{},
				jar:  nil,
			},
			want: &stgCookie{
				lock: &sync.Mutex{},
				jar:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgCookie{
				lock: tt.fields.lock,
				jar:  tt.fields.jar,
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("stgCookie.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgCookie.Close() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgCookie_Clear(t *testing.T) {
	type fields struct {
		lock *sync.Mutex
		jar  *cookiejar.Jar
	}

	jar := emptyJar
	cookies := []*http.Cookie{
		{
			Name:  "cookie_name",
			Value: "cookie_value",
		},
	}
	jar.SetCookies(getURL("sample"), cookies)

	tests := []struct {
		name    string
		fields  fields
		want    *stgCookie
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				lock: &sync.Mutex{},
				jar:  jar,
			},
			want: &stgCookie{
				lock: &sync.Mutex{},
				jar:  emptyJar,
			},
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock: &sync.Mutex{},
				jar:  nil,
			},
			want: &stgCookie{
				lock: &sync.Mutex{},
				jar:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgCookie{
				lock: tt.fields.lock,
				jar:  tt.fields.jar,
			}
			if err := s.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("stgCookie.Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgCookie.Clear() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgCookie_Cookies(t *testing.T) {
	type fields struct {
		lock *sync.Mutex
		jar  *cookiejar.Jar
	}
	type args struct {
		u       *url.URL
		cookies []*http.Cookie
	}

	jar := emptyJar
	cookies := []*http.Cookie{
		{
			Name:  "cookie_name",
			Value: "cookie_value",
		},
	}
	jar.SetCookies(getURL("http://sample.com"), cookies)

	newURL := getURL("http://newsample.com")
	newCookies := []*http.Cookie{
		{
			Name:  "new_sample_cookie_1_name",
			Value: "new_sample_cookie_1_value",
		},
		{
			Name:  "new_sample_cookie_2_name",
			Value: "new_sample_cookie_2_value",
		},
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*http.Cookie
		wantErr bool
	}{
		{
			name: "not empty",
			fields: fields{
				lock: &sync.Mutex{},
				jar:  jar,
			},
			args: args{
				u:       newURL,
				cookies: newCookies,
			},
			want:    newCookies,
			wantErr: false,
		},
		{
			name: "empty",
			fields: fields{
				lock: &sync.Mutex{},
				jar:  emptyJar,
			},
			args: args{
				u:       newURL,
				cookies: newCookies,
			},
			want:    newCookies,
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock: &sync.Mutex{},
				jar:  nil,
			},
			args: args{
				u:       newURL,
				cookies: newCookies,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgCookie{
				lock: tt.fields.lock,
				jar:  tt.fields.jar,
			}
			if err := s.SetCookies(tt.args.u, tt.args.cookies); (err != nil) != tt.wantErr {
				t.Errorf("stgCookie.SetCookies() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got, _ := s.Cookies(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stgCookie.Cookies() = %v, want %v", got, tt.want)
			}
		})
	}
}
