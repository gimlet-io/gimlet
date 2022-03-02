package dx

type StackRef struct {
	Repository string `yaml:"repository" json:"repository"`
}

type StackConfig struct {
	Stack  StackRef               `yaml:"stack" json:"stack"`
	Config map[string]interface{} `yaml:"config" json:"config"`
}
