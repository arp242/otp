// Package otp implemnts HOTP and TOTP one-time passwords.
package otp // import "code.soquee.net/otp"

import (
	"crypto"
	"crypto/hmac"
	"encoding/base32"
	"encoding/binary"
	"hash"
	"math"
	"net/url"
	"strconv"
	"time"
)

// URL returns a URL that is compatible with many popular OTP apps such as
// FreeOTP, Yubico Authenticator, and Google Authenticator.
//
// Supported hashes are SHA1, SHA256, and SHA512.
// Anything else will default to SHA1.
func URL(key []byte, step time.Duration, l int, hash crypto.Hash, domain, email string) *url.URL {
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(key)

	u := &url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   domain + ":" + email,
	}

	// TODO: Is it safe to do this as a string and avoid the heap allocations?
	// Domain looks like the only thing that would need to be explicitly URL
	// encoded.
	v := url.Values{}
	switch hash {
	case crypto.SHA1:
		v.Add("algorithm", "SHA1")
	case crypto.SHA256:
		v.Add("algorithm", "SHA256")
	case crypto.SHA512:
		v.Add("algorithm", "SHA512")
	default:
		v.Add("algorithm", "SHA1")
	}
	v.Add("secret", secret)
	v.Add("issuer", domain)
	v.Add("digits", strconv.Itoa(l))
	v.Add("period", strconv.FormatFloat(math.Floor(step.Seconds()), 'f', 0, 64))
	u.RawQuery = v.Encode()
	return u
}

// CounterFunc is a function that is called when generating a one-time password
// and returns a seed value.
// In HOTP this will be an incrementing counter, in TOTP it is a function of the
// current time.
// Offset indicates that we want the token relative to the current token by
// offset (eg. -1 for the previous token).
type CounterFunc func(offset int) uint64

// TOTP returns a counter function that can be used to generate HOTP tokens
// compatible with the Time-Based One-Time Password Algorithm (TOTP) defined in
// RFC 6238.
//
// If a zero duration is provided, a default of 30 seconds is used.
// If no time function is provided, time.Now is used.
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

// NewOTP returns a function that generates hmac-based one-time passwords.
// Each time the returned function is called it calls c and appends the one-time
// password to dst. It also returns a 31-bit representation of the value.
// The key is the shared secret, l is the length of the output number (if l is
// less than or equal to 0, NewOTP panics), h is a function that returns the
// inner and outer hash mechanisms for the HMAC, and c returns the seed used to
// generate the key.
func NewOTP(key []byte, l int, h func() hash.Hash, c CounterFunc) func(offset int, dst []byte) int32 {
	if l <= 0 {
		panic("otp: l must be greater than 0")
	}
	if c == nil {
		panic("otp: counter func must not be nil")
	}
	if len(key) == 0 {
		panic("otp: key must not be empty")
	}
	hs := hmac.New(h, key)
	return func(offset int, dst []byte) int32 {
		hs.Reset()
		err := binary.Write(hs, binary.BigEndian, c(offset))
		if err != nil {
			panic(err)
		}
		dst = hs.Sum(dst)
		dstOffset := dst[len(dst)-1] & 0xf
		value := int64(((int(dst[dstOffset]))&0x7f)<<24 |
			((int(dst[dstOffset+1] & 0xff)) << 16) |
			((int(dst[dstOffset+2] & 0xff)) << 8) |
			(int(dst[dstOffset+3]) & 0xff))

		return int32(value % int64(math.Pow10(l)))
	}
}
