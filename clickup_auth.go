package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ClickUpUser struct {
	User struct {
		ID             int    `json:"id"`
		Username       string `json:"username"`
		Email          string `json:"email"`
		ProfilePicture string `json:"profilePicture"`
	} `json:"user"`
}

type ClickUpConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

var clickupConfig ClickUpConfig

func initClickUpAuth() {
	// Configure ClickUp OAuth
	clickupConfig = ClickUpConfig{
		ClientID:     os.Getenv("CLICKUP_CLIENT_ID"),
		ClientSecret: os.Getenv("CLICKUP_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("CLICKUP_REDIRECT_URI"),
		AuthURL:      "https://app.clickup.com/api",
		TokenURL:     "https://api.clickup.com/api/v2/oauth/token",
		UserInfoURL:  "https://api.clickup.com/api/v2/user",
	}
}

func clickupHomeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userEmail, ok := session.Values["email"].(string)

	if !ok {
		// User is not authenticated
		fmt.Fprint(w, `
			<html>
				<head>
					<title>ClickUp Auth</title>
					<style>
						body { 
							font-family: Arial, sans-serif; 
							margin: 40px; 
							background-color: #f5f5f5;
							display: flex;
							flex-direction: column;
							align-items: center;
						}
						.container {
							background-color: white;
							padding: 30px;
							border-radius: 10px;
							box-shadow: 0 2px 4px rgba(0,0,0,0.1);
							text-align: center;
						}
						.login-btn { 
							background-color: #7B68EE;
							color: white;
							padding: 12px 24px;
							text-decoration: none;
							border-radius: 5px;
							display: inline-block;
							font-weight: bold;
							transition: all 0.3s ease;
						}
						.login-btn:hover {
							background-color: #6A5ACD;
							transform: translateY(-2px);
							box-shadow: 0 2px 8px rgba(123,104,238,0.4);
						}
					</style>
				</head>
				<body>
					<div class="container">
						<h1>Welcome to ClickUp Authentication</h1>
						<p>Click below to authenticate with your ClickUp account</p>
						<a href="/clickup/login" class="login-btn">Login with ClickUp</a>
					</div>
				</body>
			</html>
		`)
		return
	}

	// User is authenticated
	fmt.Fprintf(w, `
		<html>
			<head>
				<title>ClickUp Auth - Welcome</title>
				<style>
					body { 
						font-family: Arial, sans-serif; 
						margin: 40px;
						background-color: #f5f5f5;
						display: flex;
						flex-direction: column;
						align-items: center;
					}
					.container {
						background-color: white;
						padding: 30px;
						border-radius: 10px;
						box-shadow: 0 2px 4px rgba(0,0,0,0.1);
						text-align: center;
					}
					.logout-btn { 
						background-color: #DC143C;
						color: white;
						padding: 12px 24px;
						text-decoration: none;
						border-radius: 5px;
						display: inline-block;
						font-weight: bold;
						transition: all 0.3s ease;
					}
					.logout-btn:hover {
						background-color: #B22222;
						transform: translateY(-2px);
						box-shadow: 0 2px 8px rgba(220,20,60,0.4);
					}
				</style>
			</head>
			<body>
				<div class="container">
					<h1>Welcome, %s!</h1>
					<p>You have successfully authenticated with ClickUp.</p>
					<a href="/clickup/logout" class="logout-btn">Logout</a>
				</div>
			</body>
		</html>
	`, userEmail)
}

func clickupLoginHandler(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s",
		clickupConfig.AuthURL,
		clickupConfig.ClientID,
		clickupConfig.RedirectURI)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func clickupCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	data := url.Values{}
	data.Set("client_id", clickupConfig.ClientID)
	data.Set("client_secret", clickupConfig.ClientSecret)
	data.Set("code", code)

	req, err := http.NewRequest("POST", clickupConfig.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		http.Error(w, "Failed to create token request", http.StatusInternalServerError)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		http.Error(w, "Failed to decode token response", http.StatusInternalServerError)
		return
	}

	// Get user info
	userReq, err := http.NewRequest("GET", clickupConfig.UserInfoURL, nil)
	if err != nil {
		http.Error(w, "Failed to create user info request", http.StatusInternalServerError)
		return
	}
	userReq.Header.Add("Authorization", tokenResponse.AccessToken)

	userResp, err := client.Do(userReq)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	var clickupUser ClickUpUser
	if err := json.NewDecoder(userResp.Body).Decode(&clickupUser); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Store user info in session
	session, _ := store.Get(r, "session")
	session.Values["email"] = clickupUser.User.Email
	session.Values["username"] = clickupUser.User.Username
	session.Values["auth_provider"] = "clickup"
	session.Save(r, w)

	http.Redirect(w, r, "/clickup", http.StatusTemporaryRedirect)
}

func clickupLogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/clickup", http.StatusTemporaryRedirect)
}