# ClickUp GitHub Authentication

This is a Go web application that demonstrates OAuth2 authentication with both GitHub and ClickUp.

## Features

- GitHub OAuth2 authentication
- ClickUp OAuth2 authentication
- Session management
- Clean and responsive UI

## Prerequisites

- Go 1.23 or higher
- GitHub OAuth application credentials
- ClickUp OAuth application credentials

## Setup

1. Clone the repository
```bash
git clone https://github.com/Akhildas-ts/clickupgithubAUTH.git
cd clickupgithubAUTH
```

2. Copy the example environment file and update with your credentials
```bash
cp .env.example .env
```

3. Update the `.env` file with your OAuth credentials:
- GitHub Client ID and Secret
- ClickUp Client ID and Secret
- ClickUp Redirect URI
- Generate a random session key

## Running the Application

1. Install dependencies
```bash
go mod download
```

2. Run the application
```bash
go run .
```

The application will be available at `http://localhost:8080`

## Routes

- `/` - Home page with GitHub authentication
- `/login` - GitHub login
- `/callback` - GitHub OAuth callback
- `/logout` - Logout from GitHub
- `/clickup` - ClickUp home page
- `/clickup/login` - ClickUp login
- `/clickup/callback` - ClickUp OAuth callback
- `/clickup/logout` - Logout from ClickUp

## Security Notes

- Never commit the `.env` file containing your secrets
- Use HTTPS in production
- Regularly rotate your session keys
- Keep your dependencies updated

## License

MIT License