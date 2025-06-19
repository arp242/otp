package otp_test

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"math"
	"testing"
	"time"

	"zgo.at/otp"
)

var (
	secret    = []byte("12345678901234567890")
	secret256 = []byte("12345678901234567890123456789012")
	secret512 = []byte("1234567890123456789012345678901234567890123456789012345678901234")
)

func wantPanic(t *testing.T, want string) {
	t.Helper()
	have := recover()
	if have == nil {
		t.Errorf("expected panic")
		return
	}
	var str string
	switch h := have.(type) {
	case string:
		str = h
	case error:
		if h == nil {
			str = "<nil>"
		} else {
			str = h.Error()
		}
	default:
		t.Errorf("unknown panic type: %T", have)
		return
	}
	if str != want {
		t.Errorf("wrong panic\nhave: %q\nwant: %q", str, want)
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		secret []byte
		issuer string
		email  string
		want   string
	}{
		{
			secret: secret,
			issuer: "example.net",
			email:  "me@example.net",
			want:   "otpauth://totp/example.net:me@example.net?issuer=example.net&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
		},
		{
			secret: secret,
			issuer: "example.net",
			email:  "me@example.net",
			want:   "otpauth://totp/example.net:me@example.net?issuer=example.net&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
		},
		{
			secret: secret256,
			issuer: "example.com",
			email:  "me@example.com",
			want:   "otpauth://totp/example.com:me@example.com?issuer=example.com&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZA",
		},
		{
			secret: secret512,
			issuer: "example.com",
			email:  "me@example.com",
			want:   "otpauth://totp/example.com:me@example.com?issuer=example.com&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNA",
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			have := otp.URL(tt.secret, tt.issuer, tt.email).String()
			if have != tt.want {
				t.Errorf("\nhave: %q\nwant: %q", have, tt.want)
			}
		})
	}
}

func TestGenerator(t *testing.T) {
	var tests = []struct {
		want   string
		t      string
		h      func() hash.Hash
		secret []byte
		offset int
	}{
		// From RFC 6238 Appendix B
		{"94287082", "1970-01-01 00:00:59", sha1.New, secret, 0},
		{"46119246", "1970-01-01 00:00:59", sha256.New, secret256, 0},
		{"90693936", "1970-01-01 00:00:59", sha512.New, secret512, 0},
		{"07081804", "2005-03-18 01:58:29", sha1.New, secret, 0},
		{"68084774", "2005-03-18 01:58:29", sha256.New, secret256, 0},
		{"25091201", "2005-03-18 01:58:29", sha512.New, secret512, 0},
		{"14050471", "2005-03-18 01:58:31", sha1.New, secret, 0},
		{"67062674", "2005-03-18 01:58:31", sha256.New, secret256, 0},
		{"99943326", "2005-03-18 01:58:31", sha512.New, secret512, 0},
		{"89005924", "2009-02-13 23:31:30", sha1.New, secret, 0},
		{"91819424", "2009-02-13 23:31:30", sha256.New, secret256, 0},
		{"93441116", "2009-02-13 23:31:30", sha512.New, secret512, 0},
		{"69279037", "2033-05-18 03:33:20", sha1.New, secret, 0},
		{"90698825", "2033-05-18 03:33:20", sha256.New, secret256, 0},
		{"38618901", "2033-05-18 03:33:20", sha512.New, secret512, 0},
		{"65353130", "2603-10-11 11:33:20", sha1.New, secret, 0},
		{"77737706", "2603-10-11 11:33:20", sha256.New, secret256, 0},
		{"47863826", "2603-10-11 11:33:20", sha512.New, secret512, 0},

		// Non-RFC cases
		{"47863826", "2603-10-11 11:33:30", sha512.New, secret512, -1},
		{"47863826", "2603-10-11 11:34:00", sha512.New, secret512, -2},
		{"47863826", "2603-10-11 11:32:40", sha512.New, secret512, 1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			o := otp.New(tt.secret, 8, tt.h, otp.TOTP(0, func() time.Time {
				tt, err := time.Parse("2006-01-02 15:04:05", tt.t)
				if err != nil {
					panic(err)
				}
				return tt
			}))
			have := o.Token(tt.offset)
			if have != tt.want {
				t.Errorf("\nhave: %q\nwant: %q", have, tt.want)
			}
			if !o.Verify(tt.want, int(math.Abs(float64(tt.offset)))) {
				t.Error("Verify() failed")
			}
		})
	}
}

