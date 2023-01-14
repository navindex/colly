package colly

import (
	"math"
	"sort"
	"sync"
)

// ------------------------------------------------------------------------

// The eventList structure is an ordered list of items, grouped by events and their arguments.
// It is responsible for locking.
type eventList struct {
	events map[uint8]*evenArgList
	lock   *sync.RWMutex
}

// The evenArgList structure has the argument list of all events.
// It is responsible for item counting.
type evenArgList struct {
	args    map[string]*eventArgItemList
	counter int
}

// The eventArgItemList structure has the item list of an argument.
// It is responsible for item sorting.
type eventArgItemList struct {
	original map[int]any
	sorted   []any
}

// ------------------------------------------------------------------------

// NewEventArgList returns a pointer to a newly created event argument list.
func NewEventList() *eventList {
	return &eventList{
		events: map[uint8]*evenArgList{},
		lock:   &sync.RWMutex{},
	}
}

// ------------------------------------------------------------------------

// Add inserts or appends a new event argument item into the list.
// If index is not empty, the new item  will be inserted to position identified by index[0].
// If an item exists with the same index, the old item will be overwritten.
// If no index was given, the new item will be attached to the end of the list.
func (el *eventList) Add(event uint8, arg string, item any, index ...int) {
	el.lock.Lock()
	defer el.lock.Unlock()

	// Create event if missing
	if _, present := el.events[event]; !present {
		el.events[event] = newArgList()
	}

	el.events[event].addItem(arg, item, index...)
}

// ------------------------------------------------------------------------

// Remove removes a number of event argument items or all items
// attached to the event argument if index is nil.
func (el *eventList) Remove(event uint8, arg string, index ...int) {
	el.lock.Lock()
	defer el.lock.Unlock()

	el.events[event].remove(arg, index...)
}

// ------------------------------------------------------------------------

// Get returns a sorted slice of all event items, mapped by arguments.
func (el *eventList) Get(event uint8) map[string][]any {
	el.lock.RLock()
	defer el.lock.RUnlock()

	// Nothing to return if the event doesn't exist
	al, present := el.events[event]
	if !present {
		return nil
	}

	return al.getAll()
}

// ------------------------------------------------------------------------

// Get returns a sorted slice of event argument items.
func (el *eventList) GetArg(event uint8, arg string) []any {
	el.lock.RLock()
	defer el.lock.RUnlock()

	// Nothing to return if the event doesn't exist
	al, present := el.events[event]
	if !present {
		return nil
	}

	return al.getArg(arg)
}

// ------------------------------------------------------------------------

// Count returns the number of items attached to an event or argument.
func (el *eventList) Count(event uint8, args ...string) (count int) {
	el.lock.RLock()
	defer el.lock.RUnlock()

	if al, present := el.events[event]; present {
		count = al.count(args...)
	}

	return count
}

// ------------------------------------------------------------------------

// IsEmpty returns true if no items were attached to the event or argument.
func (el *eventList) IsEmpty(event uint8, arg ...string) bool {
	el.lock.RLock()
	defer el.lock.RUnlock()

	al, present := el.events[event]

	return !present || al.isEmpty(arg...)
}

// ------------------------------------------------------------------------

func newArgList() *evenArgList {
	return &evenArgList{
		args:    map[string]*eventArgItemList{},
		counter: 0,
	}
}

// --------------------------------

func (al *evenArgList) addArg(arg string) {
	if _, present := al.args[arg]; !present {
		al.args[arg] = newItemList()
	}
}

// --------------------------------

func (al *evenArgList) addItem(arg string, item any, key ...int) {
	var incr bool

	// Create argument if missing
	al.addArg(arg)

	// If key exists, set the item, otherwise append it to the end
	if len(key) > 0 {
		incr = al.args[arg].set(key[0], item)
	} else {
		incr = al.args[arg].append(item)
	}

	if incr {
		al.counter++
	}
}

// --------------------------------

func (al *evenArgList) remove(arg string, keys ...int) {
	// Nothing to remove if the argument doesn't exist
	if _, present := al.args[arg]; !present {
		return
	}

	// Remove the whole argument if no key given
	if len(keys) == 0 {
		al.counter -= al.args[arg].count()
		delete(al.args, arg)
	}

	for _, key := range keys {
		// Remove the item and set the counter
		if ok := al.args[arg].remove(key); ok {
			al.counter--
		}
	}
}

// --------------------------------

func (al *evenArgList) getArg(arg string) []any {
	if al.isEmpty(arg) {
		return nil
	}

	return al.args[arg].sorted
}

// --------------------------------

func (al *evenArgList) getAll() map[string][]any {
	if al.isEmpty() {
		return nil
	}

	var items = map[string][]any{}

	for arg, il := range al.args {
		if !il.isEmpty() {
			items[arg] = il.sorted
		}
	}

	return items
}

// --------------------------------

func (al *evenArgList) count(args ...string) (count int) {
	if len(args) == 0 {
		return al.counter
	}

	for _, arg := range args {
		if a, present := al.args[arg]; present {
			count += a.count()
		}
	}

	return count
}

// --------------------------------

func (al *evenArgList) isEmpty(args ...string) bool {
	if len(args) == 0 {
		return al.counter == 0
	}

	for _, arg := range args {
		if a, present := al.args[arg]; present && !a.isEmpty() {
			return false
		}
	}

	return true
}

// ------------------------------------------------------------------------

func newItemList() *eventArgItemList {
	return &eventArgItemList{
		original: map[int]any{},
		sorted:   []any{},
	}
}

// --------------------------------

func (il *eventArgItemList) set(key int, item any) (new bool) {
	var present bool = false

	if il.original == nil {
		il.original = map[int]any{}
	} else {
		_, present = il.original[key]
	}

	il.original[key] = item
	il.sort()

	return !present
}

// --------------------------------

func (il *eventArgItemList) append(item any) (ok bool) {
	if _, present := il.original[math.MaxInt]; present {
		return false
	}

	var key int = 0

	if il.original == nil {
		il.original = map[int]any{}
	}

	if len(il.original) > 0 {
		key = math.MinInt
		for k := range il.original {
			if key <= k {
				key = k + 1
			}
		}
	}

	il.original[key] = item
	il.sorted = append(il.sorted, item)

	return true
}

// --------------------------------

func (il *eventArgItemList) remove(key int) (ok bool) {
	if len(il.original) == 0 {
		return false
	}

	if _, present := il.original[key]; !present {
		return false
	}

	delete(il.original, key)
	il.sort()

	return true
}

// --------------------------------

func (il *eventArgItemList) sort() {
	il.sorted = []any{}

	if len(il.original) == 0 {
		return
	}

	// Collect and sort the keys
	keys := []int{}
	for k := range il.original {
		keys = append(keys, int(k))
	}

	if len(keys) > 1 {
		sort.Ints(keys)
	}

	// Populate the slice with the sorted items
	for _, k := range keys {
		il.sorted = append(il.sorted, il.original[k])
	}
}

// --------------------------------

func (il *eventArgItemList) count() int {
	return len(il.original)
}

// --------------------------------

func (il *eventArgItemList) isEmpty() bool {
	return len(il.original) == 0
}
