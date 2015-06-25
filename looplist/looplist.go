package looplist

type Element struct {
	Value interface{}
	List  *LoopList
	Pos   int
}

type LoopList struct {
	List []*Element
	Tail int //start from zero
	Cap  int
}

func NewLoopList(_cap int) *LoopList {
	l := &LoopList{
		List: make([]*Element, 0, _cap),
		Cap:  _cap,
		Tail: -1,
	}
	return l
}

func (l *LoopList) Append(value interface{}) {
	list_length := len(l.List)
	el := &Element{
		Value: value,
		List:  l,
	}
	if list_length < l.Cap {
		el.Pos = list_length
		l.List = append(l.List, el)
		l.Tail = list_length
	} else {
		l.Tail = (l.Tail + 1) % l.Cap
		el.Pos = l.Tail
		l.List[l.Tail].Value = nil // avoid memory leaks
		l.List[l.Tail].List = nil  // avoid memory leaks
		l.List[l.Tail] = el
	}
}

func (l *LoopList) Back() *Element {
	if len(l.List) == 0 {
		return nil
	}
	return l.List[l.Tail]
}

func (e *Element) Prev() *Element {
	l := e.List
	_cap := len(l.List)
	prev := (e.Pos + _cap - 1) % _cap
	if prev == l.Tail {
		return nil
	}
	return l.List[prev]
}

func (l *LoopList) Front() *Element {
	_cap := len(l.List)
	if _cap == 0 {
		return nil
	}
	front := (l.Tail + 1) % _cap
	return l.List[front]
}

func (e *Element) Next() *Element{
	l := e.List
	_cap := len(l.List)
	next := (e.Pos + 1) % _cap
	if next == (l.Tail + 1) % _cap {
		return nil
	}
	return l.List[next]
}
