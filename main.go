package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"geoip-mmdb/reader"
	"log"
	"net"
	"os"
)

var (
	asnFile string
	cityFile string
	ip string
)

func init() {
	flag.StringVar(&asnFile, "asn", "", "asn mmdb file")
	flag.StringVar(&cityFile, "city", "", "city mmdb file")
	flag.StringVar(&ip, "ip", "", "ip")

	flag.Parse()
	flag.PrintDefaults()
}

func main()  {
	if ip == "" {
		return
	}
	IP := net.ParseIP(ip)
	var result = struct {
		Ip               string                   `json:"ip"`
		ContinentName    string                   `json:"continent"`
		CountryName    string                    `json:"country"`
		Subdivision1Name    string               `json:"province"`
		Name    string                           `json:"city"`
		Subdivision2Name    string               `json:"district"`
		AutonomousSystemOrganization string      `json:"organization"`
		CountryIsoCode    string                 `json:"iso_code"`
		Location struct {
			Longitude      float64                 `json:"longitude"`
			Latitude       float64                 `json:"latitude"`
			AccuracyRadius uint16                  `json:"accuracy_radius"`
		} `json:"location"`
	}{}
	if cityFile != "" {
		f, err := reader.Open(cityFile)
		if err != nil {
			log.Printf("%s %s", cityFile, err.Error())
			return
		}
		defer f.Close()
		city, err := f.City(IP)
		if err != nil {
			log.Printf("%s %s", cityFile, err.Error())
			return
		}
		result.Subdivision2Name = city.Subdivision2Name
		result.Subdivision1Name = city.Subdivision1Name
		result.CountryIsoCode = city.CountryIsoCode
		result.CountryName = city.CountryName
		result.ContinentName = city.ContinentName
		result.Name = city.Name
		result.Location.AccuracyRadius = city.Location.AccuracyRadius
		result.Location.Latitude = city.Location.Latitude
		result.Location.Longitude = city.Location.Longitude
		result.Ip = ip
	}
	if asnFile != "" {
		f, err := reader.Open(asnFile)
		if err != nil {
			log.Printf("%s %s", cityFile, err.Error())
			return
		}
		defer f.Close()
		asn, err := f.ASN(IP)
		if err != nil {
			log.Printf("%s %s", cityFile, err.Error())
			return
		}
		result.AutonomousSystemOrganization = asn.AutonomousSystemOrganization
	}
	bytes, err := json.Marshal(&result)
	if err != nil {
		log.Printf("json %s", err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, string(bytes))
}


