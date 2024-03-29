package sortedmap

import (
	"github.com/yunbaifan/pkg/list"
)

type (
	Entry[K comparable, V any] struct {
		Key     K
		Value   V
		element *list.Element[*Entry[K, V]]
	}
)

func (e *Entry[K, V]) Next() *Entry[K, V] {
	if p := e.element.Next(); p != nil {
		return p.Value
	}
	return nil
}

func (e *Entry[K, V]) Prev() *Entry[K, V] {
	if p := e.element.Prev(); p != nil {
		return p.Value
	}
	return nil
}
