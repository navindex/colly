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
				visits: map[uint64]bool{},
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
		visits map[uint64]bool
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
				visits: map[uint64]bool{},
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
		visits map[uint64]bool
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
				visits: map[uint64]bool{
					1:  true,
					42: true,
				},
			},
			want: &stgVisit{
				lock:   &sync.RWMutex{},
				visits: map[uint64]bool{},
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
		visits map[uint64]bool
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
				visits: map[uint64]bool{
					1:  true,
					42: true,
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

func Test_stgVisit_SetVisited(t *testing.T) {
	type fields struct {
		lock   *sync.RWMutex
		visits map[uint64]bool
	}
	type args struct {
		requestID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *stgVisit
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[uint64]bool{
					1:  true,
					99: true,
				},
			},
			args: args{
				requestID: 42,
			},
			want: &stgVisit{
				lock: &sync.RWMutex{},
				visits: map[uint64]bool{
					1:  true,
					42: true,
					99: true,
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
				requestID: 42,
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
			if err := s.SetVisited(tt.args.requestID); (err != nil) != tt.wantErr {
				t.Errorf("stgVisit.SetVisited() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgVisit.SetVisited() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgVisit_IsVisited(t *testing.T) {
	type fields struct {
		lock   *sync.RWMutex
		visits map[uint64]bool
	}
	type args struct {
		requestID uint64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "visited",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[uint64]bool{
					1:  true,
					42: true,
				},
			},
			args: args{
				requestID: 42,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not visited",
			fields: fields{
				lock: &sync.RWMutex{},
				visits: map[uint64]bool{
					1:  true,
					42: true,
				},
			},
			args: args{
				requestID: 99,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				lock:   &sync.RWMutex{},
				visits: nil,
			},
			args: args{
				requestID: 42,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgVisit{
				lock:   tt.fields.lock,
				visits: tt.fields.visits,
			}
			got, err := s.IsVisited(tt.args.requestID)
			if (err != nil) != tt.wantErr {
				t.Errorf("stgVisit.IsVisited() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stgVisit.IsVisited() = %v, want %v", got, tt.want)
			}
		})
	}
}
