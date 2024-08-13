package ttlmap_test

import (
	"github.com/0xdoomxy/ttlmap"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestTTLMapInsert(t *testing.T) {
	var ttl = time.Second * 10
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	tm.Set("key", "value")
	assert.Equal(t, "value", tm.Get("key"))
}
func TestTTLMapExpire(t *testing.T) {
	var ttl = time.Second * 1
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	tm.Set("key", "value")
	time.Sleep(time.Second)
	assert.Equal(t, "", tm.Get("key"))
}
func TestTTLMapDelete(t *testing.T) {
	var ttl = time.Second * 1
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	tm.Set("key", "value")
	tm.Delete("key")
	assert.Equal(t, "", tm.Get("key"))
}

func TestTTLMapTryDelete(t *testing.T) {
	var ttl = time.Second * 1
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	tm.Set("key", "value")
	value, ok := tm.TryDelete("key")
	assert.Equal(t, "value", value)
	assert.Equal(t, true, ok)
}

func TestTTLMapTryDeleteExpireValue(t *testing.T) {
	var ttl = time.Second * 1
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	tm.Set("key", "value")
	time.Sleep(time.Second)
	value, ok := tm.TryDelete("key")
	assert.Equal(t, "", value)
	assert.Equal(t, false, ok)
}
func TestMapTryDeleteThenGet(t *testing.T) {
	var ttl = time.Second * 1
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	tm.Set("key", "value")
	tm.TryDelete("key")
	assert.Equal(t, "", tm.Get("key"))
}

func TestMapTryGetThenGet(t *testing.T) {
	var ttl = time.Second * 1
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	tm.Set("key", "value")
	val, ok := tm.TryGet("key")
	assert.Equal(t, "value", val)
	assert.Equal(t, true, ok)
	val = tm.Get("key")
	assert.Equal(t, "value", val)
}

func BenchmarkMapInsert(b *testing.B) {
	var m = make(map[string]string)
	for i := 0; i < b.N; i++ {
		m[strconv.Itoa(i)] = strconv.Itoa(i)
	}
}
func BenchmarkMapDelete(b *testing.B) {
	var m = make(map[string]string)
	for i := 0; i < b.N; i++ {
		m[strconv.Itoa(i)] = strconv.Itoa(i)
	}
	for i := 0; i < b.N; i++ {
		delete(m, strconv.Itoa(i))
	}
}
func BenchmarkMapFind(b *testing.B) {
	var m = make(map[string]string)
	for i := 0; i < b.N; i++ {
		m[strconv.Itoa(i)] = strconv.Itoa(i)
	}
}
func BenchmarkTTlMapInsert(b *testing.B) {
	var ttl = time.Second * 10
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	for i := 0; i < b.N; i++ {
		tm.Set(strconv.Itoa(i), strconv.Itoa(i))
	}
}
func BenchmarkTTlMapDelete(b *testing.B) {
	var ttl = time.Second * 10
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	for i := 0; i < b.N; i++ {
		tm.Set(strconv.Itoa(i), strconv.Itoa(i))
	}
	for i := 0; i < b.N; i++ {
		tm.Delete(strconv.Itoa(i))
	}
}
func BenchmarkTTlMapFind(b *testing.B) {
	var ttl = time.Second * 10
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl), ttlmap.WithFlushInterval[string, string](1*time.Millisecond))
	for i := 0; i < b.N; i++ {
		tm.Set(strconv.Itoa(i), strconv.Itoa(i))
	}
	for i := 0; i < b.N; i++ {
		tm.Get(strconv.Itoa(i))
	}
}
func BenchmarkTTlMapTryDelete(b *testing.B) {
	var ttl = time.Second * 10
	tm := ttlmap.NewTTLMap[string, string](ttlmap.WithTTL[string, string](ttl))
	for i := 0; i < b.N; i++ {
		tm.Set(strconv.Itoa(i), strconv.Itoa(i))
	}
	for i := 0; i < b.N; i++ {
		tm.TryDelete(strconv.Itoa(i))
	}
}
