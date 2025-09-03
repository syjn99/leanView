package types

type Config struct {
	LeanApi struct {
		Endpoint  string           `yaml:"endpoint" envconfig:"LEANAPI_ENDPOINT"`
		Endpoints []EndpointConfig `yaml:"endpoints"`
	} `yaml:"leanapi"`

	Database DatabaseConfig `yaml:"database"`
}

type EndpointConfig struct {
	Url  string `yaml:"url"`
	Name string `yaml:"name"`
}

type DatabaseConfig struct {
	File         string `yaml:"file" envconfig:"DATABASE_FILE"`
	MaxOpenConns int    `yaml:"maxOpenConns" envconfig:"DATABASE_MAX_OPEN_CONNS"`
	MaxIdleConns int    `yaml:"maxIdleConns" envconfig:"DATABASE_MAX_IDLE_CONNS"`
}
