package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	"github.com/soffokl/freelance/freelance"
)

func main() {
	ex := freelance.NewExchange()
	defer ex.Close()

	r := mux.NewRouter()

	r.HandleFunc("/users/", ex.UserList).Methods(http.MethodGet)
	r.HandleFunc("/users/", ex.UserAdd).Methods(http.MethodPost)

	r.HandleFunc("/orders/", ex.OrderList).Methods(http.MethodGet)
	r.HandleFunc("/orders/", ex.OrderAdd).Methods(http.MethodPost)
	r.HandleFunc("/orders/{order_id}/{status}", ex.OrderUpdate).Methods(http.MethodPut)

	server := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	server.Shutdown(ctx)
}
