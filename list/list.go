package list

type Element[V any] struct {
	next, prev *Element[V]
	list       *List[V]
	Value      V
}

// Next 查找下一个元素, 如果没有下一个元素则返回nil
func (e *Element[V]) Next() *Element[V] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev 查找上一个元素, 如果没有上一个元素则返回nil
func (e *Element[V]) Prev() *Element[V] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// List 链表
type List[V any] struct {
	root Element[V] // 哨兵元素, 用于简化链表操作
	len  int        // 当前链表长度
}

// Init 初始化链表
func (l *List[V]) Init() *List[V] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New 创建一个新的链表
func New[V any]() *List[V] { return new(List[V]).Init() }

// Len 返回链表长度
func (l *List[V]) Len() int { return l.len }

// Front 返回链表的第一个元素, 如果链表为空则返回nil
func (l *List[V]) Front() *Element[V] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back 返回链表的最后一个元素, 如果链表为空则返回nil
func (l *List[V]) Back() *Element[V] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit 初始化链表
func (l *List[V]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert 在at元素之后插入e元素
func (l *List[V]) insert(e, at *Element[V]) *Element[V] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

// insertValue 在at元素之后插入值为v的元素
func (l *List[V]) insertValue(v V, at *Element[V]) *Element[V] {
	return l.insert(&Element[V]{Value: v}, at)
}

// remove 移除元素e
func (l *List[V]) remove(e *Element[V]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil
	e.prev = nil
	e.list = nil
	l.len--
}

// move 移动元素e到at元素之后
func (l *List[V]) move(e, at *Element[V]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// Remove 移除元素e并返回其值
func (l *List[V]) Remove(e *Element[V]) V {
	if e.list == l {
		l.remove(e)
	}
	return e.Value
}

// PushFront 在链表头部插入一个元素
func (l *List[V]) PushFront(v V) *Element[V] {
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBack 在链表尾部插入一个元素
func (l *List[V]) PushBack(v V) *Element[V] {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore 在mark元素之前插入一个元素
func (l *List[V]) InsertBefore(v V, mark *Element[V]) *Element[V] {
	if mark.list != l {
		return nil
	}
	return l.insertValue(v, mark.prev)
}

// InsertAfter 在mark元素之后插入一个元素
func (l *List[V]) InsertAfter(v V, mark *Element[V]) *Element[V] {
	if mark.list != l {
		return nil
	}
	return l.insertValue(v, mark)
}

// MoveToFront 移动元素e到链表头部
func (l *List[V]) MoveToFront(e *Element[V]) {
	if e.list != l || l.root.next == e {
		return
	}
	l.move(e, &l.root)
}

// MoveToBack 移动元素e到链表尾部
func (l *List[V]) MoveToBack(e *Element[V]) {
	if e.list != l || l.root.prev == e {
		return
	}
	l.move(e, l.root.prev)
}

// MoveBefore 移动元素e到mark元素之前
func (l *List[V]) MoveBefore(e, mark *Element[V]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter 移动元素e到mark元素之后
func (l *List[V]) MoveAfter(e, mark *Element[V]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList 写入另一个链表到链表尾部
func (l *List[V]) PushBackList(other *List[V]) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList 写入另一个链表到链表头部
func (l *List[V]) PushFrontList(other *List[V]) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}
