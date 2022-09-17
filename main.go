package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/shubhvish4495/own-load-balancer/pkg/middleware"
	"github.com/shubhvish4495/own-load-balancer/pkg/model"
)

var (
	lb model.LoadBalancer
)

func init() {
	lb = model.LoadBalancer{
		ServiceList: []model.Service{
			{
				ServiceHost: "http://localhost:4445",
				IsServiceAvailable: true,
			},
			{
				ServiceHost: "http://localhost:4446",
				IsServiceAvailable: true,
			},			
		},
	}
}


func main() {
	r := mux.NewRouter()

	// setup logger middleware
	r.Use(middleware.LogRequest)

	r.Path(`/{adummy:.*}`).HandlerFunc(lb.RouteIncomingRequest)


	srv := &http.Server{
		Handler: r,
		Addr:    ":4444",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// this function marks down services as unhealthy
	go func() {
		for range time.Tick(time.Second * 5) {
			log.Println("Removing unhealthy service")
			lb.RemoveUnhealthyServices()
		}
	} ()

	log.Printf("Listening on 4444")
	log.Fatal(srv.ListenAndServe())
	

}