func TestRFC4226(t *testing.T) {
	// See RFC 4226 Appendix D
	tests := []string{"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489"}

	var i int
	o := otp.New(secret, 6, sha1.New, func(offset int) uint64 {
		return uint64(i + offset)
	})
	var tt string
	for i, tt = range tests {
		t.Run("", func(t *testing.T) {
			have := o.Token(0)
			if have != tt {
				t.Errorf("\nhave: %q\nwant: %q", have, tt)
			}
			if !o.Verify(tt, 0) {
				t.Error("Verify() failed")
			}

			// Run each test twice, once with a fresh generator to make sure the
			// hmac is being reset properly between each use.
			o := otp.New(secret, 6, sha1.New, func(offset int) uint64 {
				return uint64(i + offset)
			})
			have = o.Token(0)
			if have != tt {
				t.Errorf("\nhave: %q\nwant: %q", have, tt)
			}
			if !o.Verify(tt, 0) {
				t.Error("Verify() failed")
			}
		})
	}
}

func TestVerify(t *testing.T) {
	// TestGenerator() and TestRFC4226() already test success
	o := otp.New(secret, 8, sha256.New, otp.TOTP(0, nil))
	if o.Verify("XXX", 0) {
		t.Error()
	}
	if o.Verify(o.Token(-1), 0) {
		t.Error()
	}
	if o.Verify(o.Token(1), 0) {
		t.Error()
	}
	if !o.Verify(o.Token(0), 0) {
		t.Error()
	}
	if !o.Verify(o.Token(-100), 200) {
		t.Error()
	}
	if !o.Verify(o.Token(100), 200) {
		t.Error()
	}
}

func TestPanic(t *testing.T) {
	tests := []struct {
		want string
		f    func()
	}{
		{"otp.New: tokenLength must be greater than 0", func() { otp.New(secret, 0, sha1.New, otp.TOTP(0, nil)) }},
		{"otp.New: tokenLength must be greater than 0", func() { otp.New(secret, -1, sha1.New, otp.TOTP(0, nil)) }},
		{"otp.New: counter func must not be nil", func() { otp.New(secret, 8, sha1.New, nil) }},
		{"otp.New: hash func must not be nil", func() { otp.New(secret, 8, nil, otp.TOTP(0, nil)) }},
		{"otp.New: sharedSecret must not be empty", func() { otp.New(nil, 8, sha1.New, otp.TOTP(0, nil)) }},
		{"otp.New: sharedSecret must not be empty", func() { otp.New([]byte{}, 8, sha1.New, otp.TOTP(0, nil)) }},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			defer wantPanic(t, tt.want)
			tt.f()
		})
	}
}

func TestSecret(t *testing.T) {
	one, two := otp.Secret(), otp.Secret()
	if len(one) != 20 {
		t.Error()
	}
	if len(two) != 20 {
		t.Error()
	}
	if bytes.Equal(one, two) {
		t.Error()
	}
}

func BenchmarkURL(b *testing.B) {
	for b.Loop() {
		_ = otp.URL(secret, "example.com", "me@example.com")
	}
}

func BenchmarkNew(b *testing.B) {
	f := otp.TOTP(30*time.Second, func() time.Time {
		return time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	})
	b.ResetTimer()
	for b.Loop() {
		_ = otp.New(secret, 8, sha256.New, f).Token(0)
	}
}
