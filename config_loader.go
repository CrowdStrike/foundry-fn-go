package fdk

import (
	"context"
	"errors"
	"fmt"
	"os"
)

var (
	ErrCfgNotFound = errors.New("no config provided")
)

type ConfigLoader interface {
	LoadConfig(ctx context.Context) ([]byte, error)
}

func RegisterConfigLoader(loaderType string, cr ConfigLoader) {
	if _, ok := configReaders[loaderType]; ok {
		panic(fmt.Sprintf("config loader type already exists: %q", loaderType))
	}

	configReaders[loaderType] = cr
}

func loadConfigBytes(ctx context.Context) ([]byte, error) {
	crt := os.Getenv("CONFIG_LOADER_TYPE")
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
