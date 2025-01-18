package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

func main() {

	var oauth2Config = &oauth2.Config{
		ClientID:    "your-client-id",
		RedirectURL: "http://localhost:8080/callback",
		Scopes:      []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://your-auth0-domain/authorize",
			TokenURL: "https://your-auth0-domain/oauth/token",
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		codeVerifier, codeChallenge, err := generatePKCE()
		if err != nil {
			http.Error(w, "Failed to generate PKCE: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// store codeVerifier securely (e.g., in a secure cookie)
		http.SetCookie(w, &http.Cookie{
			Name:     "code_verifier",
			Value:    codeVerifier,
			HttpOnly: true,
			Secure:   true,
			Path:     "/callback",
		})

		// redirect with the code challenge
		url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"))
		http.Redirect(w, r, url, http.StatusFound)

	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			return
		}

		// retrieve codeVerifier from cookie
		cookie, err := r.Cookie("code_verifier")
		if err != nil {
			http.Error(w, "Code verifier not found", http.StatusUnauthorized)
			return
		}
		codeVerifier := cookie.Value

		// exchange authorization code for token
		token, err := oauth2Config.Exchange(context.Background(), code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// store token in an HTTP-only secure cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    token.AccessToken,
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
		})

		// todo generate jwt, apply claim and set cache

		fmt.Fprintln(w, "authentication successful!")
	})

	log.Println("Server is running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func generatePKCE() (string, string, error) {
	verifier := make([]byte, 32)
	_, err := rand.Read(verifier)
	if err != nil {
		return "", "", err
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifier)

	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return codeVerifier, codeChallenge, nil
}
