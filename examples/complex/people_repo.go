package main

import (
	"context"
	"errors"
)

type peopleRepo struct {
	people   map[string]person
	monsters map[string]monster
}

func newPeopleRepo() *peopleRepo {
	return &peopleRepo{
		people:   make(map[string]person),
		monsters: make(map[string]monster),
	}
}

func (pr *peopleRepo) CreatePerson(ctx context.Context, p person) error {
	return create(pr.people, p.Name, p)
}

func (pr *peopleRepo) ReadPeople(ctx context.Context, names ...string) ([]person, error) {
	return read(pr.people, "people", names)
}

func (pr *peopleRepo) CreateMonster(ctx context.Context, m monster) error {
	return create(pr.monsters, m.Name, m)

}

func (pr *peopleRepo) ReadMonsters(ctx context.Context, names ...string) ([]monster, error) {
	return read(pr.monsters, "monsters", names)
}

func create[T any](m map[string]T, name string, v T) error {
	if _, ok := m[name]; ok {
		return errors.New("resource exists")
	}
	m[name] = v

	return nil
}

func read[T any](m map[string]T, resource string, names []string) ([]T, error) {
	var out []T
	for _, n := range names {
		if v, ok := m[n]; ok {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("did not find any " + resource)
	}

	return out, nil
}
