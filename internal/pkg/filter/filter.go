package filter

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	macClients "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/mac_clients"
)

func Filter(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	// if r.URL.Scheme != "http" { // TODO: inplement ws and https
	// 	http.Error(w, "Scheme not supported", http.StatusBadRequest)
	// 	return errors.New("scheme not supported")
	// }

	client, err := macClients.GetInfoFromIP(r.RemoteAddr)
	if err != nil {
		return err
	}
	log.Println(r.URL.Host)
	inIt, err := data.CheckClientDomain(db, client.MAC.String(), r.URL.Host)
	if err != nil {
		return err
	}
	log.Println("Blocked: ", inIt)
	if inIt {
		http.Error(w, "Blocked By Eyeo", http.StatusForbidden)
		return errors.New("Blocked")
	}

	return nil
}
