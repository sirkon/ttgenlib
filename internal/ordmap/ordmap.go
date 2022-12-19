package ordmap

import "github.com/sirkon/errors"

// New creates empty ordered map instance.
func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		data: map[K]V{},
	}
}

// OrderedMap ordered map.
type OrderedMap[K comparable, V any] struct {
	order []K
	data  map[K]V
}

// Set sets value v for key k.
// The order depends on whether value k was taken before.
// Once it was order does not change, only value does.
// Otherwise, the key k supposed to be the last after the
// insertion.
func (m *OrderedMap[K, V]) Set(k K, v V) {
	_, ok := m.data[k]
	if !ok {
		m.order = append(m.order, k)
	}

	m.data[k] = v
}

// Has check if there is a value with the given key.
func (m *OrderedMap[K, V]) Has(k K) bool {
	_, ok := m.data[k]
	return ok
}

// Get returns the value of the given key k.
func (m *OrderedMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.data[k]
	return v, ok
}

// MustGet works the same as Get except it panics if there's
// no given key in the data collected.
func (m *OrderedMap[K, V]) MustGet(k K) V {
	v, ok := m.data[k]
	if !ok {
		panic(errors.Newf("%v: no such key", k))
	}

	return v
}

// Delete deletes an element if it exists. Do nothing otherwise.
func (m *OrderedMap[K, V]) Delete(k K) {
	if _, ok := m.data[k]; !ok {
		return
	}

	delete(m.data, k)
	for i := range m.order {
		if m.order[i] == k {
			m.order = append(m.order[:i], m.order[i+1:]...)
			return
		}
	}
}

// Keys return all keys of data collected in the proper order.
func (m *OrderedMap[K, V]) Keys() []K {
	return m.order
}

// Range return an iterator over collected data.
func (m *OrderedMap[K, V]) Range() *OrderedMapIterator[K, V] {
	return &OrderedMapIterator[K, V]{
		i: -1,
		m: m,
	}
}

// OrderedMapIterator iterator over ordered map.
type OrderedMapIterator[K comparable, V any] struct {
	i int
	m *OrderedMap[K, V]
}

// Next checks if not all elements has been iterated yet.
func (it *OrderedMapIterator[K, V]) Next() bool {
	if it.i < len(it.m.order) {
		it.i++
		return true
	}

	return false
}

// KV returns (key, value) pair of current iteration.
func (it *OrderedMapIterator[K, V]) KV() (K, V) {
	k := it.m.order[it.i]
	return k, it.m.data[k]
}
