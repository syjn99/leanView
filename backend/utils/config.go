// Config utility functions inspired by the ethpandaops/dora project
// https://github.com/ethpandaops/dora

package utils

import (
	"fmt"
	"net/url"
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"

	"github.com/syjn99/leanView/backend/config"
	"github.com/syjn99/leanView/backend/types"
)

var Config *types.Config

func ReadConfig(cfg *types.Config, path string) error {
	err := readConfigFile(cfg, path)
	if err != nil {
		return err
	}

	readConfigEnv(cfg)

	if cfg.LeanApi.Endpoints == nil && cfg.LeanApi.Endpoint != "" {
		cfg.LeanApi.Endpoints = []types.EndpointConfig{
			{
				Url:  cfg.LeanApi.Endpoint,
				Name: "default",
			},
		}
	}
	for idx, endpoint := range cfg.LeanApi.Endpoints {
		if endpoint.Name == "" {
			url, _ := url.Parse(endpoint.Url)
			if url != nil {
				cfg.LeanApi.Endpoints[idx].Name = url.Hostname()
			} else {
				cfg.LeanApi.Endpoints[idx].Name = fmt.Sprintf("endpoint-%v", idx+1)
			}
		}
	}
	if len(cfg.LeanApi.Endpoints) == 0 {
		return fmt.Errorf("missing lean node endpoints (need at least 1 endpoint to run the explorer)")
	}

	return nil
}

func readConfigFile(cfg *types.Config, path string) error {
	if path == "" {
		return yaml.Unmarshal([]byte(config.DefaultConfigYml), cfg)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening config file %v: %v", path, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("error decoding explorer config: %v", err)
	}
	return nil
}

func readConfigEnv(cfg *types.Config) error {
	return envconfig.Process("", cfg)
}
