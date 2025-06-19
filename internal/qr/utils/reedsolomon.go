package utils

import (
	"sync"
)

type ReedSolomonEncoder struct {
	gf        *GaloisField
	polynomes []*gfpoly
	m         *sync.Mutex
}

func NewReedSolomonEncoder(gf *GaloisField) *ReedSolomonEncoder {
	return &ReedSolomonEncoder{
		gf, []*gfpoly{newGFPoly(gf, []int{1})}, new(sync.Mutex),
	}
}

func (rs *ReedSolomonEncoder) getPolynomial(degree int) *gfpoly {
	rs.m.Lock()
	defer rs.m.Unlock()

	if degree >= len(rs.polynomes) {
		last := rs.polynomes[len(rs.polynomes)-1]
		for d := len(rs.polynomes); d <= degree; d++ {
			next := last.multiply(newGFPoly(rs.gf, []int{1, rs.gf.ALogTbl[d-1+rs.gf.Base]}))
			rs.polynomes = append(rs.polynomes, next)
			last = next
		}
	}
	return rs.polynomes[degree]
}

func (rs *ReedSolomonEncoder) Encode(data []int, eccCount int) []int {
	generator := rs.getPolynomial(eccCount)
	info := newGFPoly(rs.gf, data)
	info = info.multByMonominal(eccCount, 1)
	_, remainder := info.divide(generator)

	result := make([]int, eccCount)
	numZero := int(eccCount) - len(remainder.Coefficients)
	copy(result[numZero:], remainder.Coefficients)
	return result
}
