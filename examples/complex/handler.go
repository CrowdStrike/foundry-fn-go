package main

import (
	"context"
	"log/slog"
	"net/http"

	fdk "github.com/CrowdStrike/foundry-fn-go"
)

type config struct {
	IncludeMonsterDrivers bool `json:"include_monster_drivers"`
}

func (c config) OK() error {
	return nil
}

func newHandler(_ context.Context, logger *slog.Logger, cfg config) fdk.Handler {
	mux := fdk.NewMux()

	h := handler{
		includeDrivers: cfg.IncludeMonsterDrivers,
		repo:           newPeopleRepo(),
	}
	h.registerRoutes(mux)

	out := loggingMW(logger)(mux)
	return out
}

type (
	repo interface {
		CreatePerson(ctx context.Context, p person) error
		ReadPeople(ctx context.Context, names ...string) ([]person, error)
		CreateMonster(ctx context.Context, monster monster) error
		ReadMonsters(ctx context.Context, name ...string) ([]monster, error)
	}

	person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	monster struct {
		Name        string `json:"name"`
		ThreatLevel string `json:"threat_level"`
	}
)

type handler struct {
	includeDrivers bool
	repo           repo
}

func (h *handler) registerRoutes(mux *fdk.Mux) {
	mux.Get("/people", fdk.HandlerFn(h.getPeople))
	mux.Post("/people", fdk.HandleFnOf(h.createPerson))
}

func (h *handler) getPeople(ctx context.Context, r fdk.Request) fdk.Response {
	names := r.Queries["name"]

	people, err := h.repo.ReadPeople(ctx, names...)
	if err != nil {
		return fdk.Response{Errors: []fdk.APIError{{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}}}
	}
	return fdk.Response{
		Body: fdk.JSON(people),
		Code: 200,
	}
}

type createPersonReq struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (h *handler) createPerson(ctx context.Context, r fdk.RequestOf[createPersonReq]) fdk.Response {
	if threatLevel, ok := h.isMonster(r.Body); ok {
		return h.createMonster(ctx, r.Body, threatLevel)
	}

	p := person{Name: r.Body.Name, Age: r.Body.Age}
	err := h.repo.CreatePerson(ctx, p)
	if err != nil {
		return fdk.Response{
			Errors: []fdk.APIError{{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}},
		}
	}

	return fdk.Response{
		Code: http.StatusCreated,
		Body: fdk.JSON(p),
	}
}

func (h *handler) createMonster(ctx context.Context, r createPersonReq, threatLevel string) fdk.Response {
	err := h.repo.CreateMonster(ctx, monster{
		Name:        r.Name,
		ThreatLevel: threatLevel,
	})
	if err != nil {
		return fdk.Response{
			Errors: []fdk.APIError{{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}},
		}
	}
	return fdk.Response{
		Errors: []fdk.APIError{{
			Code:    http.StatusForbidden,
			Message: "monster identified",
		}},
	}
}

func (h *handler) isMonster(r createPersonReq) (string, bool) {
	if isToddler := r.Age < 4; isToddler {
		return "APOCALYPTIC", true
	}

	maxTeenAge := 19
	if !h.includeDrivers {
		maxTeenAge = 15
	}
	if isTeenager := r.Age > 12 && r.Age <= maxTeenAge; isTeenager {
		return "CATASTROPHIC", true
	}
	return "", false
}
