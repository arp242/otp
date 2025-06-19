package qr

import (
	"zgo.at/otp/internal/qr/utils"
)

type errorCorrection struct {
	rs *utils.ReedSolomonEncoder
}

var ec = &errorCorrection{utils.NewReedSolomonEncoder(utils.NewGaloisField(285, 256, 0))}

func (ec *errorCorrection) calcECC(data []byte, eccCount byte) []byte {
	dataInts := make([]int, len(data))
	for i := 0; i < len(data); i++ {
		dataInts[i] = int(data[i])
	}
	res := ec.rs.Encode(dataInts, int(eccCount))
	result := make([]byte, len(res))
	for i := 0; i < len(res); i++ {
		result[i] = byte(res[i])
	}
	return result
}
