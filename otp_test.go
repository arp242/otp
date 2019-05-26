package otp_test

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"strconv"
	"testing"
	"time"

	"code.soquee.net/otp"
)

const (
	secret    = "12345678901234567890"
	secret256 = "12345678901234567890123456789012"
	secret512 = "1234567890123456789012345678901234567890123456789012345678901234"
)

var urlTestCases = [...]struct {
	key    []byte
	step   time.Duration
	l      int
	hash   crypto.Hash
	domain string
	email  string
	out    string
}{
	0: {
		key:    []byte(secret),
		step:   30 * time.Second,
		l:      8,
		hash:   crypto.SHA1,
		domain: "example.net",
		email:  "me@example.net",
		out:    "otpauth://totp/example.net:me@example.net?algorithm=SHA1&digits=8&issuer=example.net&period=30&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
	},
	1: {
		key:    []byte(secret256),
		step:   time.Second,
		l:      6,
		hash:   crypto.SHA256,
		domain: "example.com",
		email:  "me@example.com",
		out:    "otpauth://totp/example.com:me@example.com?algorithm=SHA256&digits=6&issuer=example.com&period=1&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZA",
	},
	2: {
		key:    []byte(secret512),
		step:   0,
		l:      6,
		hash:   crypto.SHA512,
		domain: "example.com",
		email:  "me@example.com",
		out:    "otpauth://totp/example.com:me@example.com?algorithm=SHA512&digits=6&issuer=example.com&period=0&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNA",
	},
	3: {
		key:    []byte(secret),
		step:   30 * time.Second,
		l:      8,
		hash:   crypto.RIPEMD160,
		domain: "example.net",
		email:  "me@example.net",
		out:    "otpauth://totp/example.net:me@example.net?algorithm=SHA1&digits=8&issuer=example.net&period=30&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
	},
}

func TestURL(t *testing.T) {
	for i, tc := range urlTestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			u := otp.URL(tc.key, tc.step, tc.l, tc.hash, tc.domain, tc.email).String()
			if u != tc.out {
				t.Errorf("Got invalid URL: want=%q, got=%q", tc.out, u)
			}
		})
	}
}

// See RFC 6238 Appendix B
var rfc6238TestCases = [...]struct {
	t      string
	h      func() hash.Hash
	totp   int32
	secret []byte
	offset int
}{
	0: {
		t:      "1970-01-01 00:00:59",
		h:      sha1.New,
		totp:   94287082,
		secret: []byte(secret),
	},
	1: {
		t:      "1970-01-01 00:00:59",
		h:      sha256.New,
		totp:   46119246,
		secret: []byte(secret256),
	},
	2: {
		t:      "1970-01-01 00:00:59",
		h:      sha512.New,
		totp:   90693936,
		secret: []byte(secret512),
	},
	3: {
		t:      "2005-03-18 01:58:29",
		h:      sha1.New,
		totp:   7081804,
		secret: []byte(secret),
	},
	4: {
		t:      "2005-03-18 01:58:29",
		h:      sha256.New,
		totp:   68084774,
		secret: []byte(secret256),
	},
	5: {
		t:      "2005-03-18 01:58:29",
		h:      sha512.New,
		totp:   25091201,
		secret: []byte(secret512),
	},
	6: {
		t:      "2005-03-18 01:58:31",
		h:      sha1.New,
		totp:   14050471,
		secret: []byte(secret),
	},
	7: {
		t:      "2005-03-18 01:58:31",
		h:      sha256.New,
		totp:   67062674,
		secret: []byte(secret256),
	},
	8: {
		t:      "2005-03-18 01:58:31",
		h:      sha512.New,
		totp:   99943326,
		secret: []byte(secret512),
	},
	9: {
		t:      "2009-02-13 23:31:30",
		h:      sha1.New,
		totp:   89005924,
		secret: []byte(secret),
	},
	10: {
		t:      "2009-02-13 23:31:30",
		h:      sha256.New,
		totp:   91819424,
		secret: []byte(secret256),
	},
	11: {
		t:      "2009-02-13 23:31:30",
		h:      sha512.New,
		totp:   93441116,
		secret: []byte(secret512),
	},
	12: {
		t:      "2033-05-18 03:33:20",
		h:      sha1.New,
		totp:   69279037,
		secret: []byte(secret),
	},
	13: {
		t:      "2033-05-18 03:33:20",
		h:      sha256.New,
		totp:   90698825,
		secret: []byte(secret256),
	},
	14: {
		t:      "2033-05-18 03:33:20",
		h:      sha512.New,
		totp:   38618901,
		secret: []byte(secret512),
	},
	15: {
		t:      "2603-10-11 11:33:20",
		h:      sha1.New,
		totp:   65353130,
		secret: []byte(secret),
	},
	16: {
		t:      "2603-10-11 11:33:20",
		h:      sha256.New,
		totp:   77737706,
		secret: []byte(secret256),
	},
	17: {
		t:      "2603-10-11 11:33:20",
		h:      sha512.New,
		totp:   47863826,
		secret: []byte(secret512),
	},

	// Non-RFC cases
	18: {
		t:      "2603-10-11 11:33:30",
		h:      sha512.New,
		totp:   47863826,
		secret: []byte(secret512),
		offset: -1,
	},
	19: {
		t:      "2603-10-11 11:34:00",
		h:      sha512.New,
		totp:   47863826,
		secret: []byte(secret512),
		offset: -2,
	},
	20: {
		t:      "2603-10-11 11:32:40",
		h:      sha512.New,
		totp:   47863826,
		secret: []byte(secret512),
		offset: 1,
	},
}

