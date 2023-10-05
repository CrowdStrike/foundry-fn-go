package fdk

import (
	"context"
	"encoding/json"
	"net/http"
)

// Cfg marks the configuration type parameter. Any config
// type must have a validation method, OK, defined on it.
type Cfg interface {
	OK() error
}

// SkipCfg indicates the config is not needed and will skip
// the config loading procedure.
type SkipCfg struct{}

// OK is a noop validation.
func (n SkipCfg) OK() error {
	return nil
}

type cfgErr struct {
	err    error
	apiErr APIError
}

func readCfg[T Cfg](ctx context.Context) (T, *cfgErr) {
	var cfg T
	switch any(cfg).(type) {
	// exceptional case, where a func does not need/want a config
	// otherwise, we'll decode into the target type
	case SkipCfg, *SkipCfg:
		return *new(T), nil
	}

	cfgB, err := loadConfigBytes(ctx)
	if err != nil {
		return *new(T), &cfgErr{
			err:    err,
			apiErr: APIError{Code: http.StatusInternalServerError, Message: "failed to read config source"},
		}
	}

	err = json.Unmarshal(cfgB, &cfg)
	if err != nil {
		return *new(T), &cfgErr{
			err:    err,
			apiErr: APIError{Code: http.StatusBadRequest, Message: "failed to unmarshal config into config type"},
		}

	}

	err = cfg.OK()
	if err != nil {
		return *new(T), &cfgErr{
			apiErr: APIError{Code: http.StatusBadRequest, Message: "config is invalid: " + err.Error()},
		}
	}

	return cfg, nil
}
