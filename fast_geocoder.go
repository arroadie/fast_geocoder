/*Fast Geocoder
 * Provides a simple interface to geocode ip addresses
 * The application works on 3 ways: entity based via console, file based via console, http server
 * Entity based via console:
 * 		You call the executable passing a single param, the ip address you want to geocode
 *		the response will be on stdout, a json formatted of the latitude, longitude, country and timezone
 * File based via console:
 *		You call the executable passing a single param: the path of a csv file containing on each line a IP address
 *		the response will follow the idea of the single param. For each line, you'll receive the above mentioned formatted response on the stdout
 * Http server:
 *		Calling the executable passing a flag, a server will run and receive calls on the geocode endpoint and return a json response on the same format above mentioned
 */
package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"flag"

	router "github.com/julienschmidt/httprouter"
	db "github.com/oschwald/maxminddb-golang"
)

var isServer bool
var port int
var file db.Reader

type record struct {
	Location struct {
		AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
		Latitude       float64 `maxminddb:"latitude"`
		Longitude      float64 `maxminddb:"longitude"`
		MetroCode      uint    `maxminddb:"metro_code"`
		TimeZone       string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
	Country struct {
		GeoNameID uint              `maxminddb:"geoname_id"`
		IsoCode   string            `maxminddb:"iso_code"`
		Names     map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
}

//Response Is the default response type for marshalling
type Response struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Country  string  `json:"country"`
	Timezone string  `json:"timezone"`
}

func main() {

	initialize()
	loadFile()

	if !isServer {
		argument := os.Args[1]
		var j []byte

		if _, err := os.Stat(argument); err != nil {
			j = geocode(argument)
			fmt.Println(string(j[:]))
		} else {
			// Is a file, consider is a CSV, one ip per line
			ipsfile, err := os.Open(argument)
			if err != nil {
				panic(err)
			}
			r := csv.NewReader(bufio.NewReader(ipsfile))
			for {
				elements, err := r.Read()
				if err != nil {
					break
				}
				j = geocode(elements[0])
				fmt.Println(string(j[:]))
			}

		}
	} else {
		r := router.New()
		r.Handle("GET", "/", defaultHandler)
		r.GET("/geocode/:ip", geocodeHandler)
		fmt.Printf("Running the server on http://localhost:%v\n", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
	}

}

func defaultHandler(w http.ResponseWriter, r *http.Request, ps router.Params) {
	fmt.Fprint(w, "Fast Geocoder")
}

func geocodeHandler(w http.ResponseWriter, r *http.Request, ps router.Params) {
	s := time.Now()
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(geocode(ps.ByName("ip"))))
	log.Printf(" - GET %v - total=%v", r.RequestURI, time.Since(s))
}

func geocode(ips string) []byte {
	var locRecord record
	ip := net.ParseIP(ips)
	_ = file.Lookup(ip, &locRecord)
	res := Response{locRecord.Location.Latitude, locRecord.Location.Longitude, locRecord.Country.IsoCode, locRecord.Location.TimeZone}
	j, _ := json.Marshal(res)
	return j
}

func initialize() {
	flag.BoolVar(&isServer, "server", false, "Starts a server that can be used to geocode ip's on demand")
	flag.IntVar(&port, "port", 8080, "Network port where the server should be running on")

	flag.Parse()
}

func loadFile() {
	tfile, err := db.Open("/tmp/maxmind.mmdb")
	if err != nil {
		if _, err := os.Stat("/tmp/maxmind.mmdb"); err != nil {
			fmt.Println("Database file not present\nDownloading file to \"/tmp/maxmind.mmdb\" (one time operation, unless file is removed)")
			resp, err := http.Get("http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.mmdb.gz")
			defer resp.Body.Close()
			if err != nil {
				panic(err)
			}
			zip, _ := gzip.NewReader(resp.Body)
			out, err := os.Create("/tmp/maxmind.mmdb")
			_, err = io.Copy(out, zip)
			if err != nil {
				fmt.Println("Error trying to save maxmind database to /tmp\nSkipping...")
				err = nil
			}
		}
		tfile, err = db.Open("/tmp/maxmind.mmdb")
		if err != nil {
			panic(err)
		}
	}
	file = *tfile
}
