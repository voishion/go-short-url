package middleware

import (
	"log"
	"net/http"
	"time"
)

type Middleware struct {
}

// LoggingHandler log the time-consuming of http request
func (m Middleware) LoggingHandler(next http.Handler) http.Handler {
	fn := func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()
		next.ServeHTTP(writer, request)
		endTime := time.Now()
		log.Printf("[%s] %q %v", request.Method, request.URL.String(), endTime.Sub(startTime))
	}
	return http.HandlerFunc(fn)
}

// RecoverHandler recover panic
func (m Middleware) RecoverHandler(next http.Handler) http.Handler {
	fn := func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("recover from panic: %+v", err)
				http.Error(writer, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(writer, request)
	}
	return http.HandlerFunc(fn)
}
