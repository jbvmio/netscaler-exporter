package main

import (
	"strings"
	"sync"
)

// FlipBit controls no/go operations.
type FlipBit struct {
	bit  bool
	lock sync.Mutex
}

func (c *FlipBit) good() bool {
	var ok bool
	c.lock.Lock()
	if !c.bit {
		c.bit = true
		ok = true
	}
	c.lock.Unlock()
	return ok
}

func (c *FlipBit) flip() {
	c.lock.Lock()
	c.bit = !c.bit
	c.lock.Unlock()
}

// MiscMap for general control of data:
type MiscMap struct {
	data map[string]interface{}
	lock sync.Mutex
}

// Update inserts or updates data within the MiscMap:
func (m *MiscMap) Update(key string, v interface{}) {
	m.lock.Lock()
	m.data[key] = v
	m.lock.Unlock()
}

// Remove deletes data within the MiscMap by key value:
func (m *MiscMap) Remove(key string) {
	m.lock.Lock()
	delete(m.data, key)
	m.lock.Unlock()
}

// Get retrieves data within the MiscMap by key value:
func (m *MiscMap) Get(key string) interface{} {
	m.lock.Lock()
	v := m.data[key]
	m.lock.Unlock()
	return v
}

// Exists returns true if the given key exists within the MiscMap, false otherwise:
func (m *MiscMap) Exists(key string) bool {
	m.lock.Lock()
	_, v := m.data[key]
	m.lock.Unlock()
	return v
}

func nsInstance(url string) string {
	n := strings.TrimLeft(url, "https://")
	n = strings.TrimLeft(n, "http://")
	n = strings.Trim(n, " /")
	shortname := strings.Split(n, `.`)
	if len(shortname) > 0 {
		return shortname[0]
	}
	return n
}
