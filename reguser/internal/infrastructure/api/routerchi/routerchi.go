package routerchi

import (
	"fmt"
	"goback1/lesson7/reguser/internal/infrastructure/api/auth"
	"goback1/lesson7/reguser/internal/infrastructure/api/handler"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type RouterChi struct {
	*chi.Mux
	hs *handler.Handlers
}

func NewRouterChi(hs *handler.Handlers) *RouterChi {
	r := chi.NewRouter()
	ret := &RouterChi{
		hs: hs,
	}

	r.Group(func(ur chi.Router) {
		ur.Use(auth.AuthMiddleware)

		ur.Post("/create", ret.CreateUser)
		ur.Get("/read/{id}", ret.ReadUser)
		ur.Delete("/delete/{id}", ret.DeleteUser)
		ur.Get("/search/{q}", ret.SearchUser)
	})

	ret.Mux = r
	return ret
}

type User handler.User

func (User) Bind(r *http.Request) error {
	return nil
}

func (User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (rt *RouterChi) CreateUser(w http.ResponseWriter, r *http.Request) {
	ru := User{}
	if err := render.Bind(r, &ru); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	u, err := rt.hs.CreateUser(r.Context(), handler.User(ru))
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Render(w, r, User(u))
}

func (rt *RouterChi) ReadUser(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "id")

	uid, err := uuid.Parse(sid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	u, err := rt.hs.ReadUser(r.Context(), uid)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Render(w, r, User(u))
}

func (rt *RouterChi) DeleteUser(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "id")

	uid, err := uuid.Parse(sid)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	u, err := rt.hs.DeleteUser(r.Context(), uid)
	if err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}

	render.Render(w, r, User(u))
}

func (rt *RouterChi) SearchUser(w http.ResponseWriter, r *http.Request) {
	q := chi.URLParam(r, "q")

	fmt.Fprintln(w, "[")
	comma := false
	err := rt.hs.SearchUser(r.Context(), q,
		func(u handler.User) error {
			if comma {
				fmt.Fprintln(w, ",")
			} else {
				comma = true
			}
			if err := render.Render(w, r, User(u)); err != nil {
				return err
			}
			w.(http.Flusher).Flush()
			return nil
		})
	if err != nil {
		if comma {
			fmt.Fprint(w, ",")
		}
		render.Render(w, r, ErrRender(err))
	}
	fmt.Fprintln(w, "]")
}
