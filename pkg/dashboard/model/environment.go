package model

type Environment struct {
	ID                  int64  `json:"-" meddler:"id,pk"`
	Name                string `json:"name"  meddler:"name"`
	RepoPerEnv          bool   `json:"repoPerEnv"  meddler:"repo_per_env"`
	KustomizationPerApp bool   `json:"kustomizationPerApp"  meddler:"kustomization_per_app"`
	InfraRepo           string `json:"infraRepo"  meddler:"infra_repo"`
	AppsRepo            string `json:"appsRepo"  meddler:"apps_repo"`
	BuiltIn             bool   `json:"builtIn"  meddler:"built_in"`
}
