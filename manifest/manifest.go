package manifest

type Manifest struct {
	App       string                 `yaml:"app"`
	Env       string                 `yaml:"env"`
	Namespace string                 `yaml:"namespace"`
	Chart     Chart                  `yaml:"chart"`
	Values    map[string]interface{} `yaml:"values"`
}

type Chart struct {
	Repository string `yaml:"repository"`
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
}
