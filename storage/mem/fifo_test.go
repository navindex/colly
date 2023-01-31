// FIXME
package mem

import (
	"bytes"
	"io"
	"reflect"
	"sync"
	"testing"
)

// ------------------------------------------------------------------------

func TestNewFIFOStorage(t *testing.T) {
	type args struct {
		maxLength uint
	}
	tests := []struct {
		name string
		args args
		want *stgMultiFIFO
	}{
		{
			name: "default",
			args: args{
				maxLength: 42,
			},
			want: &stgMultiFIFO{
				threads:  map[uint32]*stgFIFO{},
				capacity: 42,
				lock:     &sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFIFOStorage(tt.args.maxLength); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFIFOStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_Close(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		want    *stgMultiFIFO
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				threads: map[uint32]*stgFIFO{
					6: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 10,
						lock:  &sync.Mutex{},
					},
					42: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
					69: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 99,
				lock:     &sync.RWMutex{},
			},
			want: &stgMultiFIFO{
				threads:  map[uint32]*stgFIFO{},
				capacity: 99,
				lock:     &sync.RWMutex{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("stgMultiFIFO.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgFIFO.Close() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_Clear(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	type args struct {
		ids []uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *stgMultiFIFO
		wantErr bool
	}{
		{
			name: "by id",
			fields: fields{
				threads: map[uint32]*stgFIFO{
					6: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 10,
						lock:  &sync.Mutex{},
					},
					42: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
					69: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 77,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 999,
				lock:     &sync.RWMutex{},
			},
			args: args{
				ids: []uint32{6, 69, 99},
			},
			want: &stgMultiFIFO{
				threads: map[uint32]*stgFIFO{
					42: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 999,
				lock:     &sync.RWMutex{},
			},
			wantErr: false,
		},
		{
			name: "all",
			fields: fields{
				threads: map[uint32]*stgFIFO{
					6: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 10,
						lock:  &sync.Mutex{},
					},
					42: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
					69: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 99,
				lock:     &sync.RWMutex{},
			},
			args: args{
				ids: nil,
			},
			want: &stgMultiFIFO{
				threads:  map[uint32]*stgFIFO{},
				capacity: 99,
				lock:     &sync.RWMutex{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			err := s.Clear(tt.args.ids...)
			if (err != nil) != tt.wantErr {
				t.Errorf("stgMultiFIFO.Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgMultiFIFO.Clear() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_Capacity(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		want   uint
	}{
		{
			name: "empty",
			fields: fields{
				threads:  map[uint32]*stgFIFO{},
				capacity: 999,
				lock:     &sync.RWMutex{},
			},
			want: 999,
		},
		{
			name: "with data",
			fields: fields{
				threads: map[uint32]*stgFIFO{
					6: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 10,
						lock:  &sync.Mutex{},
					},
					42: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
					69: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 999,
				lock:     &sync.RWMutex{},
			},
			want: 999,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			if got := s.Capacity(); got != tt.want {
				t.Errorf("stgMultiFIFO.Capacity() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_Len(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	type args struct {
		id uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				threads: map[uint32]*stgFIFO{
					6: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 10,
						lock:  &sync.Mutex{},
					},
					42: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 99,
						lock:  &sync.Mutex{},
					},
					69: {
						head:  &dataNode{},
						tail:  &dataNode{},
						count: 88,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 99,
				lock:     &sync.RWMutex{},
			},
			args: args{
				id: 42,
			},
			want:    99,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			got, err := s.Len(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("stgMultiFIFO.Len() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stgMultiFIFO.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_Push(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	type args struct {
		id   uint32
		item io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			if err := s.Push(tt.args.id, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("stgMultiFIFO.Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_Pop(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	type args struct {
		id uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    io.Reader
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			got, err := s.Pop(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("stgMultiFIFO.Pop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stgMultiFIFO.Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_Peek(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	type args struct {
		id uint32
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    io.Reader
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			got, err := s.Peek(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("stgMultiFIFO.Peek() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stgMultiFIFO.Peek() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_addThread(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	type args struct {
		id uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *stgMultiFIFO
	}{
		{
			name: "new",
			fields: fields{
				threads:  map[uint32]*stgFIFO{},
				capacity: 100,
				lock:     &sync.RWMutex{},
			},
			args: args{
				id: 42,
			},
			want: &stgMultiFIFO{
				threads: map[uint32]*stgFIFO{
					42: {
						head:  nil,
						tail:  nil,
						count: 0,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 100,
				lock:     &sync.RWMutex{},
			},
		},
		{
			name: "existing",
			fields: fields{
				threads: map[uint32]*stgFIFO{
					42: {
						head: &dataNode{
							data: []byte("something"),
						},
						tail:  &dataNode{},
						count: 6,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 100,
				lock:     &sync.RWMutex{},
			},
			args: args{
				id: 42,
			},
			want: &stgMultiFIFO{
				threads: map[uint32]*stgFIFO{
					42: {
						head: &dataNode{
							data: []byte("something"),
						},
						tail:  &dataNode{},
						count: 6,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 100,
				lock:     &sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			if s.addThread(tt.args.id); !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgMultiFIFO.addThread() = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgMultiFIFO_hasThread(t *testing.T) {
	type fields struct {
		threads  map[uint32]*stgFIFO
		capacity uint
		lock     *sync.RWMutex
	}
	type args struct {
		id uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "missing",
			fields: fields{
				threads:  map[uint32]*stgFIFO{},
				capacity: 100,
				lock:     &sync.RWMutex{},
			},
			args: args{
				id: 42,
			},
			want: false,
		},
		{
			name: "existing",
			fields: fields{
				threads: map[uint32]*stgFIFO{
					42: {
						head: &dataNode{
							data: []byte("something"),
						},
						tail:  &dataNode{},
						count: 6,
						lock:  &sync.Mutex{},
					},
				},
				capacity: 100,
				lock:     &sync.RWMutex{},
			},
			args: args{
				id: 42,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgMultiFIFO{
				threads:  tt.fields.threads,
				capacity: tt.fields.capacity,
				lock:     tt.fields.lock,
			}
			if got := s.hasThread(tt.args.id); got != tt.want {
				t.Errorf("stgMultiFIFO.hasThread() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_len(t *testing.T) {
	type fields struct {
		head  *dataNode
		tail  *dataNode
		count uint
		lock  *sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   uint
	}{
		{
			name: "default",
			fields: fields{
				head:  &dataNode{},
				tail:  &dataNode{},
				count: 10,
				lock:  &sync.Mutex{},
			},
			want: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgFIFO{
				head:  tt.fields.head,
				tail:  tt.fields.tail,
				count: tt.fields.count,
				lock:  tt.fields.lock,
			}
			if got := s.len(); got != tt.want {
				t.Errorf("stgFIFO.len() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_push(t *testing.T) {
	id := uint32(42)
	s := NewFIFOStorage(100)
	s.addThread(id)
	ss := s.threads[id]

	type args struct {
		data []byte
	}
	tests := []struct {
		name     string
		args     args
		wantLen  uint
		wantPeek io.Reader
		wantErr  bool
	}{
		{
			name: "first",
			args: args{
				data: []byte("first item"),
			},
			wantLen:  1,
			wantPeek: bytes.NewReader([]byte("first item")),
			wantErr:  false,
		},
		{
			name: "second",
			args: args{
				data: []byte("second item"),
			},
			wantLen:  2,
			wantPeek: bytes.NewReader([]byte("first item")),
			wantErr:  false,
		},
		{
			name: "third",
			args: args{
				data: []byte("third item"),
			},
			wantLen:  3,
			wantPeek: bytes.NewReader([]byte("first item")),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ss.push(tt.args.data, s.capacity); (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.push() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len := ss.len(); len != tt.wantLen {
				t.Errorf("stgFIFO.len() = %v, want %v", len, tt.wantLen)
			}
			if got, _ := ss.peek(); !reflect.DeepEqual(got, tt.wantPeek) {
				t.Errorf("stgFIFO.peek() = %v, want %v", got, tt.wantPeek)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_pop(t *testing.T) {
	tests := []struct {
		name    string
		thread  *stgFIFO
		want    io.Reader
		wantLen uint
		wantErr bool
	}{
		{
			name:    "empty",
			thread:  &stgFIFO{lock: &sync.Mutex{}},
			want:    nil,
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "single",
			thread: func() *stgFIFO {
				s := &stgFIFO{lock: &sync.Mutex{}}
				s.push([]byte("first item"), uint(100))
				return s
			}(),
			want:    bytes.NewReader([]byte("first item")),
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "multi",
			thread: func() *stgFIFO {
				s := &stgFIFO{lock: &sync.Mutex{}}
				s.push([]byte("first item"), uint(100))
				s.push([]byte("second item"), uint(100))
				return s
			}(),
			want:    bytes.NewReader([]byte("first item")),
			wantLen: 1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.thread.pop()
			if (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.pop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stgFIFO.pop() = %v, want %v", got, tt.want)
			}
			if len := tt.thread.len(); len != tt.wantLen {
				t.Errorf("stgFIFO.len() = %v, want %v", len, tt.wantLen)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_peek(t *testing.T) {
	tests := []struct {
		name    string
		thread  *stgFIFO
		want    io.Reader
		wantLen uint
		wantErr bool
	}{
		{
			name:    "empty",
			thread:  &stgFIFO{lock: &sync.Mutex{}},
			want:    nil,
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "single",
			thread: func() *stgFIFO {
				s := &stgFIFO{lock: &sync.Mutex{}}
				s.push([]byte("first item"), uint(100))
				return s
			}(),
			want:    bytes.NewReader([]byte("first item")),
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "multi",
			thread: func() *stgFIFO {
				s := &stgFIFO{lock: &sync.Mutex{}}
				s.push([]byte("first item"), uint(100))
				s.push([]byte("second item"), uint(100))
				return s
			}(),
			want:    bytes.NewReader([]byte("first item")),
			wantLen: 2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.thread.peek()
			if (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.peek() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stgFIFO.peek() = %v, want %v", got, tt.want)
			}
			if len := tt.thread.len(); len != tt.wantLen {
				t.Errorf("stgFIFO.Len() = %v, want %v", len, tt.wantLen)
			}
		})
	}
}
