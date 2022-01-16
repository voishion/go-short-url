package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"gopkg.in/validator.v2"
	"log"
	"net/http"
)

// App encapsulates Env, Router and middleware
type App struct {
	Router *mux.Router
}

// shortenReq 短链请求体
type shortenReq struct {
	URL                 string `json:"url" validate:"nonzero"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

// shortlinkResp 短链响应体
type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

// Initialize is initialization of App
func (a *App) Initialize() {
	// Set log formatter
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/api/shorten", a.createShortlink).Methods("POST")
	a.Router.HandleFunc("/api/info", a.getShortlinkInfo).Methods("POST")
	a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.redirect).Methods("GET")
}

func (a *App) createShortlink(writer http.ResponseWriter, request *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		return
	}
	if err := validator.Validate(req); err != nil {
		return
	}
	defer request.Body.Close()
	fmt.Printf("%v\n", req)
}

func (a *App) getShortlinkInfo(writer http.ResponseWriter, request *http.Request) {
	parameters := request.URL.Query()
	shortlink := parameters.Get("shortlink")
	fmt.Printf("%s\n", shortlink)
}

func (a *App) redirect(writer http.ResponseWriter, request *http.Request) {
	parameters := mux.Vars(request)
	fmt.Printf("%s\n", parameters["shortlink"])
}

// Run start listen and server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
