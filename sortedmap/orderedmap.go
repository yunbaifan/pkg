package sortedmap

import (
	"github.com/yunbaifan/pkg/list"
)

type (
	OrderedMap[K comparable, V any] struct {
		entries map[K]*Entry[K, V]
		l       *list.List[*Entry[K, V]]
	}
)

func NewInit[K comparable, V any]() *OrderedMap[K, V] {
	m := OrderedMap[K, V]{
		entries: make(map[K]*Entry[K, V]),
		l:       list.New[*Entry[K, V]](),
	}
	return &m
}

func (m *OrderedMap[K, V]) Get(key K) (val V, ok bool) {
	if entry, ok := m.entries[key]; ok {
		return entry.Value, true
	}
	return
}

func (m *OrderedMap[K, V]) Set(key K, value V) (val V, ok bool) {
	if entry, ok := m.entries[key]; ok {
		oldValue := entry.Value // 保存旧值
		entry.Value = value     // 更新值
		return oldValue, true   // 返回旧值
	}

	entry := &Entry[K, V]{
		Key:   key,
		Value: value,
	}
	// 将entry插入到链表中
	entry.element = m.l.PushBack(entry) // 将entry插入到链表尾部
	m.entries[key] = entry              // 更新map
	return value, false
}

func (m *OrderedMap[K, V]) Delete(key K) (val V, ok bool) {
	if entry, ok := m.entries[key]; ok {
		m.l.Remove(entry.element) // 从链表中删除
		delete(m.entries, key)    // 从map中删除
		return entry.Value, true
	}
	return
}

func (m *OrderedMap[K, V]) Range(fun func(key K, value V) bool) {
	maps := m.l
	// 遍历链表
	for e := maps.Front(); e != nil; e = e.Next() {
		if e.Value != nil {
			if ok := fun(e.Value.Key, e.Value.Value); !ok {
				return
			}
		}
	}
}
