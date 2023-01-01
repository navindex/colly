package mem

import (
	"reflect"
	"sync"
	"testing"
)

// ------------------------------------------------------------------------

var queueData = func(s *stgFIFO) [][]byte {
	bytes := [][]byte{}
	for node := s.head; node != nil; node = node.next {
		bytes = append(bytes, node.data)
	}
	if len(bytes) == 0 {
		return nil
	}
	return bytes
}

// ------------------------------------------------------------------------

func TestNewFIFOStorage(t *testing.T) {
	type args struct {
		maxLength uint
	}
	tests := []struct {
		name string
		args args
		want *stgFIFO
	}{
		{
			name: "default",
			args: args{
				maxLength: 42,
			},
			want: &stgFIFO{
				head:     nil,
				tail:     nil,
				count:    0,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   false,
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

func Test_stgFIFO_Close(t *testing.T) {
	type fields struct {
		head     *dataNode
		tail     *dataNode
		count    uint
		maxCount uint
		lock     *sync.Mutex
		closed   bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    *stgFIFO
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   false,
			},
			want: &stgFIFO{
				head:     nil,
				tail:     nil,
				count:    0,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			want: &stgFIFO{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgFIFO{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				count:    tt.fields.count,
				maxCount: tt.fields.maxCount,
				lock:     tt.fields.lock,
				closed:   tt.fields.closed,
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgFIFO.Close() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_Clear(t *testing.T) {
	type fields struct {
		head     *dataNode
		tail     *dataNode
		count    uint
		maxCount uint
		lock     *sync.Mutex
		closed   bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    *stgFIFO
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   false,
			},
			want: &stgFIFO{
				head:     nil,
				tail:     nil,
				count:    0,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   false,
			},
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			want: &stgFIFO{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgFIFO{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				count:    tt.fields.count,
				maxCount: tt.fields.maxCount,
				lock:     tt.fields.lock,
				closed:   tt.fields.closed,
			}
			if err := s.Clear(); (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.Clear() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("stgFIFO.Close() storage = %v, want %v", s, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_Len(t *testing.T) {
	type fields struct {
		head     *dataNode
		tail     *dataNode
		count    uint
		maxCount uint
		lock     *sync.Mutex
		closed   bool
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
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   false,
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgFIFO{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				count:    tt.fields.count,
				maxCount: tt.fields.maxCount,
				lock:     tt.fields.lock,
				closed:   tt.fields.closed,
			}
			got, err := s.Len()
			if (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.Len() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("stgFIFO.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_Push(t *testing.T) {
	type fields struct {
		head     *dataNode
		tail     *dataNode
		count    uint
		maxCount uint
		lock     *sync.Mutex
		closed   bool
	}

	items := [][]byte{
		[]byte("first item"),
		[]byte("second item"),
		[]byte("third item"),
		[]byte("fourth item"),
		[]byte("fifth item"),
	}

	tests := []struct {
		name    string
		fields  fields
		items   [][]byte
		want    [][]byte
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				head:     nil,
				tail:     nil,
				count:    0,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   false,
			},
			items:   items,
			want:    items,
			wantErr: false,
		},
		{
			name: "full",
			fields: fields{
				head:     nil,
				tail:     nil,
				count:    0,
				maxCount: 3,
				lock:     &sync.Mutex{},
				closed:   false,
			},
			items:   items,
			want:    items[:3],
			wantErr: true,
		},
		{
			name: "closed",
			fields: fields{
				head:     nil,
				tail:     nil,
				count:    0,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			items:   items,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgFIFO{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				count:    tt.fields.count,
				maxCount: tt.fields.maxCount,
				lock:     tt.fields.lock,
				closed:   tt.fields.closed,
			}
			var err error
			for _, item := range tt.items {
				err = s.Push(item)
				if err != nil {
					break
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.Push() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := queueData(s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stgFIFO.Push() data = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_stgFIFO_Pop(t *testing.T) {
	data := map[int][]byte{
		1: []byte("first item"),
		2: []byte("second item"),
		3: []byte("third item"),
	}

	s := NewFIFOStorage(42)
	for i := 1; i <= 3; i++ {
		s.Push(data[i])
	}

	t.Run("default", func(t *testing.T) {
		for i := 1; i <= 4; i++ {
			wantErr := (i == 4)
			var wantData []byte
			if i < 4 {
				wantData = data[i]
			}

			got, err := s.Pop()
			if (err != nil) != wantErr {
				t.Errorf("stgFIFO.Pop() #1 error = %v, wantErr %v", err, wantErr)
			}
			if !reflect.DeepEqual(got, wantData) {
				t.Errorf("stgFIFO.Pop() = %v, want %v", got, wantData)
			}
		}
	})
}

// ------------------------------------------------------------------------

func Test_stgFIFO_Peek(t *testing.T) {
	type fields struct {
		head     *dataNode
		tail     *dataNode
		count    uint
		maxCount uint
		lock     *sync.Mutex
		closed   bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantLen uint
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				head: &dataNode{
					data: []byte("head_data"),
				},
				tail: &dataNode{
					data: []byte("tail_data"),
				},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   false,
			},
			want:    []byte("head_data"),
			wantLen: 10,
			wantErr: false,
		},
		{
			name: "closed",
			fields: fields{
				head:     &dataNode{},
				tail:     &dataNode{},
				count:    10,
				maxCount: 42,
				lock:     &sync.Mutex{},
				closed:   true,
			},
			want:    nil,
			wantLen: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &stgFIFO{
				head:     tt.fields.head,
				tail:     tt.fields.tail,
				count:    tt.fields.count,
				maxCount: tt.fields.maxCount,
				lock:     tt.fields.lock,
				closed:   tt.fields.closed,
			}
			got, err := s.Peek()
			if (err != nil) != tt.wantErr {
				t.Errorf("stgFIFO.Peek() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stgFIFO.Peek() = %v, want %v", got, tt.want)
			}
			if len, _ := s.Len(); len != tt.wantLen {
				t.Errorf("stgFIFO.Len() = %v, want %v", len, tt.wantLen)
			}

		})
	}
}
