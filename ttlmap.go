package ttlmap

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type nocopy uintptr
type wrappedV[V any] struct {
	v V
	t int64
}

type TtlMap[K comparable, V any] struct {
	_             [0]func() // cant ==
	nocopy        nocopy
	value         map[K]wrappedV[V]
	rw            sync.RWMutex
	ttl           int64
	trigger       chan struct{}
	cleanF        func(*TtlMap[K, V])
	finalizer     chan struct{}
	flushInterval time.Duration
}

func NewTTLMap[K comparable, V any](options ...TTLMapOption[K, V]) (res *TtlMap[K, V]) {
	res = &TtlMap[K, V]{
		value:     make(map[K]wrappedV[V]),
		rw:        sync.RWMutex{},
		ttl:       int64(3 * time.Minute),
		trigger:   make(chan struct{}, 0),
		finalizer: make(chan struct{}, 0),
	}
	res.flushInterval = time.Duration(res.ttl / 3)
	for _, option := range options {
		option(res)
	}
	res.check()
	res.cleanF = func(tm *TtlMap[K, V]) {
		tm.rw.Lock()
		defer tm.rw.Unlock()
		for k, v := range tm.value {
			if v.t+res.ttl <= time.Now().UnixNano() {
				delete(tm.value, k)
			}
		}
	}
	go func(inner *TtlMap[K, V]) {
		timer := time.NewTimer(inner.flushInterval)
		pause := time.Millisecond * 10
		for {
			select {
			case <-res.finalizer:
				close(inner.trigger)
				close(inner.finalizer)
				*inner = TtlMap[K, V]{}
				return
			case <-timer.C:
				inner.cleanF(inner)
			case <-inner.trigger:
				inner.cleanF(inner)
			default:
				time.Sleep(pause)
			}
		}
	}(res)
	return
}

type TTLMapOption[K comparable, V any] func(*TtlMap[K, V])

func WithTTL[K comparable, V any](duration time.Duration) TTLMapOption[K, V] {
	return func(t *TtlMap[K, V]) {
		t.ttl = int64(duration)
	}
}
func WithFlushInterval[K comparable, V any](interval time.Duration) TTLMapOption[K, V] {
	return func(t *TtlMap[K, V]) {
		t.flushInterval = interval
	}
}

func (tm *TtlMap[K, V]) Set(key K, value V) {
	tm.check()
	tm.rw.Lock()
	defer tm.rw.Unlock()
	tm.value[key] = wrappedV[V]{
		v: value,
		t: time.Now().UnixNano(),
	}
}
func (tm *TtlMap[K, V]) Flush() {
	tm.check()
	tm.trigger <- struct{}{}
}
func (tm *TtlMap[K, V]) Drain() {
	tm.check()
	tm.finalizer <- struct{}{}
}
func (tm *TtlMap[K, V]) check() {
	if uintptr(tm.nocopy) != uintptr(unsafe.Pointer(tm)) && !atomic.CompareAndSwapUintptr((*uintptr)(&tm.nocopy), 0, uintptr(unsafe.Pointer(tm))) && uintptr(tm.nocopy) != uintptr(unsafe.Pointer(tm)) {
		panic(any("object has copied"))
	}
}
func (tm *TtlMap[K, V]) SetWithExpire(key K, value V, ttl time.Duration) {
	tm.check()
	tm.rw.Lock()
	defer tm.rw.Unlock()
	tm.value[key] = wrappedV[V]{
		v: value,
		t: time.Now().UnixNano() + int64(ttl) - tm.ttl,
	}
}

func (tm *TtlMap[K, V]) Get(key K) (value V) {
	tm.check()
	tm.rw.RLock()
	defer tm.rw.RUnlock()
	var v wrappedV[V]
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
func (tm *TtlMap[K, V]) TryGet(key K) (value V, ok bool) {
	tm.check()
	tm.rw.RLock()
	defer tm.rw.RUnlock()
	var v wrappedV[V]
	var vk bool
	v, vk = tm.value[key]
	if v.t+tm.ttl <= time.Now().UnixNano() {
		delete(tm.value, key)
		return
	}
	return v.v, vk
}
func (tm *TtlMap[K, V]) TryDelete(key K) (value V, ok bool) {
	tm.check()
	tm.rw.Lock()
	defer tm.rw.Unlock()
	var v wrappedV[V]
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
func (tm *TtlMap[K, V]) Delete(key K) {
	tm.check()
	tm.rw.Lock()
	defer tm.rw.Unlock()
	delete(tm.value, key)
}
