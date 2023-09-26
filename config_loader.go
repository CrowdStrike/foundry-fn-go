package fdk

import (
	"context"
	"errors"
	"fmt"
	"os"
)

var (
	// ErrCfgNotFound defines the inability to find a config at the expected location.
	ErrCfgNotFound = errors.New("no config provided")
)

// ConfigLoader defines the behavior for loading config.
type ConfigLoader interface {
	LoadConfig(ctx context.Context) ([]byte, error)
}

// RegisterConfigLoader will register a config loader at the specified type. Similar to registering
// a database with the database/sql, you're able to provide a config for use at runtime. During Run,
// the config loader defined by the env var, CS_CONFIG_LOADER_TYPE, is used. If one is not provided,
// then the fs config loader will be used.
func RegisterConfigLoader(loaderType string, cr ConfigLoader) {
	if _, ok := configReaders[loaderType]; ok {
		panic(fmt.Sprintf("config loader type already exists: %q", loaderType))
	}

	configReaders[loaderType] = cr
}

func loadConfigBytes(ctx context.Context) ([]byte, error) {
	crt := os.Getenv("CS_CONFIG_LOADER_TYPE")
	if crt == "" {
		crt = "fs"
	}

	loader := configReaders[crt]
	if loader == nil {
		panic(fmt.Sprintf("unmatched config loader type provided: %q", crt))
	}

	return loader.LoadConfig(ctx)
}

var configReaders = map[string]ConfigLoader{
	"fs": new(localCfgLoader),
}

type localCfgLoader struct{}

func (*localCfgLoader) LoadConfig(ctx context.Context) ([]byte, error) {
	b, err := os.ReadFile(os.Getenv("CS_FN_CONFIG_PATH"))
	if os.IsNotExist(err) {
		return nil, ErrCfgNotFound
	}
	return b, err
}
