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

// Mux defines a handler that will dispatch to a matching route/method combination. Much
// like the std lib http.ServeMux, but with slightly more opinionated route setting. We
// only support the DELETE, GET, POST, and PUT.
type Mux struct {
	routes      map[string]bool
	meth2Routes map[string]map[string]bool

	handlers map[routeKey]Handler
}

// NewMux creates a new Mux that is ready for assignment.
func NewMux() *Mux {
	return &Mux{
		routes:      make(map[string]bool),
		meth2Routes: make(map[string]map[string]bool),
		handlers:    make(map[routeKey]Handler),
	}
}

// Handle enacts the handler to process the request/response lifecycle. The mux fulfills the
// Handler interface and can dispatch to any number of sub routes.
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

// Delete creates a DELETE route.
func (m *Mux) Delete(route string, h Handler) {
	m.registerRoute(http.MethodDelete, route, h)
}

// Get creates a GET route.
func (m *Mux) Get(route string, h Handler) {
	m.registerRoute(http.MethodGet, route, h)
}

// Post creates a POST route.
func (m *Mux) Post(route string, h Handler) {
	m.registerRoute(http.MethodPost, route, h)
}

// Put creates a PUT route.
func (m *Mux) Put(route string, h Handler) {
	m.registerRoute(http.MethodPut, route, h)
}

func (m *Mux) registerRoute(method, route string, h Handler) {
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

	{
		// nil checks, make the zero value useful
		if m.routes == nil {
			m.routes = map[string]bool{}
		}
		if m.meth2Routes == nil {
			m.meth2Routes = map[string]map[string]bool{}
		}
		if m.handlers == nil {
			m.handlers = map[routeKey]Handler{}
		}
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
