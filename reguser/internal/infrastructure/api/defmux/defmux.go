package defmux

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"goback1/lesson7/reguser/internal/infrastructure/api/auth"
	"goback1/lesson7/reguser/internal/infrastructure/api/handler"
	"net/http"

	"github.com/google/uuid"
)

type Router struct {
	*http.ServeMux
	hs *handler.Handlers
}

func NewRouter(hs *handler.Handlers) *Router {
	r := &Router{
		ServeMux: http.NewServeMux(),
		hs:       hs,
	}

	r.Handle("/create",
		// r.AuthMiddleware(
		auth.AuthMiddleware(
			http.HandlerFunc(r.CreateUser),
		),
		// ),
	)
	r.Handle("/read", auth.AuthMiddleware(http.HandlerFunc(r.ReadUser)))
	r.Handle("/delete", auth.AuthMiddleware(http.HandlerFunc(r.DeleteUser)))
	r.Handle("/search", auth.AuthMiddleware(http.HandlerFunc(r.SearchUser)))
	return r
}

func (rt *Router) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	u := handler.User{}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.hs.CreateUser(r.Context(), u)
	if err != nil {
		http.Error(w, "error when creating", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(nbu)
}

// read?uid=...
func (rt *Router) ReadUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.hs.ReadUser(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(nbu)
}

func (rt *Router) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	suid := r.URL.Query().Get("uid")
	if suid == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := uuid.Parse(suid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if (uid == uuid.UUID{}) {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	nbu, err := rt.hs.DeleteUser(r.Context(), uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "error when reading", http.StatusInternalServerError)
		}
		return
	}

	_ = json.NewEncoder(w).Encode(nbu)
}

// /search?q=...
func (rt *Router) SearchUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "bad method", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)

	first := true
	fmt.Fprintf(w, "[")
	defer fmt.Fprintln(w, "]")

	err := rt.hs.SearchUser(r.Context(), q, func(u handler.User) error {
		if first {
			first = false
		} else {
			fmt.Fprintf(w, ",")
		}
		_ = enc.Encode(u)
		w.(http.Flusher).Flush()
		return nil
	})
	if err != nil {
		// FIXME: отправлять объект ошибки json
		http.Error(w, "error when reading", http.StatusInternalServerError)
		return
	}
}
