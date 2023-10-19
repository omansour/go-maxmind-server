package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/oschwald/geoip2-golang"
)

func lookupIP(db *geoip2.Reader, ip net.IP) (*geoip2.City, error) {
	return db.City(ip)
}

func main() {
	db, err := geoip2.Open("GeoLite2-City.mmdb") // Path to the MaxMind database
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/lookup/{ip:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ipStr := vars["ip"]

		var ip net.IP

		if ipStr == "" {
			// Get IP address from the incoming request
			ipStr = r.RemoteAddr
			host, _, err := net.SplitHostPort(ipStr) // Splitting to get the IP from "IP:port" format
			if err != nil {
				http.Error(w, "Error extracting IP from request", http.StatusInternalServerError)
				return
			}
			ip = net.ParseIP(host)
		} else {
			ip = net.ParseIP(ipStr)
		}

		if ip == nil {
			http.Error(w, "Invalid IP address", http.StatusBadRequest)
			return
		}

		city, err := lookupIP(db, ip)
		if err != nil {
			http.Error(w, "Error looking up IP", http.StatusInternalServerError)
			return
		}

		/*respData := map[string]interface{}{
			"city":    city.City.Names["en"],
			"country": city.Country.Names["en"],
			"isoCode": city.Country.IsoCode,
		}*/

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(city)
	})

	http.ListenAndServe(":8080", router)
}
