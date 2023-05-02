package types

type IVec2 [2]int

func (v *IVec2) X() int {
	return v[0]
}

func (v *IVec2) Y() int {
	return v[1]
}
