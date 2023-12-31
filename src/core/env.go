package main

import (
	"log"
	"os"
	"strconv"
)

type Env struct {
	S Storage
}

func GetEnv() *Env {
	addr := os.Getenv("APP_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:26379"
	}
	passwd := os.Getenv("APP_REDIS_PASSWD")
	if passwd == "" {
		passwd = "redispwd"
	}
	dbs := os.Getenv("APP_REDIS_DB")
	if dbs == "" {
		dbs = "0"
	}
	db, err := strconv.Atoi(dbs)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connect to redis (addr:%s, passwd:%s, db:%d)", addr, passwd, db)
	client := NewRedisClient(addr, passwd, db)
	return &Env{S: client}
}
