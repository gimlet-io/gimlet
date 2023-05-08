package model

type Config struct {
	ID    int64  `json:"id" meddler:"id,pk"`
	Key   string `json:"key"  meddler:"key"`
	Value string `json:"value"  meddler:"value"`
}
