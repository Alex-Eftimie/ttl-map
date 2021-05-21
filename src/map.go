// Credit OneOfOne @ stackoverflow
package ttlmap

import (
	"encoding/json"
	"sync"
	"time"
)

type item struct {
	Value      interface{}
	LastAccess int64
}

type TTLMap struct {
	m      map[string]*item
	l      sync.Mutex
	maxTTL int
}

func New(ln int, maxTTL int) (m *TTLMap) {
	m = &TTLMap{m: make(map[string]*item, ln), maxTTL: maxTTL}
	m.runcleaner()
	return m
}

func (m *TTLMap) runcleaner() {
	go func() {
		for now := range time.Tick(time.Second) {
			m.l.Lock()
			for k, v := range m.m {
				if now.Unix()-v.LastAccess > int64(m.maxTTL) {
					delete(m.m, k)
				}
			}
			m.l.Unlock()
		}
	}()
}

func (m *TTLMap) Len() int {
	return len(m.m)
}

func (m *TTLMap) Put(k string, v interface{}) {
	m.l.Lock()
	it, ok := m.m[k]
	if !ok || (ok && v != it.Value) {
		delete(m.m, k)
		it = &item{Value: v}
		m.m[k] = it
	}
	it.LastAccess = time.Now().Unix()
	m.l.Unlock()
}

func (m *TTLMap) Get(k string) (v interface{}) {
	m.l.Lock()
	if it, ok := m.m[k]; ok {
		v = it.Value
		it.LastAccess = time.Now().Unix()
	}
	m.l.Unlock()
	return

}

func (m *TTLMap) MarshalJSON() ([]byte, error) {
	s := &struct {
		M map[string]*item
		T int
	}{
		M: m.m,
		T: m.maxTTL,
	}
	return json.Marshal(s)
}

func (m *TTLMap) UnmarshalJSON(b []byte) error {
	mymap := make(map[string]*item)
	s := &struct {
		M map[string]*item
		T int
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

	return nil
}
