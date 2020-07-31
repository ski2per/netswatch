package netswatch

type void struct{} //Empty stuct, 0 byte

type Set struct {
	box map[string]void
}

func NewSet() *Set {
	s := &Set{}
	s.box = make(map[string]void)
	return s
}

func (s *Set) Has(v string) bool {
	_, ok := s.box[v]
	return ok
}

func (s *Set) Add(v string) {
	s.box[v] = void{}
}

func (s *Set) Remove(v string) {
	delete(s.box, v)
}

func (s *Set) Size() int {
	return len(s.box)
}

func (s *Set) Clear() {
	s.box = make(map[string]void)
}

func (s *Set) Union(s2 *Set) *Set {
	ns := NewSet()
	for v := range s.box {
		ns.Add(v)
	}
	for v := range s2.box {
		ns.Add(v)
	}
	return ns
}

func (s *Set) Intersect(s2 *Set) *Set {
	ns := NewSet()
	for v := range s.box {
		if s2.Has(v) {
			ns.Add(v)
		}
	}
	return ns
}

func (s *Set) Difference(s2 *Set) *Set {
	ns := NewSet()
	for v := range s.box {
		if s2.Has(v) {
			continue
		}
		ns.Add(v)
	}
	return ns
}
