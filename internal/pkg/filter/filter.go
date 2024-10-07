package filter

import (
	"database/sql"
	"errors"
	"net/http"

	"gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/data"
	macClients "gitlab.com/eyeo/network-filtering/router-adfilter-go/internal/pkg/mac_clients"
	"github.com/boltdb/bolt"
)

func Filter(w http.ResponseWriter, r *http.Request, db *sql.DB, boltdb *bolt.DB) error {
	// if r.URL.Scheme != "http" { // TODO: inplement ws and https
	// 	http.Error(w, "Scheme not supported", http.StatusBadRequest)
	// 	return errors.New("scheme not supported")
	// }

	client, err := macClients.GetInfoFromIP(r.RemoteAddr)
	if err != nil {
		return err
	}
	ok, err := data.CheckClientDomain(db, boltdb, client.MAC.String(), r.URL.Host)
	if err != nil {
		return err
	}
	if ok {
		http.Error(w, "Blocked By Eyeo", http.StatusForbidden)
		return errors.New("Blocked")
	}

	return nil
}
