package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	db "github.com/oschwald/maxminddb-golang"
)

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

type Response struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Country  string  `json:"country"`
	Timezone string  `json:"timezone"`
}

func main() {

	file := loadFile(&file)
	defer file.Close()

	argument := os.Args[1]

	if _, err := os.Stat(argument); err != nil {
		ip := net.ParseIP(argument)
		_ = file.Lookup(ip, &record)
		res := Response{record.Location.Latitude, record.Location.Longitude, record.Country.IsoCode, record.Location.TimeZone}
		j, _ := json.Marshal(res)
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
			ip := net.ParseIP(elements[0])
			_ = file.Lookup(ip, &record)
			res := Response{record.Location.Latitude, record.Location.Longitude, record.Country.IsoCode, record.Location.TimeZone}
			j, _ := json.Marshal(res)
			fmt.Println(string(j[:]))
		}

	}

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
