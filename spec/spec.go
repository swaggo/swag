package spec

type SwaggerSpec struct {
	BasePath string `yaml:"basePath"`
	Host     string `yaml:"host"`
	Info     struct {
		Contact struct {
			Email string `yaml:"email"`
		} `yaml:"contact"`
		Description string `yaml:"description"`
		License     struct {
			Name string `yaml:"name"`
			URL  string `yaml:"url"`
		} `yaml:"license"`
		TermsOfService string `yaml:"termsOfService"`
		Title          string `yaml:"title"`
		Version        string `yaml:"version"`
	} `yaml:"info"`
	Schemes []string `yaml:"schemes"`
	Swagger string   `yaml:"swagger"`
	Tags    []struct {
		Description  string `yaml:"description"`
		ExternalDocs struct {
			Description string `yaml:"description"`
			URL         string `yaml:"url"`
		} `yaml:"externalDocs"`
		Name string `yaml:"name"`
	} `yaml:"tags"`
}