package middlewares

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)


func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		requestURI := request.URL.RequestURI()
		host := request.Host
		requestUUID := uuid.NewString()

		log.Println(fmt.Sprintf(`{"requestUUID" : "%s", "method" : "%s", "url": "%s", "host" : "%s"}`, requestUUID, request.Method, requestURI, host))

		next.ServeHTTP(writer, request)
	})
}