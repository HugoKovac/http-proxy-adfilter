package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/db"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/api"
	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/proxy"
)


func main(){
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)


	boltdb := db.NewDatabase()
	stopHTTP, stopHTTPS, errorChan := proxy.ListenProxy(boltdb)
	go api.ListenHandler(boltdb)


	go data.GetCategorizedDomainList(boltdb, []string{"base", "mobile-monetization", "oisd_nsfw", "override"}) // , "tif"})

	select {
	case <- sigs:
		boltdb.Close()
		close(stopHTTP)
		close(stopHTTPS)
	case err := <- errorChan:
		log.Println("Main Thread Intercepted: ", err)
		boltdb.Close()
		close(stopHTTP)
		close(stopHTTPS)
	}
}
