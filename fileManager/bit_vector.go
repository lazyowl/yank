package fileManager

// Type BitVector provides an interface to a bit vector
type BitVector struct {
	BitVec uint64
}

// PercentSet returns fraction of bits set to 1 in the first n bits
func (b BitVector) PercentSet(n int) int {
	count := 0
	one := uint64(0x01)
	for i := 0; i < n; i++ {
		if (b.BitVec & one) > 0 {
			count++
		}
		one <<= 1
	}
	if n == 0 {
		return -1
	}
	return 100 * count / n
}

// BitVectorOnes returns bit vector with all bits set
func BitVectorOnes() BitVector {
	b := BitVector{}
	b.BitVec = 0xffffffffffffffff;
	return b
}

// BitVectorZero returns bit vector with all bits zero
func BitVectorZero() BitVector {
	b := BitVector{}
	b.BitVec = 0x0000000000000000;
	return b
}

// BitVectorOr performs a | b and stores result in a
func (a *BitVector) BitVectorOr(b BitVector) {
	a.BitVec |= b.BitVec
}
