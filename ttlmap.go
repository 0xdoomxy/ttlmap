package ttlmap

import (
	"sync"
	"time"
)

type WrappedV[V any] struct {
	v V
	t int64
}

type ttlMap[K comparable, V any] struct {
	value         map[K]WrappedV[V]
	rw            sync.RWMutex
	ttl           int64
	trigger       chan struct{}
	cleanF        func(*ttlMap[K, V])
	finalizer     chan struct{}
	flushInterval time.Duration
}

func NewTTLMap[K comparable, V any](options ...TTLMapOption[K, V]) (res *ttlMap[K, V]) {
	res = &ttlMap[K, V]{
		value:     make(map[K]WrappedV[V]),
		rw:        sync.RWMutex{},
		ttl:       int64(3 * time.Minute),
		trigger:   make(chan struct{}),
		finalizer: make(chan struct{}, 0),
	}
	res.flushInterval = time.Duration(res.ttl / 3)
	for _, option := range options {
		option(res)
	}
	res.cleanF = func(tm *ttlMap[K, V]) {
		tm.rw.Lock()
		defer tm.rw.Unlock()
		for k, v := range tm.value {
			if v.t+res.ttl <= time.Now().UnixNano() {
				delete(tm.value, k)
			}
		}
	}
	go func(inner *ttlMap[K, V]) {
		timer := time.NewTimer(inner.flushInterval)
		maxPause := time.Duration(inner.flushInterval / 3)
		pause := time.Millisecond * 10
		mutex := sync.Mutex{}
		if pause > maxPause {
			pause = maxPause
		}
		for {
			select {
			case <-res.finalizer:
				*inner = ttlMap[K, V]{}
				return
			case <-timer.C:
				if mutex.TryLock() {
					inner.cleanF(inner)
				}
			case <-inner.trigger:
				inner.cleanF(inner)
			default:
				time.Sleep(pause)
			}
		}
	}(res)
	return
}

type TTLMapOption[K comparable, V any] func(*ttlMap[K, V])

func WithTTL[K comparable, V any](duration time.Duration) TTLMapOption[K, V] {
	return func(t *ttlMap[K, V]) {
		t.ttl = int64(duration)
	}
}
func WithFlushInterval[K comparable, V any](interval time.Duration) TTLMapOption[K, V] {
	return func(t *ttlMap[K, V]) {
		t.flushInterval = interval
	}
}

func (tm *ttlMap[K, V]) Set(key K, value V) {
	tm.rw.Lock()
	defer tm.rw.Unlock()
	tm.value[key] = WrappedV[V]{
		v: value,
		t: time.Now().UnixNano(),
	}
}
func (tm *ttlMap[K, V]) Flush() {
	tm.trigger <- struct{}{}
}
func (tm *ttlMap[K, V]) Drain() {
	tm.finalizer <- struct{}{}
}

func (tm *ttlMap[K, V]) SetWithExpire(key K, value V, ttl time.Duration) {
	tm.rw.Lock()
	defer tm.rw.Unlock()
	tm.value[key] = WrappedV[V]{
		v: value,
		t: time.Now().UnixNano() + int64(ttl),
	}
}

func (tm *ttlMap[K, V]) Get(key K) (value V) {
	tm.rw.RLock()
	defer tm.rw.RUnlock()
	var v WrappedV[V]
	var vk bool
	v, vk = tm.value[key]
	if !vk {
		return
	}
	if v.t+tm.ttl <= time.Now().UnixNano() {
		delete(tm.value, key)
		return
	}
	return v.v
}
func (tm *ttlMap[K, V]) TryGet(key K) (value V, ok bool) {
	tm.rw.RLock()
	defer tm.rw.RUnlock()
	var v WrappedV[V]
	var vk bool
	v, vk = tm.value[key]
	if v.t+tm.ttl <= time.Now().UnixNano() {
		delete(tm.value, key)
		return
	}
	return v.v, vk
}
func (tm *ttlMap[K, V]) TryDelete(key K) (value V, ok bool) {
	tm.rw.Lock()
	defer tm.rw.Unlock()
	var v WrappedV[V]
	var vk bool
	v, vk = tm.value[key]
	if !vk {
		return
	}
	defer delete(tm.value, key)
	if v.t+tm.ttl <= time.Now().UnixNano() {
		return
	}
	return v.v, true
}
func (tm *ttlMap[K, V]) Delete(key K) {
	tm.rw.Lock()
	defer tm.rw.Unlock()
	delete(tm.value, key)
}
