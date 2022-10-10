package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/shubhvish4495/own-load-balancer/pkg/middleware"
	"github.com/shubhvish4495/own-load-balancer/pkg/model"
)

var (
	lb *model.LoadBalancer
)

func init() {
	lb = model.NewLoadBalancer()
}

func main() {
	r := mux.NewRouter()

	// setup logger middleware
	r.Use(middleware.LogRequest)

	r.Path(`/{adummy:.*}`).HandlerFunc(lb.RouteIncomingRequestRoundRobin)


	fmt.Printf("Initial lb address %p\n", lb)

	srv := &http.Server{
		Handler: r,
		Addr:    ":4444",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// this function marks down services as unhealthy
	go func() {
		for range time.Tick(time.Second * lb.ServiceRefreshTime) {
			lb.RemoveUnhealthyServices()
		}
	} ()

	// // this function removes unnecessary locks
	// go func() {
	// 	for range time.Tick(time.Second * 1) {
	// 		lb.RemoveLockedServices()
	// 	}
	// } ()

	log.Printf("Listening on 4444")
	log.Fatal(srv.ListenAndServe())
	

}