package fdk

import (
	"context"
	"fmt"
	"net/http"
)

type routeKey struct {
	method string
	route  string
}

type Mux struct {
	routes      map[string]bool
	meth2Routes map[string]map[string]bool

	handlers map[routeKey]Handler
}

func NewMux() *Mux {
	return &Mux{
		routes:      make(map[string]bool),
		meth2Routes: make(map[string]map[string]bool),
		handlers:    make(map[routeKey]Handler),
	}
}

func (m *Mux) Handle(ctx context.Context, r Request) Response {
	route := r.URL
	if route == "" {
		route = "/"
	}

	rk := routeKey{route: route, method: r.Method}
	if !m.routes[rk.route] {
		return Response{Errors: []APIError{{Code: http.StatusNotFound, Message: "route not found"}}}
	}
	if !m.meth2Routes[rk.method][rk.route] {
		return Response{Errors: []APIError{{Code: http.StatusMethodNotAllowed, Message: "method not allowed"}}}
	}

	h := m.handlers[rk] // checks above guarantee this exists here
	return h.Handle(ctx, r)
}

func (m *Mux) Delete(route string, h Handler) {
	m.registerRoute(http.MethodDelete, route, h)
}

func (m *Mux) Get(route string, h Handler) {
	m.registerRoute(http.MethodGet, route, h)
}

func (m *Mux) Post(route string, h Handler) {
	m.registerRoute(http.MethodPost, route, h)
}

func (m *Mux) Put(route string, h Handler) {
	m.registerRoute(http.MethodPut, route, h)
}

func (m *Mux) registerRoute(method, route string, h Handler) {
	// TODO(berg): add additional checks for validity of method/route pairs
	if route == "" {
		panic("route must be provided")
	}
	if h == nil {
		panic("handler must not be nil")
	}

	rk := routeKey{route: route, method: method}
	if _, ok := m.handlers[rk]; ok {
		panic(fmt.Sprintf("multiple handlers added for: %q ", method+" "+route))
	}

	m.routes[route] = true

	m2r := m.meth2Routes[method]
	if m2r == nil {
		m2r = map[string]bool{}
	}
	m2r[route] = true
	m.meth2Routes[method] = m2r

	m.handlers[rk] = h
}
