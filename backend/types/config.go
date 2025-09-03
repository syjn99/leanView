package types

type Config struct {
	LeanApi struct {
		Endpoint  string           `yaml:"endpoint" envconfig:"LEANAPI_ENDPOINT"`
		Endpoints []EndpointConfig `yaml:"endpoints"`
	} `yaml:"leanapi"`
}

type EndpointConfig struct {
	Url  string `yaml:"url"`
	Name string `yaml:"name"`
}
