package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
	// "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/proxy"
)


func main(){
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)


	db, boltdb := db.NewDatabase()
	// go proxy.ListenProxy(db, boltdb)
	
	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	data.GetCategorizedDomainList(db, boltdb)

	<-sigs
	db.Close()	
}

/*
	curl -v --interface en0 -x http://localhost:8080 http://www.google.com
	curl -v --interface en0 -X POST http://localhost:8080/add_sub_list --data category=gambling
	curl -v --interface en0 -x http://localhost:8080 http://stake.com
*/
