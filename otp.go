// Package otp implemnts HOTP and TOTP one-time passwords.
package otp

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"hash"
	"image/png"
	"math"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"zgo.at/otp/internal/qr"
)

// Secret generates a new shared secret.
//
// It's not required to use this function specifically. It's just here for
// convenience.
func Secret() []byte {
	// rfc4226 says length MUST be at least 16 bytes, and RECOMMENDs 20 bytes. I
	// don't see any service using more than 20, so should be fine to just
	// hard-code it here. Downside of making it larger is that the QR code
	// becomes larger.
	s := make([]byte, 20)
	_, _ = rand.Read(s) // Documented as never returning an error
	return s
}

type (
	url struct {
		url *neturl.URL
	}
	generator struct {
		length  int
		counter CounterFunc
		h       func() hash.Hash
		secret  []byte
	}

	// CounterFunc is a function that is called when generating a one-time password
	// and returns a seed value.
	//
	// In HOTP this will be an incrementing counter, in TOTP it returns the current
	// time.
	//
	// Offset indicates that we want the token relative to the current token by
	// offset (eg. -1 for the previous token).
	CounterFunc func(offset int) uint64
)

// Token generates a new token.
//
// Offset indicates that we want the token relative to the current token by
// offset (eg. -1 for the previous token).
func (g generator) Token(offset int) string {
	hm := hmac.New(g.h, g.secret)

	// Never returns an error unless the write fails (which is a hash writer) or
	// invalid input.
	_ = binary.Write(hm, binary.BigEndian, g.counter(offset))

	var (
		h   = hm.Sum(nil)
		off = h[len(h)-1] & 0xf
		v   = ((int(h[off]))&0x7f)<<24 |
			((int(h[off+1] & 0xff)) << 16) |
			((int(h[off+2] & 0xff)) << 8) |
			(int(h[off+3]) & 0xff)
		s = strconv.Itoa(v % int(math.Pow10(g.length)))
	)
	if ll := g.length - len(s); ll > 0 {
		s = strings.Repeat("0", ll) + s
	}
	return s
}

// Verify a token.
//
// If offset is higher than 0, it will also accept tokens from -offset to
// +offset. This can be useful to allow some clock skew for e.g. TOTP.
func (g generator) Verify(token string, offset int) bool {
	for i := -offset; i <= offset; i++ {
		if token == g.Token(i) {
			return true
		}
	}
	return false
}

// New returns a generator to generate and verify HMAC one-time passwords.
//
// Panics if tokenLength is <= 0 or if any of the other parameters are nil.
func New(sharedSecret []byte, tokenLength int, hash func() hash.Hash, c CounterFunc) generator {
	if tokenLength <= 0 {
		panic("otp.New: tokenLength must be greater than 0")
	}
	if c == nil {
		panic("otp.New: counter func must not be nil")
	}
	if hash == nil {
		panic("otp.New: hash func must not be nil")
	}
	if len(sharedSecret) == 0 {
		panic("otp.New: sharedSecret must not be empty")
	}
	return generator{tokenLength, c, hash, sharedSecret}
}

// TOTP returns a counter function to generate TOTP tokens as defined in
// RFC6238.
//
// TOTP tokens are time-based and valid for step duration. It will use 30
// seconds if zero, which is a reasonable default, but in some cases where clock
// skew is expected a longer value may be used.
//
// Providing the time can be useful to provide a fixed time for testing. It uses
// time.Now() if nil.
func TOTP(step time.Duration, t func() time.Time) CounterFunc {
	if step == 0 {
		step = 30 * time.Second
	}
	if t == nil {
		t = time.Now
	}
	return func(offset int) uint64 {
		return uint64(math.Floor(float64(t().Add(time.Duration(offset)*step).Unix()) / step.Seconds()))
	}
}

// String returns the URL.
func (u url) String() string {
	return u.url.String()
}

// PNGDataURL returns a QR code as a PNG data URL.
func (u url) PNGDataURL(size int) (string, error) {
	code, err := qr.Encode(qr.M, u.String())
	if err != nil {
		return "", err
	}
	err = code.Scale(size, size)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("data:image/png;base64,")
	err = png.Encode(base64.NewEncoder(base64.StdEncoding, buf), code)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// URL creates an URL that can be used by most OTP apps, such as FreeOTP, Yubico
// Authenticator, Google Authenticator, etc.
func URL(key []byte, issuer, email string) url {
	return url{&neturl.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   issuer + ":" + email,
		RawQuery: neturl.Values{
			"issuer": {issuer},
			"secret": {base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(key)},
		}.Encode(),
	}}
}
