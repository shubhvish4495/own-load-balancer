package model

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type LoadBalancer struct {
	ServiceList []Service
	RoundRobinCounter int
	ServiceRefreshTime int
}


// RouteIncomingRequest routes the incoming request to the first 
// available server. It follows round robin technique for server 
// allocation.
func (lb *LoadBalancer) RouteIncomingRequest(rw http.ResponseWriter, r *http.Request) {
	var err error
	var resp *http.Response

	// get the next available server in line 
	i := lb.RoundRobinCounter

	// iterate over all the available servers
	for {
		currService := lb.ServiceList[i]

		// if current service is not available then skip it
		if currService.IsServiceAvailable {
			resp, err = currService.ForwardRequest(r)
			if err == nil {
				rw.WriteHeader(resp.StatusCode)
				respDecoded, _ := ioutil.ReadAll(resp.Body)
				rw.Write(respDecoded)
				//reset the continuous err count
				lb.ServiceList[i].ServiceErrorContinuousErrCount = 0
				break
			}
		} else {
			err = errors.New("current service is down")
		}
		log.Println("Service down. Moving forward to other service")
		// increment the continuous err count
		lb.ServiceList[i].ServiceErrorContinuousErrCount += 1
		i = (i + 1) % len(lb.ServiceList)

		if i == lb.RoundRobinCounter {
			break
		}
	}

	// if no server are available then we send back 500 error
	if err != nil {
		log.Println("No service is up for taking request")
		rw.WriteHeader(http.StatusInternalServerError)
	}

	// increment the counter for next usage
	lb.RoundRobinCounter = (lb.RoundRobinCounter + 1) % len(lb.ServiceList)

}


// RemoveUnhealthyServices removes unhealthy service
func (lb * LoadBalancer) RemoveUnhealthyServices() {
	sl := lb.ServiceList

	for i := 0; i < len(sl);  i++ {
		service := &sl[i]
	
		err := service.Ping()
		if err != nil {
			// if service is thorwing error for continuously for a threshold value
			// then mark the service as unhealthy.
			if service.ServiceErrorContinuousErrCount >= 5 {
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
	}
}