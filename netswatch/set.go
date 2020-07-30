package netswatch

type void struct{} //Empty stuct, 0 byte

type IDSet struct {
	box map[string]void
}

func NewSet() *IDSet {
	s := &IDSet{}
	s.box = make(map[string]void)
	return s
}

func (s *IDSet) Has(v string) bool {
	_, ok := s.box[v]
	return ok
}

func (s *IDSet) Add(v string) {
	s.box[v] = void{}
}

func (s *IDSet) Remove(v string) {
	delete(s.box, v)
}

func (s *IDSet) Size() int {
	return len(s.box)
}

func (s *IDSet) Clear() {
	s.box = make(map[string]void)
}

func (s *IDSet) Union(s2 *IDSet) *IDSet {
	ns := NewSet()
	for v := range s.box {
		ns.Add(v)
	}
	for v := range s2.box {
		ns.Add(v)
	}
	return ns
}

func (s *IDSet) Intersect(s2 *IDSet) *IDSet {
	ns := NewSet()
	for v := range s.box {
		if s2.Has(v) {
			ns.Add(v)
		}
	}
	return ns
}

func (s *IDSet) Difference(s2 *IDSet) *IDSet {
	ns := NewSet()
	for v := range s.box {
		if s2.Has(v) {
			continue
		}
		ns.Add(v)
	}
	return ns
}
