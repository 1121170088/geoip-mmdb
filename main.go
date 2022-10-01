package main

import (
	"encoding/json"
	"flag"
	"geoip-mmdb/reader"
	"log"
	"net"
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
}

func main()  {
	if ip == "" {
		return
	}
	IP := net.ParseIP(ip)
	var result = struct {
		AutonomousSystemOrganization string
		Name    string
		ContinentName    string
		CountryName    string
		CountryIsoCode    string
		Subdivision1Name    string
		Subdivision2Name    string
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
	println(string(bytes))
}


