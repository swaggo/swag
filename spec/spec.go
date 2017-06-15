package spec

const SwaggerVesion = "2.0"

func New() *SwaggerSpec {
	return &SwaggerSpec{}
}

type SwaggerSpec struct {
	Swagger  string `yaml:"swagger"`
	BasePath string `yaml:"basePath"`
	Host     string `yaml:"host"`
	Info     struct {
		Contact struct {
			Email string `yaml:"email"`
			Name  string `yaml:"name"`
			URL   string `yaml:"url"`
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
	Tags    []struct {
		Description  string `yaml:"description"`
		ExternalDocs struct {
			Description string `yaml:"description"`
			URL         string `yaml:"url"`
		} `yaml:"externalDocs"`
		Name string `yaml:"name"`
	} `yaml:"tags"`
}


//TODO: fix Paths
type Get struct {
	Description string   `yaml:"description"`
	OperationID string   `yaml:"operationId"`
	Produces    []string `yaml:"produces"`

	//TODO: fix response
	Responses   struct {
		Two00 struct {
			Description string `yaml:"description"`
			Schema      struct {
				Items struct {
					Ref string `yaml:"$ref"`
				} `yaml:"items"`
				Type string `yaml:"type"`
			} `yaml:"schema"`
		} `yaml:"200"`
		Default struct {
			Description string `yaml:"description"`
			Schema      struct {
				Ref string `yaml:"$ref"`
			} `yaml:"schema"`
		} `yaml:"default"`
	} `yaml:"responses"`
	Summary string `yaml:"summary"`
}
