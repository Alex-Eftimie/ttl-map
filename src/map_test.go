package ttlmap

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	m := New(0, 60)
	if m == nil {
		t.Errorf("New(0, 60) = nil; want pointer")
	}
}

func TestPutGet(t *testing.T) {
	m := New(0, 60)
	key := "somekey"
	value := "SomeStringValue"
	m.Put(key, value)

	res := m.Get(key).(string)

	if res != value {
		t.Errorf("Put(%s, %s); Get(%s)=%s, Expected: %s", key, value, key, res, value)
	}
}

func TestTtlExpired(t *testing.T) {
	m := New(0, 1)
	key := "somekey"
	value := "SomeStringValue"
	m.Put(key, value)

	time.Sleep(3 * time.Second)

	res := m.Get(key)

	if res != nil {
		str := res.(string)
		t.Errorf("Put(%s, %s); [Expiration] Get(%s)=%s, Expected: nil", key, value, key, str)
	}
}

func TestTtlNotExpired(t *testing.T) {
	m := New(0, 1)
	key := "somekey"
	value := "SomeStringValue"
	m.Put(key, value)

	res := m.Get(key)

	if res == nil {
		str := res.(string)
		t.Errorf("Put(%s, %s); Get(%s)=%s, Expected: %s", key, value, key, str, value)
	}
}

func TestJson(t *testing.T) {
	m := New(0, 1)
	key := "somekey"
	value := "SomeStringValue"
	m.Put(key, value)
	jsb, _ := json.MarshalIndent(m, "", "    ")

	n := &TTLMap{}
	err := json.Unmarshal(jsb, n)
	if err != nil {
		t.Error(err.Error())
	}

	res := n.Get(key)
	if res != value {
		if res == nil {
			t.Errorf("Put(%s, %s); [Marshal, Unmarshal] Get(%s)=nil, Expected: %s", key, value, key, value)
			return
		}
		str := res.(string)
		t.Errorf("Put(%s, %s); [Marshal, Unmarshal] Get(%s)=%s, Expected: %s", key, value, key, str, value)
	}

}
