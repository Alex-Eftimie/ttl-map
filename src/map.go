// Credit OneOfOne @ stackoverflow
package ttlmap

import (
	"container/heap"
	"encoding/json"
	"sync"
	"time"
)

type item struct {
	Value interface{}
	// LastAccess int64
	HeapNode *ttl
}

type ttl struct {
	*time.Time
	Key string
}
type TTLMap struct {
	filled bool
	m      map[string]*item
	l      sync.Mutex
	maxTTL int64
	heap   *TTLHeap
}

func New(ln int, maxTTL int64) (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*item, ln), maxTTL: maxTTL}
	m.runcleaner()
	m.filled = true
	m.heap = &TTLHeap{}
	return m
}

func (m *TTLMap) IsNil() bool {
	return !m.filled
}
func (m *TTLMap) runcleaner() {
	go func() {
		for now := range time.Tick(10 * time.Second) {
			m.l.Lock()

			dropoff := now.Unix() - m.maxTTL

			for m.heap.Len() > 0 && (*m.heap)[0].Time.Unix() < dropoff {
				e := heap.Pop(m.heap).(*ttl)

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
		ttl := &ttl{Time: &now, Key: k}
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

	// m = New(0, s.T)
	m.m = mymap
	m.maxTTL = s.T
	m.runcleaner()
	m.filled = true

	return nil
}
