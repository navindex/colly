package mem

import (
	"reflect"
	"sync"
	"testing"
)

// ------------------------------------------------------------------------

func TestNewVisitStorage(t *testing.T) {
	tests := []struct {
		name string
		want *stgVisit
	}{
		{
			name: "default",
			want: &stgVisit{
				lock:   &sync.RWMutex{},
				visits: map[string]uint{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewVisitStorage(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewVisitStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgVisit_Close(t *testing.T) {
	type fields struct {
		lock   *sync.RWMutex
		visits map[string]uint
	}
	tests := []struct {
		name    string
		fields  fields
		want    *stgVisit
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				lock:   &sync.RWMutex{},
				visits: map[string]uint{},
			},
			want: &stgVisit{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			want: &stgVisit{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgVisit{
				lock:   tt.fields.lock,
				visits: tt.fields.visits,
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("stgVisit.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgVisit.Close() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgVisit_Clear(t *testing.T) {
	type fields struct {
		lock   *sync.RWMutex
		visits map[string]uint
	}
	tests := []struct {
		name    string
		fields  fields
		want    *stgVisit
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 2,
					"xyz": 3,
				},
			},
			want: &stgVisit{
				lock:   &sync.RWMutex{},
				visits: map[string]uint{},
			},
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			want: &stgVisit{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgVisit{
				lock:   tt.fields.lock,
				visits: tt.fields.visits,
			}
			if err := s.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("stgVisit.Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgVisit.Clear() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgVisit_Len(t *testing.T) {
	type fields struct {
		lock   *sync.RWMutex
		visits map[string]uint
	}
	tests := []struct {
		name    string
		fields  fields
		want    uint
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 2,
					"xyz": 3,
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgVisit{
				lock:   tt.fields.lock,
				visits: tt.fields.visits,
			}
			got, err := s.Len()
			if (err != nil) != tt.wantErr {
				t.Errorf("stgVisit.Len() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("stgVisit.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgVisit_AddVisit(t *testing.T) {
	type fields struct {
		lock   *sync.RWMutex
		visits map[string]uint
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *stgVisit
		wantErr bool
	}{
		{
			name: "new",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 2,
					"pqr": 6,
				},
			},
			args: args{
				key: "xyz",
			},
			want: &stgVisit{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 2,
					"xyz": 1,
					"pqr": 6,
				},
			},
			wantErr: false,
		},
		{
			name: "existing",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 2,
					"xyz": 6,
				},
			},
			args: args{
				key: "abc",
			},
			want: &stgVisit{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 3,
					"xyz": 6,
				},
			},
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			args: args{
				key: "abc",
			},
			want: &stgVisit{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgVisit{
				lock:   tt.fields.lock,
				visits: tt.fields.visits,
			}
			if err := s.AddVisit(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("stgVisit.AddVisit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgVisit.AddVisit() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgVisit_PastVisits(t *testing.T) {
	type fields struct {
		lock   *sync.RWMutex
		visits map[string]uint
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint
		wantErr bool
	}{
		{
			name: "visited",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 3,
					"xyz": 6,
				},
			},
			args: args{
				key: "pqr",
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "not visited",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[string]uint{
					"abc": 3,
					"xyz": 6,
				},
			},
			args: args{
				key: "pqr",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			args: args{
				key: "pqr",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgVisit{
				lock:   tt.fields.lock,
				visits: tt.fields.visits,
			}
			got, err := s.PastVisits(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("stgVisit.PastVisits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stgVisit.PastVisits() = %v, want %v", got, tt.want)
			}
		})
	}
}
