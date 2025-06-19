//go:build ignore

package main

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"

	"zgo.at/otp"
)

type User struct {
	ID         int
	Email      string
	TOTPSecret []byte
}

var userStore = make(map[int]*User)

func findUser(id int) *User {
	// Normally this would be e.g. "select * from users where id = ?
	if u, ok := userStore[id]; ok {
		return u
	}
	userStore[id] = &User{ID: id, Email: "test@example.com"}
	return userStore[id]
}

func (u *User) setSecret(secret []byte) {
	// Normally this wouild be e.g. update users set totp_secret = ? where id = ?
	u.TOTPSecret = secret
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Store shared secret for this user in the database, e.g.
		user := findUser(1)
		user.setSecret(otp.Secret())

		// Generate URL that authenticator apps can pick up on.
		url := otp.URL(user.TOTPSecret, "example.com", user.Email)
		png, err := url.PNGDataURL(200)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// It's typically a good idea to also print the URL or shared secret (as
		// base32), so users can enter it manually.
		fmt.Fprintf(w, `
			<h1>Activate OTP</h1>
			<p>Secret: <code>%s</code></p>
			<img src="%s">

			<form method="POST" action="verify">
				<input type="text" name="token">
				<button>Verify</button>
			</form>
		`, url.String(), png)
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		// Read token from form.
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		token := r.Form.Get("token")
		if token == "" {
			http.Error(w, "no token?", 400)
			return
		}

		// Find our user, should have sharedsecret set.
		user := findUser(1)

		// Verify token.
		o := otp.New(user.TOTPSecret, 6, sha1.New, otp.TOTP(0, nil))
		if !o.Verify(token, 1) {
			http.Error(w, "error: invalid token", 400)
			return
		}
		fmt.Fprintf(w, `Okay!`)
	})

	fmt.Println("listening on localhost:3000")
	err := http.ListenAndServe("localhost:3000", mux)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
