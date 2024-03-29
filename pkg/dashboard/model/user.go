package model

// User is the user representation
type User struct {
	// ID for this user
	// required: true
	ID int64 `json:"-" meddler:"id,pk"`

	// Login is the username for this user
	// required: true
	Login string `json:"login"  meddler:"login"`

	// Token is the user's api JWT token - not persisted
	Token string `json:"token"  meddler:"-"`

	// If the user is admin
	Admin bool `json:"admin"  meddler:"admin"`

	// Name is the full name for this user
	Name string `json:"name"  meddler:"name"`

	// Login is the username for this user
	// required: true
	Email string `json:"-"  meddler:"email"`

	// GithubToken is the Github oauth token
	AccessToken string `json:"-"  meddler:"access_token,encrypted"`

	// RefreshToken is the Github refresh token
	RefreshToken string `json:"-"  meddler:"refresh_token,encrypted"`

	// Expires is the Github token expiry date
	Expires int64 `json:"-"  meddler:"expires"`

	// Secret is the PEM formatted RSA private key used to sign JWT and CSRF tokens
	Secret string `json:"-" meddler:"secret,encrypted"`

	Repos []string `json:"-" meddler:"repos,json"`

	FavoriteRepos []string `json:"favoriteRepos"  meddler:"favorite_repos,json"`

	FavoriteServices []string `json:"favoriteServices"  meddler:"favorite_services,json"`
}