func TestRFC6238Vectors(t *testing.T) {
	for i, tc := range rfc6238TestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			o := otp.NewOTP(tc.secret, 8, tc.h, otp.TOTP(0, func() time.Time {
				tt, err := time.Parse("2006-01-02 15:04:05", tc.t)
				if err != nil {
					panic(err)
				}
				return tt
			}))
			dst := make([]byte, 0, 64)
			otp := o(tc.offset, dst)
			if otp != tc.totp {
				t.Errorf("Unexpected TOTP: got=%d, want=%d", otp, tc.totp)
			}
		})
	}
}

// See RFC 4226 Appendix D
var rfc4226TestCases = [...]int32{
	0: 755224,
	1: 287082,
	2: 359152,
	3: 969429,
	4: 338314,
	5: 254676,
	6: 287922,
	7: 162583,
	8: 399871,
	9: 520489,
}

func TestRFC4226Vectors(t *testing.T) {
	var i int
	o := otp.NewOTP([]byte(secret), 6, sha1.New, func(offset int) uint64 {
		return uint64(i + offset)
	})
	var tc int32
	for i, tc = range rfc4226TestCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			dst := make([]byte, 0, 20)
			hotp := o(0, dst)
			if hotp != tc {
				t.Errorf("Unexpected HOTP: got=%d, want=%d", hotp, tc)
			}

			// Run each test twice, once with a fresh generator to make sure the hmac
			// is being reset properly between each use.
			o := otp.NewOTP([]byte(secret), 6, sha1.New, func(offset int) uint64 {
				return uint64(i + offset)
			})
			hotp = o(0, nil)
			if hotp != tc {
				t.Errorf("Unexpected fresh HOTP: got=%d, want=%d", hotp, tc)
			}
		})
	}
}

func TestExpectedPanics(t *testing.T) {
	t.Run("0l", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected 0 l-value to panic")
			}
		}()
		otp.NewOTP([]byte(secret), 0, sha1.New, otp.TOTP(0, nil))
	})

	t.Run("negl", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected negative l-value to panic")
			}
		}()
		otp.NewOTP([]byte(secret), -1, sha1.New, otp.TOTP(0, nil))
	})

	t.Run("nilcounter", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected nil counter func to panic")
			}
		}()
		otp.NewOTP([]byte(secret), 8, sha1.New, nil)
	})

	t.Run("nilhash", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected nil hash to panic")
			}
		}()
		otp.NewOTP([]byte(secret), 8, nil, otp.TOTP(0, nil))
	})

	t.Run("nilkey", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected nil key to panic")
			}
		}()
		otp.NewOTP(nil, 8, sha1.New, otp.TOTP(0, nil))
	})

	t.Run("emptykey", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected empty key to panic")
			}
		}()
		otp.NewOTP([]byte{}, 8, sha1.New, otp.TOTP(0, nil))
	})
}

func TestMallocs(t *testing.T) {
	o := otp.NewOTP([]byte(secret), 6, sha1.New, otp.TOTP(0, nil))
	buf := make([]byte, 0, 20)
	n := testing.AllocsPerRun(1000, func() {
		_ = o(0, buf)
	})
	if n != 2 {
		t.Errorf("Want 2 allocs, got %f", n)
	}
}
