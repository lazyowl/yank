package file_control

type Bit_vector struct {
	Bit_vec uint64
}

/* returns fraction of bits set to 1 in the first n bits */
func (b Bit_vector) Percent_set(n int) int {
	count := 0
	one := uint64(0x01)
	for i := 0; i < n; i++ {
		if (b.Bit_vec & one) > 0 {
			count++
		}
		one <<= 1
	}
	if n == 0 {
		return -1
	}
	return 100 * count / n
}

/* returns bit vector with all bits set */
func Bit_vector_ones() Bit_vector {
	b := Bit_vector{}
	b.Bit_vec = 0xffffffffffffffff;
	return b
}

/* returns bit vector with all bits zero */
func Bit_vector_zero() Bit_vector {
	b := Bit_vector{}
	b.Bit_vec = 0x0000000000000000;
	return b
}

/* a | b and stores result in a */
func (a *Bit_vector) Bit_vector_or(b Bit_vector) {
	a.Bit_vec |= b.Bit_vec
}
