package qr

import (
	"fmt"
	"image/png"
	"os"
	"strings"
	"testing"
)

type test struct {
	Text   string
	ECL    ErrorCorrectionLevel
	Result string
}

var tests = []test{
	test{
		Text: "hello world",
		ECL:  H,
		Result: `
+++++++.+.+.+...+.+++++++
+.....+.++...+++..+.....+
+.+++.+.+.+.++.++.+.+++.+
+.+++.+....++.++..+.+++.+
+.+++.+..+...++.+.+.+++.+
+.....+.+..+..+++.+.....+
+++++++.+.+.+.+.+.+++++++
........++..+..+.........
..+++.+.+++.+.++++++..+++
+++..+..+...++.+...+..+..
+...+.++++....++.+..++.++
++.+.+.++...+...+.+....++
..+..+++.+.+++++.++++++++
+.+++...+..++..++..+..+..
+.....+..+.+.....+++++.++
+.+++.....+...+.+.+++...+
+.+..+++...++.+.+++++++..
........+....++.+...+.+..
+++++++......++++.+.+.+++
+.....+....+...++...++.+.
+.+++.+.+.+...+++++++++..
+.+++.+.++...++...+.++..+
+.+++.+.++.+++++..++.+..+
+.....+..+++..++.+.++...+
+++++++....+..+.+..+..+++`,
	},
}

func imgStrToBools(str string) []bool {
	res := make([]bool, 0, len(str))
	for _, r := range str {
		if r == '+' {
			res = append(res, true)
		} else if r == '.' {
			res = append(res, false)
		}
	}
	return res
}

func Test_Encode(t *testing.T) {
	for _, tst := range tests {
		qrCode, err := Encode(tst.ECL, tst.Text)
		if err != nil {
			t.Error(err)
		}
		testRes := imgStrToBools(tst.Result)
		if (qrCode.dimension * qrCode.dimension) != len(testRes) {
			t.Fail()
		}
		t.Logf("dim %d", qrCode.dimension)
		for i := 0; i < len(testRes); i++ {
			x := i % qrCode.dimension
			y := i / qrCode.dimension
			if qrCode.Get(x, y) != testRes[i] {
				t.Errorf("Failed at index %d", i)
			}
		}
	}
}

func ExampleEncode() {
	f, _ := os.Create("qrcode.png")
	defer f.Close()

	qrcode, err := Encode(L, "hello world")
	if err != nil {
		fmt.Println(err)
	} else {
		err := qrcode.Scale(100, 100)
		if err != nil {
			fmt.Println(err)
		} else {
			png.Encode(f, qrcode)
		}
	}
}

func Test_UnicodeEncoding(t *testing.T) {
	// XXX
	//q, err := Encode(H, "A") // 65
	//if err != nil {
	//	t.Fatal(err)
	//}
	//if !bytes.Equal(q, []byte{64, 20, 16, 236, 17, 236, 17, 236, 17}) {
	//	t.Errorf("\"A\" failed to encode: %s", q.Content())
	//}
	_, err := Encode(H, strings.Repeat("A", 3000))
	if err == nil {
		t.Error("Unicode encoding should not be able to encode a 3kb string")
	}
}
