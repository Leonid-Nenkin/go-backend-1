package main

import (
	"context"
	"goback1/lesson7/reguser/internal/infrastructure/api/handler"
	"goback1/lesson7/reguser/internal/infrastructure/api/routeropenapi"
	"goback1/lesson7/reguser/internal/infrastructure/db/pgstore"
	"goback1/lesson7/reguser/internal/infrastructure/server"
	"goback1/lesson7/reguser/internal/usecases/app/repos/userrepo"
	"log"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	// ust := usermemstore.NewUsers()
	// ust, err := userfilemanager.NewUsers("./data.json", "mem://userRefreshTopic")
	ust, err := pgstore.NewUsers("postgres://postgres:1110@localhost/test?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	us := userrepo.NewUsers(ust)
	hs := handler.NewHandlers(us)
	// h := defmux.NewRouter(hs)
	// h := routerchi.NewRouterChi(hs)
	h := routeropenapi.NewRouterOpenAPI(hs)
	srv := server.NewServer(":8000", h)

	srv.Start(us)
	log.Print("Start")

	<-ctx.Done()

	srv.Stop()
	cancel()
	ust.Close()

	log.Print("Exit")
}
