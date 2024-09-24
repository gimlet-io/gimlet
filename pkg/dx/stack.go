package dx

type StackRef struct {
	Repository string `yaml:"repository" json:"repository"`
}

type StackConfig struct {
	Stack  StackRef               `yaml:"stack" json:"stack"`
	Config map[string]interface{} `yaml:"config" json:"config"`
}

type PlainModule struct {
	URL      string                   `json:"url"`
	Schema   map[string]interface{}   `json:"schema"`
	UISchema []map[string]interface{} `json:"uiSchema"`
	Template string                   `json:"-"`
}

type Component struct {
	Name        string `json:"name,omitempty" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description"`
	Category    string `json:"category,omitempty" yaml:"category"`
	Variable    string `json:"variable,omitempty" yaml:"variable"`
	Logo        string `json:"logo,omitempty" yaml:"logo"`
	OnePager    string `json:"onePager,omitempty" yaml:"onePager"`
	Schema      string `json:"schema,omitempty" yaml:"schema"`
	UISchema    string `json:"uiSchema,omitempty" yaml:"uiSchema"`
}

type StackDefinition struct {
	Name        string        `json:"name,omitempty" yaml:"name"`
	Description string        `json:"description,omitempty" yaml:"description"`
	Intro       string        `json:"intro,omitempty" yaml:"intro"`
	Categories  []interface{} `json:"categories" yaml:"categories"`
	Components  []*Component  `json:"components,omitempty" yaml:"components"`
	ChangLog    string        `json:"changeLog,omitempty" yaml:"changeLog"`
	Message     string        `json:"message,omitempty" yaml:"message"`
}
