package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"go-short-url/src/middleware"
	"gopkg.in/validator.v2"
	"log"
	"net/http"
)

// App encapsulates Config, Router and Middleware
type App struct {
	Router     *mux.Router
	Middleware *middleware.Middleware
	Config     *Env
}

// ShortenReq 短链请求体
type ShortenReq struct {
	URL                 string `json:"url" validate:"nonzero"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate:"min=0"`
}

// ShortlinkResp 短链响应体
type ShortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

// Initialize is initialization of App
func (a *App) Initialize(e *Env) {
	// Set log formatter
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Config = e
	a.Router = mux.NewRouter()
	a.Middleware = &middleware.Middleware{}
	a.InitializeRoutes()
}

func (a *App) InitializeRoutes() {
	//a.Router.HandleFunc("/api/shorten", a.CreateShortlink).Methods("POST")
	//a.Router.HandleFunc("/api/info", a.GetShortlinkInfo).Methods("GET")
	//a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.Redirect).Methods("GET")

	// Alice包的使用，合并中间件
	m := alice.New(a.Middleware.LoggingHandler, a.Middleware.RecoverHandler)
	a.Router.Handle("/api/shorten", m.ThenFunc(a.CreateShortlink)).Methods("POST")
	a.Router.Handle("/api/info", m.ThenFunc(a.GetShortlinkInfo)).Methods("GET").Queries("url", "{url}")
	a.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}}", m.ThenFunc(a.Redirect)).Methods("GET")
}

func (a *App) CreateShortlink(writer http.ResponseWriter, request *http.Request) {
	var req ShortenReq
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		ResponseWithError(writer, StatusError{http.StatusBadRequest, fmt.Errorf("parse parameters failed %v", request.Body)})
		return
	}
	if err := validator.Validate(req); err != nil {
		ResponseWithError(writer, StatusError{http.StatusBadRequest, fmt.Errorf("validate parameters failed %v", req)})
		return
	}
	defer request.Body.Close()
	//fmt.Printf("%v\n", req)

	s, err := a.Config.S.Shorten(req.URL, req.ExpirationInMinutes)
	if err != nil {
		ResponseWithError(writer, err)
	} else {
		ResponseWithJSON(writer, http.StatusCreated, ShortlinkResp{Shortlink: s})
	}
}

func (a *App) GetShortlinkInfo(writer http.ResponseWriter, request *http.Request) {
	parameters := request.URL.Query()
	shortUrl := parameters.Get("url")
	//fmt.Printf("%s\n", shortUrl)
	//panic(shortUrl) // 手动panic
	detail, err := a.Config.S.ShortlinkInfo(shortUrl)
	if err != nil {
		ResponseWithError(writer, err)
	} else {
		ResponseWithJSON(writer, http.StatusOK, detail)
	}
}

func (a *App) Redirect(writer http.ResponseWriter, request *http.Request) {
	parameters := mux.Vars(request)
	//fmt.Printf("%s\n", parameters["shortlink"])
	u, err := a.Config.S.Unshorten(parameters["shortlink"])
	if err != nil {
		ResponseWithError(writer, err)
	} else {
		http.Redirect(writer, request, u, http.StatusTemporaryRedirect)
	}
}

func ResponseWithError(writer http.ResponseWriter, err error) {
	switch e := err.(type) {
	case Error:
		log.Printf("HTTP %d - %s", e.Status(), e)
		ResponseWithJSON(writer, e.Status(), e.Error())
	default:
		ResponseWithJSON(writer, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func ResponseWithJSON(writer http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	writer.Write(resp)
}

// Run start listen and server
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
