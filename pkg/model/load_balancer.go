package model

import (
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type LoadBalancer struct {
	ServiceList []Service
	RoundRobinCounter int
	RRMuLock *sync.Mutex
	ServiceRefreshTime time.Duration
	MaxUnhealtyCounter int
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		ServiceList: []Service{
			{
				ServiceHost: "http://localhost:4445",
				IsServiceAvailable: true,
				ServiceLock: &sync.Mutex{},
			},
			{
				ServiceHost: "http://localhost:4446",
				IsServiceAvailable: true,
				ServiceLock: &sync.Mutex{},
			},			
		},
		ServiceRefreshTime: time.Duration(5),
		RRMuLock: &sync.Mutex{},
		MaxUnhealtyCounter: 5,
	}
}


// RouteIncomingRequestRoundRobin routes the incoming request to the first 
// available server. It follows round robin technique for server 
// allocation.
func (lb *LoadBalancer) RouteIncomingRequestRoundRobin(rw http.ResponseWriter, r *http.Request) {
	var err error
	var resp *http.Response

	// get the next available server in line
	lb.RRMuLock.Lock()
	i := lb.RoundRobinCounter

	//var to determine whether to break loop or not
	loopBreaker := 0

	// iterate over all the available servers
	for {
		currService := lb.ServiceList[i]
		currService.ServiceLock.Lock()

		// if current service is not available then skip it
		if currService.IsServiceAvailable {
			resp, err = currService.ForwardRequest(r)
			if err == nil {
				rw.WriteHeader(resp.StatusCode)
				respDecoded, _ := io.ReadAll(resp.Body)
				rw.Write(respDecoded)
				//reset the continuous err count
				lb.ServiceList[i].ServiceErrorContinuousErrCount = 0
				// currService.ServiceLock.Unlock()
				loopBreaker = 1
			}
		} else {
			err = errors.New("current service is down")
			log.Println("Service down. Moving forward to other service")

			// increment the continuous err count
			lb.ServiceList[i].ServiceErrorContinuousErrCount += 1
			i = (i + 1) % len(lb.ServiceList)

			if i == lb.RoundRobinCounter {
				// currService.ServiceLock.Unlock()
				loopBreaker = 1
			}
		}

		// unlock the service and break out of loop if 
		// loopBreaker counter is set
		currService.ServiceLock.Unlock()
		if loopBreaker == 1{
			break
		}
	}

	// if no server are available then we send back 500 error
	if err != nil {
		log.Println("No service is up for taking request")
		rw.WriteHeader(http.StatusInternalServerError)
	}

	nxtCtrVal := (i + 1) % len(lb.ServiceList)

	// increment the counter for next usage
	lb.RoundRobinCounter = nxtCtrVal
	lb.RRMuLock.Unlock()
}


// RemoveUnhealthyServices removes unhealthy service
func (lb *LoadBalancer) RemoveUnhealthyServices() {
	sl := lb.ServiceList

	
	for i := 0; i < len(sl);  i++ {
		service := &sl[i]
		service.ServiceLock.Lock()
		err := service.Ping()
		if err != nil {
			// if service is thorwing error for continuously for a threshold value
			// then mark the service as unhealthy.
			if service.ServiceErrorContinuousErrCount >= lb.MaxUnhealtyCounter && service.IsServiceAvailable {
				log.Printf("service %s is unhealthy, marking it as unhealthy", service.ServiceHost)
				service.IsServiceAvailable = false
			}
			service.ServiceErrorContinuousErrCount += 1
		} else {
				// if service is back up, then mark it as available with 0
				// continuous errors
				service.IsServiceAvailable = true
				service.ServiceErrorContinuousErrCount = 0
		}
		service.ServiceLock.Unlock()
	}

	
}