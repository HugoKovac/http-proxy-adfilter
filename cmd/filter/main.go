package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/proxy"
)

func main(){
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)


	db := db.NewDatabase()
	go proxy.ListenProxy(db)
	
	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	data.GetCategorizedDomainList(db)

	<-sigs
	db.Close()	
}