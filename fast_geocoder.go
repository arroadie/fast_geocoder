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

	"flag"

	router "github.com/julienschmidt/httprouter"
	db "github.com/oschwald/maxminddb-golang"
)

var isServer bool
var port int

var file db.Reader
var record struct {
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

	file := loadFile(&file)
	defer file.Close()

	if !isServer {
		argument := os.Args[1]
		var j []byte

		if _, err := os.Stat(argument); err != nil {
			geocode(argument)
			j = net.ParseIP(argument)
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
		r.GET("/geocode/:ip", geocodeHandler)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
	}

}

func geocodeHandler(w http.ResponseWriter, r *http.Request, ps router.Params) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(geocode(ps.ByName("ip"))))
}

func geocode(ips string) []byte {
	ip := net.ParseIP(ips)
	_ = file.Lookup(ip, &record)
	res := Response{record.Location.Latitude, record.Location.Longitude, record.Country.IsoCode, record.Location.TimeZone}
	j, _ := json.Marshal(res)
	return j
}

func initialize() {
	flag.BoolVar(&isServer, "server", false, "Starts a server that can be used to geocode ip's on demand")
	flag.IntVar(&port, "port", 8080, "Network port where the server should be running on")

	flag.Parse()
}

func loadFile(file *db.Reader) db.Reader {
	file, err := db.Open("/tmp/maxmind.mmdb")
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
		file, err = db.Open("/tmp/maxmind.mmdb")
		if err != nil {
			panic(err)
		}
	}
	return *file
}
