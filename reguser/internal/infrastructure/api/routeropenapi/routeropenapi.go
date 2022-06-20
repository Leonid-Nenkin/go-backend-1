package routeropenapi

import (
	"encoding/json"
	"fmt"
	"goback1/lesson7/reguser/internal/infrastructure/api/auth"
	"goback1/lesson7/reguser/internal/infrastructure/api/handler"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type RouterOpenAPI struct {
	*chi.Mux
	hs *handler.Handlers
}

func NewRouterOpenAPI(hs *handler.Handlers) *RouterOpenAPI {
	r := chi.NewRouter()
	r.Use(auth.AuthMiddleware)

	ret := &RouterOpenAPI{
		hs: hs,
	}

	r.Mount("/", Handler(ret))

	swg, err := GetSwagger()
	if err != nil {
		log.Fatal("swagger fail")
	}

	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		enc := json.NewEncoder(w)
		_ = enc.Encode(swg)
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

func (rt *RouterOpenAPI) PostCreate(w http.ResponseWriter, r *http.Request) {
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

func (rt *RouterOpenAPI) GetReadId(w http.ResponseWriter, r *http.Request, sid string) {
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

func (rt *RouterOpenAPI) DeleteDeleteId(w http.ResponseWriter, r *http.Request, sid string) {
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

func (rt *RouterOpenAPI) FindUsers(w http.ResponseWriter, r *http.Request, q string) {
	fmt.Fprintln(w, "[")
	comma := false
	err := rt.hs.SearchUser(r.Context(), q, func(u handler.User) error {
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
