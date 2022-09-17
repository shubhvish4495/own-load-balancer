package model

import (
	"errors"
	"log"
	"net/http"
)

var client *http.Client

func init(){
	client = &http.Client{}
}

//service interface
// type ServiceInterface interface {
// 	//Ping method to Ping service
// 	Ping() error

// 	//ForwardRequest to forward the request
// 	ForwardRequest() error
// }


type Service struct {
	ServiceHost string
	IsServiceAvailable bool
	ServiceErrorContinuousErrCount int
}

// Ping the requested service to check health
func (s *Service) Ping() error {

	resp, err := http.Get(s.ServiceHost + "/ping")

	// if error is there we consider the service to be down
	if err != nil {
		log.Printf("Error while pinging service host %s",  s.ServiceHost)
		return err
	}

	// check and raise error if underlying service is down
	if resp.StatusCode != http.StatusOK {
		log.Printf("Service %s is down.", s.ServiceHost)
		return errors.New("Service down")
	}

	// normal case
	return nil
}

func (s *Service) ForwardRequest(r *http.Request) (*http.Response ,error) {
	if err := s.Ping(); err != nil {
		return nil,err
	}

	// create new request
	request, err := http.NewRequest(r.Method, s.ServiceHost + r.URL.RequestURI() ,r.Body)
	if err != nil {
		log.Println("Error while creating request ", err)
		return nil, err
	}

	// forward the request to service
	resp,err := client.Do(request)
	if err != nil {
		log.Println("Error while creating request ", err)
		return nil, err
	}

	return resp, nil
}