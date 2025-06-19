package qr

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"zgo.at/otp/internal/qr/utils"
)

type qrcode struct {
	dimension int
	data      *utils.BitList
	content   string
	rect      image.Rectangle
	scaleFunc func(x, y int) color.Color
}

func (qr *qrcode) Content() string         { return qr.content }
func (qe *qrcode) ColorModel() color.Model { return color.GrayModel }
func (qr *qrcode) Bounds() image.Rectangle { return qr.rect }
func (qr *qrcode) Get(x, y int) bool       { return qr.data.GetBit(x*qr.dimension + y) }
func (qr *qrcode) Set(x, y int, val bool)  { qr.data.SetBit(x*qr.dimension+y, val) }
func (qr *qrcode) At(x, y int) color.Color {
	if qr.scaleFunc != nil {
		return qr.scaleFunc(x, y)
	}
	return qr.at(x, y)
}
func (qr *qrcode) at(x, y int) color.Color {
	if qr.Get(x, y) {
		return color.Black
	}
	return color.White
}

func (qr *qrcode) Scale(width, height int) error {
	var (
		orgBounds = qr.Bounds()
		orgWidth  = orgBounds.Max.X - orgBounds.Min.X
		orgHeight = orgBounds.Max.Y - orgBounds.Min.Y
		factor    = int(math.Min(float64(width)/float64(orgWidth), float64(height)/float64(orgHeight)))
		offsetX   = (width - (orgWidth * factor)) / 2
		offsetY   = (height - (orgHeight * factor)) / 2
	)
	if factor <= 0 {
		return fmt.Errorf("can not scale barcode to an image smaller than %dx%d", orgWidth, orgHeight)
	}

	qr.rect = image.Rect(0, 0, width, height)
	qr.scaleFunc = func(x, y int) color.Color {
		if x < offsetX || y < offsetY {
			return color.White
		}
		x, y = (x-offsetX)/factor, (y-offsetY)/factor
		if x >= orgWidth || y >= orgHeight {
			return color.White
		}
		return qr.at(x, y)
	}
	return nil
}

func (qr *qrcode) calcPenalty() uint {
	return qr.calcPenaltyRule1() + qr.calcPenaltyRule2() + qr.calcPenaltyRule3() + qr.calcPenaltyRule4()
}

func (qr *qrcode) calcPenaltyRule1() uint {
	var result uint
	for x := 0; x < qr.dimension; x++ {
		var (
			checkForX, checkForY bool
			cntX, cntY           uint
		)
		for y := 0; y < qr.dimension; y++ {
			if qr.Get(x, y) == checkForX {
				cntX++
			} else {
				checkForX = !checkForX
				if cntX >= 5 {
					result += cntX - 2
				}
				cntX = 1
			}

			if qr.Get(y, x) == checkForY {
				cntY++
			} else {
				checkForY = !checkForY
				if cntY >= 5 {
					result += cntY - 2
				}
				cntY = 1
			}
		}

		if cntX >= 5 {
			result += cntX - 2
		}
		if cntY >= 5 {
			result += cntY - 2
		}
	}

	return result
}

func (qr *qrcode) calcPenaltyRule2() uint {
	var result uint
	for x := 0; x < qr.dimension-1; x++ {
		for y := 0; y < qr.dimension-1; y++ {
			check := qr.Get(x, y)
			if qr.Get(x, y+1) == check && qr.Get(x+1, y) == check && qr.Get(x+1, y+1) == check {
				result += 3
			}
		}
	}
	return result
}

func (qr *qrcode) calcPenaltyRule3() uint {
	var (
		pattern1 = []bool{true, false, true, true, true, false, true, false, false, false, false}
		pattern2 = []bool{false, false, false, false, true, false, true, true, true, false, true}
		result   uint
	)
	for x := 0; x <= qr.dimension-len(pattern1); x++ {
		for y := 0; y < qr.dimension; y++ {
			pattern1XFound := true
			pattern2XFound := true
			pattern1YFound := true
			pattern2YFound := true

			for i := 0; i < len(pattern1); i++ {
				iv := qr.Get(x+i, y)
				if iv != pattern1[i] {
					pattern1XFound = false
				}
				if iv != pattern2[i] {
					pattern2XFound = false
				}
				iv = qr.Get(y, x+i)
				if iv != pattern1[i] {
					pattern1YFound = false
				}
				if iv != pattern2[i] {
					pattern2YFound = false
				}
			}
			if pattern1XFound || pattern2XFound {
				result += 40
			}
			if pattern1YFound || pattern2YFound {
				result += 40
			}
		}
	}

	return result
}

func (qr *qrcode) calcPenaltyRule4() uint {
	totalNum := qr.data.Len()
	trueCnt := 0
	for i := 0; i < totalNum; i++ {
		if qr.data.GetBit(i) {
			trueCnt++
		}
	}
	percDark := float64(trueCnt) * 100 / float64(totalNum)
	floor := math.Abs(math.Floor(percDark/5) - 10)
	ceil := math.Abs(math.Ceil(percDark/5) - 10)
	return uint(math.Min(floor, ceil) * 10)
}
