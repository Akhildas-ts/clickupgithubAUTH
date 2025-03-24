package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	config *oauth2.Config
	store  *sessions.CookieStore
)

type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func init() {
	// Initialize the session store with a secret key
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	// Configure OAuth2
	config = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			"user:email",
			"read:user",
		},
		Endpoint: github.Endpoint,
	}
}

func main() {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/login", loginHandler)
	r.HandleFunc("/callback", callbackHandler)
	r.HandleFunc("/logout", logoutHandler)

	// Start server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userEmail, ok := session.Values["email"].(string)

	if !ok {
		// User is not authenticated
		fmt.Fprint(w, `<a href="/login">Login with GitHub</a>`)
		return
	}

	// User is authenticated
	fmt.Fprintf(w, `
		<h1>Welcome, %s!</h1>
		<a href="/logout">Logout</a>
	`, userEmail)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := config.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := config.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Store user info in session
	session, _ := store.Get(r, "session")
	session.Values["email"] = user.Email
	session.Values["username"] = user.Login
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}