// Credit OneOfOne @ stackoverflow
package ttlmap

import (
	"container/heap"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"
)

// ErrAlreadyRunning is returned when you try to start the cleaner twice
var ErrAlreadyRunning error = errors.New("Cleaner already running")

// ErrEmptyTime is returned when you try to start the cleaner with an empty time
var ErrEmptyTime error = errors.New("Empty time")

type item struct {
	Value interface{}
	// LastAccess int64
	HeapNode *TTLItem
}

type TTLItem struct {
	*time.Time
	Key string
}
type TTLMap struct {
	filled  bool
	m       map[string]*item
	l       sync.Mutex
	maxTTL  int64
	heap    *TTLHeap
	stopped bool
	ticker  *time.Ticker
}

func New(ln int, maxTTL int64) (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*item, ln), maxTTL: maxTTL}
	m.stopped = true
	err := m.RunCleaner(time.Duration(maxTTL) * time.Second / 2)
	log.Println(err)
	m.filled = true
	m.heap = &TTLHeap{}
	return m
}

func (m *TTLMap) IsNil() bool {
	return !m.filled
}
func (m *TTLMap) Stop() {
	m.stopped = true
	if m.ticker == nil {
		return
	}

	m.ticker.Stop()
}
func (m *TTLMap) RunCleaner(t time.Duration) error {

	if !m.stopped {
		return ErrAlreadyRunning
	}

	if t == 0 {
		return ErrEmptyTime
	}

	tk := time.NewTicker(t)
	m.ticker = tk
	m.stopped = false

	go func() {
		for now := range tk.C {
			if m.stopped {
				return
			}
			m.l.Lock()

			dropoff := now.Unix() - m.maxTTL

			for m.heap.Len() > 0 && (*m.heap)[0].Time.Unix() < dropoff {
				e := heap.Pop(m.heap).(*TTLItem)

				// utils.Debug(9999, "Delete: ", e.Key, e.Time)
				// for k, v := range m.m {
				// if now.Unix()-v.LastAccess > int64(m.maxTTL) {
				// if now.Unix()-v.HeapNode.Time.Unix() > int64(m.maxTTL) {
				delete(m.m, e.Key)
				// }
			}
			m.l.Unlock()
		}
	}()
	return nil
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Put(k string, v interface{}) {
	now := time.Now()
	m.l.Lock()
	it, ok := m.m[k]
	if !ok || (ok && v != it.Value) {
		delete(m.m, k)
		ttl := &TTLItem{Time: &now, Key: k}
		it = &item{Value: v, HeapNode: ttl}
		heap.Push(m.heap, ttl)
		m.m[k] = it
	} else {
		it.HeapNode.Time = &now
	}
	heap.Init(m.heap)
	// it.LastAccess = time.Now().Unix()
	m.l.Unlock()
}

func (m *TTLMap) Get(k string) (v interface{}) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = it.Value
		now := time.Now()
		// if it.HeapNode.Time.Unix() < now.Unix() {
		// // expired item
		// no need to search the heap too, it'll just slow things down
		// delete(m.m, k)
		// 	return nil
		// }
		it.HeapNode.Time = &now
		heap.Init(m.heap)
		// it.LastAccess = time.Now().Unix()
	}
	m.l.Unlock()
	return v

}

func (m *TTLMap) MarshalJSON() ([]byte, error) {
	s := &struct {
		M map[string]*item
		T int64
	}{
		M: m.m,
		T: m.maxTTL,
	}
	return json.Marshal(s)
}

func (m *TTLMap) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "{}" {
		return nil
	}
	mymap := make(map[string]*item)
	s := &struct {
		M map[string]*item
		T int64
	}{
		M: mymap,
		T: 0,
	}
	err := json.Unmarshal(b, s)
	if err != nil {
		return err
	}

	// populate the heap
	m.heap = &TTLHeap{}
	for k, v := range mymap {
		v.HeapNode = &TTLItem{Time: v.HeapNode.Time, Key: k}
		heap.Push(m.heap, v.HeapNode)
	}

	// m = New(0, s.T)
	m.m = mymap
	m.maxTTL = s.T
	m.stopped = true

	err = m.RunCleaner(time.Duration(s.T) * time.Second / 2)
	if err != nil {
		return err
	}
	m.filled = true

	return nil
}

func (m *TTLItem) UnmarshalJSON(b []byte) error {

	t := time.Time{}

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	m.Time = &t

	return nil
}
