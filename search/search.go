package search

import (
	"geoip-mmdb/reader"
	"log"
	"net"
)

var (
	cityReader *reader.Reader
	asnReader *reader.Reader
)

type Res struct {
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
}

func Init(asnFile, cityFile string)  {
	if asnFile == "" && cityFile == "" {
		log.Panic("at lease need one mmdb file")
	}
	var err error
	if cityFile != "" {
		cityReader, err = reader.Open(cityFile)
		if err != nil {
			log.Panic(err)
			return
		}
	}
	if asnFile != "" {
		asnReader, err = reader.Open(asnFile)
		if err != nil {
			log.Panic(err)
			return
		}
	}
}

func Search(ip string) (result *Res){
	result = &Res{}
	IP := net.ParseIP(ip)
	if cityReader != nil {
		city, err := cityReader.City(IP)
		if err != nil {
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
	if asnReader != nil {
		asn, err := asnReader.ASN(IP)
		if err != nil {
			return
		}
		result.AutonomousSystemOrganization = asn.AutonomousSystemOrganization
	}
	return
}

func UnInit()  {
	if cityReader != nil {
		cityReader.Close()
	}
	if asnReader != nil {
		asnReader.Close()
	}
}
