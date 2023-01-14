package colly

import (
	"math"
	"reflect"
	"sync"
	"testing"
)

// ------------------------------------------------------------------------

func TestNewEventList(t *testing.T) {
	tests := []struct {
		name string
		want *eventList
	}{
		{
			name: "default",
			want: &eventList{
				events: map[uint8]*evenArgList{},
				lock:   &sync.RWMutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEventList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEventArgList() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_eventList_Add(t *testing.T) {
	type fields struct {
		events map[uint8]*evenArgList
		lock   *sync.RWMutex
	}
	type args struct {
		event uint8
		arg   string
		item  any
		index []int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[uint8]*evenArgList
	}{
		{
			name: "append to empty arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_1",
				item:  "sixty-nine",
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{
								0: "sixty-nine",
							},
							sorted: []any{"sixty-nine"},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 6,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
		{
			name: "add to empty arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_1",
				item:  "sixty-nine",
				index: []int{3},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{
								3: "sixty-nine",
							},
							sorted: []any{"sixty-nine"},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 6,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
		{
			name: "append",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_A",
				item:  "sixty-nine",
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
								201: "sixty-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred", "sixty-nine"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 6,
				},
			},
		},
		{
			name: "add",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_A",
				item:  "sixty-nine",
				index: []int{69},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
								69:  "sixty-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "sixty-nine", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 6,
				},
			},
		},
		{
			name: "replace",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_A",
				item:  "two hundred and more",
				index: []int{200},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred and more",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred and more"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
		{
			name: "append to missing arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_C",
				item:  "ten",
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_C": {
							original: map[int]any{
								0: "ten",
							},
							sorted: []any{"ten"},
						},
					},
					counter: 6,
				},
			},
		},
		{
			name: "add to missing arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_C",
				item:  "ten",
				index: []int{10},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_C": {
							original: map[int]any{
								10: "ten",
							},
							sorted: []any{"ten"},
						},
					},
					counter: 6,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &eventList{
				events: tt.fields.events,
				lock:   tt.fields.lock,
			}
			el.Add(tt.args.event, tt.args.arg, tt.args.item, tt.args.index...)
			if !reflect.DeepEqual(el.events, tt.want) {
				t.Errorf("eventList.Add() events = %v, want %v", el.events, tt.want)
			}
			for i := range el.events {
				if !reflect.DeepEqual(*el.events[i], *tt.want[i]) {
					t.Errorf("eventList.Add() events[%d] = %v, want %v", i, *el.events[i], *tt.want[i])
				}
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_eventList_Remove(t *testing.T) {
	type fields struct {
		events map[uint8]*evenArgList
		lock   *sync.RWMutex
	}
	type args struct {
		event uint8
		arg   string
		index []int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[uint8]*evenArgList
	}{
		{
			name: "empty arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_1",
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
		{
			name: "index in empty arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_1",
				index: []int{3},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
		{
			name: "full arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_A",
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 0,
				},
			},
		},
		{
			name: "index",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_A",
				index: []int{7},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 4,
				},
			},
		},
		{
			name: "missing arg",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_C",
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
		{
			name: "missing index",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_2",
				index: []int{222},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
					},
					counter: 5,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
		{
			name: "multi index",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_2",
				index: []int{222, 99, 7},
			},
			want: map[uint8]*evenArgList{
				10: {
					args: map[string]*eventArgItemList{
						"arg_1": {
							original: map[int]any{},
							sorted:   []any{},
						},
						"arg_2": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
							},
							sorted: []any{"six", "forty-two", "two hundred"},
						},
					},
					counter: 3,
				},
				20: {
					args: map[string]*eventArgItemList{
						"arg_A": {
							original: map[int]any{
								6:   "six",
								42:  "forty-two",
								200: "two hundred",
								7:   "seven",
								99:  "ninety-nine",
							},
							sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
						},
						"arg_B": {
							original: map[int]any{},
							sorted:   []any{},
						},
					},
					counter: 5,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &eventList{
				events: tt.fields.events,
				lock:   tt.fields.lock,
			}
			el.Remove(tt.args.event, tt.args.arg, tt.args.index...)
			if !reflect.DeepEqual(el.events, tt.want) {
				t.Errorf("eventList.Remove() events = %v, want %v", el.events, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_eventList_Get(t *testing.T) {
	type fields struct {
		events map[uint8]*evenArgList
		lock   *sync.RWMutex
	}
	type args struct {
		event uint8
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string][]any
	}{
		{
			name: "empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 0,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
			},
			want: map[string][]any{},
		},
		{
			name: "not empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 0,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{
									16:  "sixteen",
									22:  "twenty-two",
									500: "five hundred",
									37:  "thirty-seven",
									89:  "eighty-nine",
								},
								sorted: []any{"sixteen", "twenty-two", "thirty-seven", "eighty-nine", "five hundred"},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
			},
			want: map[string][]any{
				"arg_A": {"six", "seven", "forty-two", "ninety-nine", "two hundred"},
				"arg_B": {"sixteen", "twenty-two", "thirty-seven", "eighty-nine", "five hundred"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &eventList{
				events: tt.fields.events,
				lock:   tt.fields.lock,
			}
			if got := el.Get(tt.args.event); !(len(got) == 0 && len(tt.want) == 0) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("eventList.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_eventList_GetArg(t *testing.T) {
	type fields struct {
		events map[uint8]*evenArgList
		lock   *sync.RWMutex
	}
	type args struct {
		event uint8
		arg   string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []any
	}{
		{
			name: "empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_1",
			},
			want: []any{},
		},
		{
			name: "not empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 5,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 5,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_A",
			},
			want: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &eventList{
				events: tt.fields.events,
				lock:   tt.fields.lock,
			}
			if got := el.GetArg(tt.args.event, tt.args.arg); !(len(got) == 0 && len(tt.want) == 0) && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("eventList.GetArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_eventList_Count(t *testing.T) {
	type fields struct {
		events map[uint8]*evenArgList
		lock   *sync.RWMutex
	}
	type args struct {
		event uint8
		arg   string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 0,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 0,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   "arg_1",
			},
			want: 0,
		},
		{
			name: "not empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 0,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 0,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   "arg_A",
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &eventList{
				events: tt.fields.events,
				lock:   tt.fields.lock,
			}
			if got := el.Count(tt.args.event, tt.args.arg); got != tt.want {
				t.Errorf("eventList.Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_eventList_IsEmpty(t *testing.T) {
	type fields struct {
		events map[uint8]*evenArgList
		lock   *sync.RWMutex
	}
	type args struct {
		event uint8
		arg   []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 0,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 0,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 10,
				arg:   []string{"arg_1"},
			},
			want: true,
		},
		{
			name: "not empty",
			fields: fields{
				events: map[uint8]*evenArgList{
					10: {
						args: map[string]*eventArgItemList{
							"arg_1": {
								original: map[int]any{},
								sorted:   []any{},
							},
							"arg_2": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
						},
						counter: 0,
					},
					20: {
						args: map[string]*eventArgItemList{
							"arg_A": {
								original: map[int]any{
									6:   "six",
									42:  "forty-two",
									200: "two hundred",
									7:   "seven",
									99:  "ninety-nine",
								},
								sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
							},
							"arg_B": {
								original: map[int]any{},
								sorted:   []any{},
							},
						},
						counter: 0,
					},
				},
				lock: &sync.RWMutex{},
			},
			args: args{
				event: 20,
				arg:   []string{"arg_A"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			el := &eventList{
				events: tt.fields.events,
				lock:   tt.fields.lock,
			}
			if got := el.IsEmpty(tt.args.event, tt.args.arg...); got != tt.want {
				t.Errorf("eventList.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_newArgList(t *testing.T) {
	tests := []struct {
		name string
		want *evenArgList
	}{
		{
			name: "default",
			want: &evenArgList{
				args:    map[string]*eventArgItemList{},
				counter: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newArgList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_argList_addArg(t *testing.T) {
	type fields struct {
		args    map[string]*eventArgItemList
		counter int
	}
	type args struct {
		arg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   evenArgList
	}{
		{
			name: "add new",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg: "arg_3",
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
					"arg_3": {
						original: map[int]any{},
						sorted:   []any{},
					},
				},
				counter: 5,
			},
		},
		{
			name: "existing",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg: "arg_2",
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &evenArgList{
				args:    tt.fields.args,
				counter: tt.fields.counter,
			}
			al.addArg(tt.args.arg)
			if !reflect.DeepEqual(*al, tt.want) {
				t.Errorf("evenArgList.addArg() evenArgList = %v, want %v", *al, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_argList_addItem(t *testing.T) {
	type fields struct {
		args    map[string]*eventArgItemList
		counter int
	}
	type args struct {
		arg  string
		item any
		key  []int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   evenArgList
	}{
		{
			name: "append to empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg:  "arg_1",
				item: "new item",
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{
							0: "new item",
						},
						sorted: []any{"new item"},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 6,
			},
		},
		{
			name: "append to not empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg:  "arg_2",
				item: "new item",
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
							201: "new item",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred", "new item"},
					},
				},
				counter: 6,
			},
		},
		{
			name: "add to not empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg:  "arg_2",
				item: "new item",
				key:  []int{100},
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
							100: "new item",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "new item", "two hundred"},
					},
				},
				counter: 6,
			},
		},
		{
			name: "replace in not empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg:  "arg_2",
				item: "new item",
				key:  []int{42},
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "new item",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "new item", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &evenArgList{
				args:    tt.fields.args,
				counter: tt.fields.counter,
			}
			al.addItem(tt.args.arg, tt.args.item, tt.args.key...)
			if !reflect.DeepEqual(*al, tt.want) {
				t.Errorf("evenArgList.addItem() evenArgList = %v, want %v", *al, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_argList_remove(t *testing.T) {
	type fields struct {
		args    map[string]*eventArgItemList
		counter int
	}
	type args struct {
		arg  string
		keys []int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   evenArgList
	}{
		{
			name: "remove from empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg:  "arg_1",
				keys: []int{100},
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
		},
		{
			name: "remove existing",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg:  "arg_2",
				keys: []int{200, 99},
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:  "six",
							42: "forty-two",
							7:  "seven",
						},
						sorted: []any{"six", "seven", "forty-two"},
					},
				},
				counter: 3,
			},
		},
		{
			name: "remove some existing",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg:  "arg_2",
				keys: []int{99, 100},
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
						},
						sorted: []any{"six", "seven", "forty-two", "two hundred"},
					},
				},
				counter: 4,
			},
		},
		{
			name: "remove arg",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg: "arg_1",
			},
			want: evenArgList{
				args: map[string]*eventArgItemList{
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &evenArgList{
				args:    tt.fields.args,
				counter: tt.fields.counter,
			}
			al.remove(tt.args.arg, tt.args.keys...)
			if !reflect.DeepEqual(*al, tt.want) {
				t.Errorf("evenArgList.remove() evenArgList = %v, want %v", *al, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_argList_count(t *testing.T) {
	type fields struct {
		args    map[string]*eventArgItemList
		counter int
	}
	type args struct {
		args []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "count all",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{
							69: "syxty-nine",
							55: "fifty-five",
							42: "forty-two",
						},
						sorted: []any{"forty-two", "fifty-five", "syxty-nine"},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 8,
			},
			args: args{
				args: nil,
			},
			want: 8,
		},
		{
			name: "count single",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{
							69: "syxty-nine",
							55: "fifty-five",
							42: "forty-two",
						},
						sorted: []any{"forty-two", "fifty-five", "syxty-nine"},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 8,
			},
			args: args{
				args: []string{"arg_2"},
			},
			want: 5,
		},
		{
			name: "count empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{
							69: "syxty-nine",
							55: "fifty-five",
							42: "forty-two",
						},
						sorted: []any{"forty-two", "fifty-five", "syxty-nine"},
					},
					"arg_2": {
						original: map[int]any{},
						sorted:   []any{},
					},
				},
				counter: 8,
			},
			args: args{
				args: []string{"arg_2"},
			},
			want: 0,
		},
		{
			name: "count non-existing",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{
							69: "syxty-nine",
							55: "fifty-five",
							42: "forty-two",
						},
						sorted: []any{"forty-two", "fifty-five", "syxty-nine"},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 8,
			},
			args: args{
				args: []string{"arg_3"},
			},
			want: 0,
		},
		{
			name: "count multi",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{
							69: "syxty-nine",
							55: "fifty-five",
							42: "forty-two",
						},
						sorted: []any{"forty-two", "fifty-five", "syxty-nine"},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
					"arg_3": {
						original: map[int]any{
							16: "sixteen",
							42: "forty-two",
							21: "twenty-one",
							77: "seventy-seven",
							19: "nineteen",
						},
						sorted: []any{"sixteen", "nineteen", "twenty-one", "forty-two", "seventy-seven"},
					},
				},
				counter: 12,
			},
			args: args{
				args: []string{"arg_3", "arg_1"},
			},
			want: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &evenArgList{
				args:    tt.fields.args,
				counter: tt.fields.counter,
			}
			if got := al.count(tt.args.args...); got != tt.want {
				t.Errorf("evenArgList.count() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_argList_isEmpty(t *testing.T) {
	type fields struct {
		args    map[string]*eventArgItemList
		counter int
	}
	type args struct {
		arg []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg: []string{"arg_1"},
			},
			want: true,
		},
		{
			name: "not empty",
			fields: fields{
				args: map[string]*eventArgItemList{
					"arg_1": {
						original: map[int]any{},
						sorted:   []any{},
					},
					"arg_2": {
						original: map[int]any{
							6:   "six",
							42:  "forty-two",
							200: "two hundred",
							7:   "seven",
							99:  "ninety-nine",
						},
						sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
					},
				},
				counter: 5,
			},
			args: args{
				arg: []string{"arg_2"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := &evenArgList{
				args:    tt.fields.args,
				counter: tt.fields.counter,
			}
			if got := al.isEmpty(tt.args.arg...); got != tt.want {
				t.Errorf("event.isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ------------------------------------------------------------------------

func Test_newArgument(t *testing.T) {
	tests := []struct {
		name string
		want *eventArgItemList
	}{
		{
			name: "default",
			want: &eventArgItemList{
				original: map[int]any{},
				sorted:   []any{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newItemList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newArgument() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_itemList_set(t *testing.T) {
	type fields struct {
		original map[int]any
		sorted   []any
	}
	type args struct {
		key  int
		item any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantArg eventArgItemList
	}{
		{
			name: "new",
			fields: fields{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					99:  "ninety-nine",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "ninety-nine", "two hundred"},
			},
			args: args{
				key:  42,
				item: "forty-two",
			},
			want: true,
			wantArg: eventArgItemList{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					99:  "ninety-nine",
					200: "two hundred",
					42:  "forty-two",
				},
				sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
			},
		},
		{
			name: "replace",
			fields: fields{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					99:  "ninety-nine",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "ninety-nine", "two hundred"},
			},
			args: args{
				key:  99,
				item: "ninety-nine and a half",
			},
			want: false,
			wantArg: eventArgItemList{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					99:  "ninety-nine and a half",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "ninety-nine and a half", "two hundred"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			il := &eventArgItemList{
				original: tt.fields.original,
				sorted:   tt.fields.sorted,
			}
			if got := il.set(tt.args.key, tt.args.item); got != tt.want {
				t.Errorf("eventArgItemList.setItem() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(*il, tt.wantArg) {
				t.Errorf("eventArgItemList.setItem() eventArgItemList = %v, want %v", *il, tt.wantArg)
			}
		})
	}
}

// --------------------------------

func Test_itemList_append(t *testing.T) {
	type fields struct {
		original map[int]any
		sorted   []any
	}
	type args struct {
		item any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantArg eventArgItemList
	}{
		{
			name: "append",
			fields: fields{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					99:  "ninety-nine",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "ninety-nine", "two hundred"},
			},
			args: args{
				item: "forty-two",
			},
			want: true,
			wantArg: eventArgItemList{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					99:  "ninety-nine",
					200: "two hundred",
					201: "forty-two",
				},
				sorted: []any{"six", "seven", "ninety-nine", "two hundred", "forty-two"},
			},
		},
		{
			name: "int overflow",
			fields: fields{
				original: map[int]any{
					6:           "six",
					7:           "seven",
					99:          "ninety-nine",
					math.MaxInt: "max",
				},
				sorted: []any{"six", "seven", "ninety-nine", "max"},
			},
			args: args{
				item: "forty-two",
			},
			want: false,
			wantArg: eventArgItemList{
				original: map[int]any{
					6:           "six",
					7:           "seven",
					99:          "ninety-nine",
					math.MaxInt: "max",
				},
				sorted: []any{"six", "seven", "ninety-nine", "max"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			il := &eventArgItemList{
				original: tt.fields.original,
				sorted:   tt.fields.sorted,
			}
			if got := il.append(tt.args.item); got != tt.want {
				t.Errorf("eventArgItemList.appendItem() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(*il, tt.wantArg) {
				t.Errorf("eventArgItemList.appendItem() eventArgItemList = %v, want %v", *il, tt.wantArg)
			}
		})
	}
}

// --------------------------------

func Test_itemList_remove(t *testing.T) {
	type fields struct {
		original map[int]any
		sorted   []any
	}
	type args struct {
		key int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantArg eventArgItemList
	}{
		{
			name: "existing",
			fields: fields{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					42:  "forty-two",
					99:  "ninety-nine",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
			},
			args: args{
				key: 42,
			},
			want: true,
			wantArg: eventArgItemList{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					99:  "ninety-nine",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "ninety-nine", "two hundred"},
			},
		},
		{
			name: "missing",
			fields: fields{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					42:  "forty-two",
					99:  "ninety-nine",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
			},
			args: args{
				key: 69,
			},
			want: false,
			wantArg: eventArgItemList{
				original: map[int]any{
					6:   "six",
					7:   "seven",
					42:  "forty-two",
					99:  "ninety-nine",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			il := &eventArgItemList{
				original: tt.fields.original,
				sorted:   tt.fields.sorted,
			}
			if got := il.remove(tt.args.key); got != tt.want {
				t.Errorf("eventArgItemList.removeItem() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(*il, tt.wantArg) {
				t.Errorf("eventArgItemList.removeItem() eventArgItemList = %v, want %v", *il, tt.wantArg)
			}
		})
	}
}

// --------------------------------

func Test_itemList_sort(t *testing.T) {
	type fields struct {
		original map[int]any
		sorted   []any
	}
	tests := []struct {
		name   string
		fields fields
		want   []any
	}{
		{
			name: "empty",
			fields: fields{
				original: map[int]any{},
				sorted:   []any{},
			},
			want: []any{},
		},
		{
			name: "not sorted",
			fields: fields{
				original: map[int]any{
					6:   "six",
					99:  "ninety-nine",
					200: "two hundred",
					7:   "seven",
				},
				sorted: nil,
			},
			want: []any{"six", "seven", "ninety-nine", "two hundred"},
		},
		{
			name: "marked sorted incorrectly",
			fields: fields{
				original: map[int]any{
					6:   "six",
					99:  "ninety-nine",
					200: "two hundred",
					7:   "seven",
				},
				sorted: []any{"six", "ninety-nine", "two hundred", "seven"},
			},
			want: []any{"six", "seven", "ninety-nine", "two hundred"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			il := &eventArgItemList{
				original: tt.fields.original,
				sorted:   tt.fields.sorted,
			}
			il.sort()
			if !reflect.DeepEqual(il.sorted, tt.want) {
				t.Errorf("eventArgItemList.setItem() = %v, want %v", il.sorted, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_itemList_count(t *testing.T) {
	type fields struct {
		original map[int]any
		sorted   []any
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "empty",
			fields: fields{
				original: map[int]any{},
				sorted:   []any{},
			},
			want: 0,
		},
		{
			name: "not empty",
			fields: fields{
				original: map[int]any{
					6:   "six",
					99:  "ninety-nine",
					42:  "forty-two",
					7:   "seven",
					200: "two hundred",
				},
				sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			il := &eventArgItemList{
				original: tt.fields.original,
				sorted:   tt.fields.sorted,
			}
			if got := il.count(); got != tt.want {
				t.Errorf("eventArgItemList.count() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --------------------------------

func Test_itemList_isEmpty(t *testing.T) {
	type fields struct {
		original map[int]any
		sorted   []any
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "empty",
			fields: fields{
				original: map[int]any{},
				sorted:   []any{},
			},
			want: true,
		},
		{
			name: "not empty",
			fields: fields{
				original: map[int]any{
					6:   "six",
					42:  "forty-two",
					200: "two hundred",
					7:   "seven",
					99:  "ninety-nine",
				},
				sorted: []any{"six", "seven", "forty-two", "ninety-nine", "two hundred"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			il := &eventArgItemList{
				original: tt.fields.original,
				sorted:   tt.fields.sorted,
			}
			if got := il.isEmpty(); got != tt.want {
				t.Errorf("eventArgItemList.isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
