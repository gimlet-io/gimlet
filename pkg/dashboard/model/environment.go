package model

type Environment struct {
	ID   int64  `json:"-" meddler:"id,pk"`
	Name string `json:"name"  meddler:"name"`
}
