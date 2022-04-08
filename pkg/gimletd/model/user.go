package model

// User is the user representation
type User struct {
	// ID for this user
	// required: true
	ID int64 `json:"-"  meddler:"id,pk"`

	// Login is the username for this user
	// required: true
	Login string `json:"login"  meddler:"login"`

	// Token is the user's api JWT token - not persisted
	Token string `json:"token"  meddler:"-"`

	// Secret is the key used to sign JWT and CSRF tokens
	Secret string `json:"-" meddler:"secret"`

	// If the user is admin
	Admin bool `json:"admin"  meddler:"admin"`
}
