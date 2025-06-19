package qr

import (
	"image/color"
	"testing"
)

func Test_NewQRCode(t *testing.T) {
	bc := newQR(2)
	if bc == nil {
		t.Fail()
		return
	}
	if bc.data.Len() != 4 {
		t.Fail()
	}
	if bc.dimension != 2 {
		t.Fail()
	}
}

func Test_QRBasics(t *testing.T) {
	code, _ := Encode(L, "test")
	if code.Content() != "test" {
		t.Fail()
	}
	bounds := code.Bounds()
	if bounds.Min.X != 0 || bounds.Min.Y != 0 || bounds.Max.X != 21 || bounds.Max.Y != 21 {
		t.Fail()
	}
	if code.At(0, 0) != color.Black || code.At(0, 7) != color.White {
		t.Fail()
	}
	sum := code.calcPenaltyRule1() + code.calcPenaltyRule2() + code.calcPenaltyRule3() + code.calcPenaltyRule4()
	if code.calcPenalty() != sum {
		t.Fail()
	}
}

func Test_Penalty1(t *testing.T) {
	qr := newQR(7)
	if qr.calcPenaltyRule1() != 70 {
		t.Fail()
	}
	qr.Set(0, 0, true)
	if qr.calcPenaltyRule1() != 68 {
		t.Fail()
	}
	qr.Set(0, 6, true)
	if qr.calcPenaltyRule1() != 66 {
		t.Fail()
	}
}

func Test_Penalty2(t *testing.T) {
	qr := newQR(3)
	if qr.calcPenaltyRule2() != 12 {
		t.Fail()
	}
	qr.Set(0, 0, true)
	qr.Set(1, 1, true)
	qr.Set(2, 0, true)
	if qr.calcPenaltyRule2() != 0 {
		t.Fail()
	}
	qr.Set(1, 1, false)
	if qr.calcPenaltyRule2() != 6 {
		t.Fail()
	}
}

func Test_Penalty4(t *testing.T) {
	qr := newQR(3)
	if qr.calcPenaltyRule4() != 100 {
		t.Fail()
	}
	qr.Set(0, 0, true)
	if qr.calcPenaltyRule4() != 70 {
		t.Fail()
	}
	qr.Set(0, 1, true)
	if qr.calcPenaltyRule4() != 50 {
		t.Fail()
	}
	qr.Set(0, 2, true)
	if qr.calcPenaltyRule4() != 30 {
		t.Fail()
	}
	qr.Set(1, 0, true)
	if qr.calcPenaltyRule4() != 10 {
		t.Fail()
	}
	qr.Set(1, 1, true)
	if qr.calcPenaltyRule4() != 10 {
		t.Fail()
	}
	qr = newQR(2)
	qr.Set(0, 0, true)
	qr.Set(1, 0, true)
	if qr.calcPenaltyRule4() != 0 {
		t.Fail()
	}
}
