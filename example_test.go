package otp_test

import (
	"crypto/sha256"
	"fmt"
	"time"

	"code.soquee.net/otp"
)

func Example_totp() {
	const secret = "12345678901234567890123456789012"

	o := otp.NewOTP([]byte(secret), 8, sha256.New, otp.TOTP(30*time.Second, func() time.Time {
		// You would normally pass in time.Now, or possibly a time function that
		// subtracts some multiple of the period to correct for clock-drift.
		tt, _ := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:59")
		return tt
	}))
	fmt.Println(o(0, nil))
	// Output: 46119246
}
