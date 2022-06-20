package server

import (
	"context"
	"goback1/lesson7/reguser/internal/usecases/app/repos/userrepo"
	"net/http"
	"time"
)

type Server struct {
	srv http.Server
	us  *userrepo.Users
}

func NewServer(addr string, h http.Handler) *Server {
	s := &Server{}

	s.srv = http.Server{
		Addr:              addr,
		Handler:           h,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return s
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	s.srv.Shutdown(ctx)
	cancel()
}

func (s *Server) Start(us *userrepo.Users) {
	s.us = us
	// TODO: migrations
	go s.srv.ListenAndServe()
}
