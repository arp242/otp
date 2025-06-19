package utils

type gfpoly struct {
	gf           *GaloisField
	Coefficients []int
}

func (gp *gfpoly) Degree() int {
	return len(gp.Coefficients) - 1
}

func (gp *gfpoly) Zero() bool {
	return gp.Coefficients[0] == 0
}

// getCoefficient returns the coefficient of x ^ degree
func (gp *gfpoly) getCoefficient(degree int) int {
	return gp.Coefficients[gp.Degree()-degree]
}

func (gp *gfpoly) addOrSubstract(other *gfpoly) *gfpoly {
	if gp.Zero() {
		return other
	} else if other.Zero() {
		return gp
	}
	smallCoeff := gp.Coefficients
	largeCoeff := other.Coefficients
	if len(smallCoeff) > len(largeCoeff) {
		largeCoeff, smallCoeff = smallCoeff, largeCoeff
	}
	sumDiff := make([]int, len(largeCoeff))
	lenDiff := len(largeCoeff) - len(smallCoeff)
	copy(sumDiff, largeCoeff[:lenDiff])
	for i := lenDiff; i < len(largeCoeff); i++ {
		sumDiff[i] = int(gp.gf.AddOrSub(int(smallCoeff[i-lenDiff]), int(largeCoeff[i])))
	}
	return newGFPoly(gp.gf, sumDiff)
}

func (gp *gfpoly) multByMonominal(degree int, coeff int) *gfpoly {
	if coeff == 0 {
		return gp.gf.Zero()
	}
	size := len(gp.Coefficients)
	result := make([]int, size+degree)
	for i := 0; i < size; i++ {
		result[i] = int(gp.gf.Multiply(int(gp.Coefficients[i]), int(coeff)))
	}
	return newGFPoly(gp.gf, result)
}

func (gp *gfpoly) multiply(other *gfpoly) *gfpoly {
	if gp.Zero() || other.Zero() {
		return gp.gf.Zero()
	}
	aCoeff := gp.Coefficients
	aLen := len(aCoeff)
	bCoeff := other.Coefficients
	bLen := len(bCoeff)
	product := make([]int, aLen+bLen-1)
	for i := 0; i < aLen; i++ {
		ac := int(aCoeff[i])
		for j := 0; j < bLen; j++ {
			bc := int(bCoeff[j])
			product[i+j] = int(gp.gf.AddOrSub(int(product[i+j]), gp.gf.Multiply(ac, bc)))
		}
	}
	return newGFPoly(gp.gf, product)
}

func (gp *gfpoly) divide(other *gfpoly) (quotient *gfpoly, remainder *gfpoly) {
	quotient = gp.gf.Zero()
	remainder = gp
	fld := gp.gf
	denomLeadTerm := other.getCoefficient(other.Degree())
	inversDenomLeadTerm := fld.Invers(int(denomLeadTerm))
	for remainder.Degree() >= other.Degree() && !remainder.Zero() {
		degreeDiff := remainder.Degree() - other.Degree()
		scale := int(fld.Multiply(int(remainder.getCoefficient(remainder.Degree())), inversDenomLeadTerm))
		term := other.multByMonominal(degreeDiff, scale)
		itQuot := newMonominalPoly(fld, degreeDiff, scale)
		quotient = quotient.addOrSubstract(itQuot)
		remainder = remainder.addOrSubstract(term)
	}
	return
}

func newMonominalPoly(field *GaloisField, degree int, coeff int) *gfpoly {
	if coeff == 0 {
		return field.Zero()
	}
	result := make([]int, degree+1)
	result[0] = coeff
	return newGFPoly(field, result)
}

func newGFPoly(field *GaloisField, coefficients []int) *gfpoly {
	for len(coefficients) > 1 && coefficients[0] == 0 {
		coefficients = coefficients[1:]
	}
	return &gfpoly{field, coefficients}
}
