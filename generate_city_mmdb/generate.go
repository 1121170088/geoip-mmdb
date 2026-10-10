package generate_city_mmdb

import (
	"bufio"
	"encoding/csv"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	cityIpBlocks = make(map[string]*CityIpBlock)
	CityLocations = make(map[string]*CityLocation)
	writer *mmdbwriter.Tree
)

type CityIpBlock struct {
	NetWork string
	GeoNameId string
	RegisteredCountryGeonameId string
	RepresentedCountryGeonameId string
	IsAnonymousProxy string
	IsSatelliteProvider string
	PostalCode string
	Latitude string
	Longitude string
	AccuracyRadius string
	AutonomousSystemOrganization string
}
type CityLocation struct {
	GeoNameId string
	LocaleCode string
	ContinentCode string
	ContinentName string
	CountryIsoCode string
	CountryName string
	Subdivision1IsoCode string
	Subdivision1Name string
	Subdivision2IsoCode string
	Subdivision2Name string
	CityName string
	MetroCode string
	TimeZone string
	IsInEuropeanUnion string
}

func ReadCityIpBlockCsvFile(filename string)  {
	csvFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)
	reader.Read()
	for {
		strs, err := reader.Read()
		if err == io.EOF {
			break
		}
		cityipBlock := CityIpBlock{
			NetWork:                     strs[0],
			GeoNameId:                   strs[1],
			RegisteredCountryGeonameId:  strs[2],
			RepresentedCountryGeonameId: strs[3],
			IsAnonymousProxy:            strs[4],
			IsSatelliteProvider:         strs[5],
			PostalCode:                  strs[6],
			Latitude:                    strs[7],
			Longitude:                   strs[8],
			AccuracyRadius:              strs[9],
		}
		cityIpBlocks[cityipBlock.NetWork] = &cityipBlock
	}
}

func ReadCityLocationCsvFile(filename string)  {
	csvFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)
	reader.Read()
	for {
		strs, err := reader.Read()
		if err == io.EOF {
			break
		}
		citylocation := CityLocation{
			GeoNameId:           strs[0],
			LocaleCode:          strs[1],
			ContinentCode:       strs[2],
			ContinentName:       strs[3],
			CountryIsoCode:      strs[4],
			CountryName:         strs[5],
			Subdivision1IsoCode: strs[6],
			Subdivision1Name:    strs[7],
			Subdivision2IsoCode: strs[8],
			Subdivision2Name:    strs[9],
			CityName:            strs[10],
			MetroCode:           strs[11],
			TimeZone:            strs[12],
			IsInEuropeanUnion:   strs[13],
		}
		CityLocations[citylocation.GeoNameId] = &citylocation
	}
}
func ReadCidrTxt(txtfile string, dst *[]string)  {
	f, err := os.Open(txtfile)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, "\n")
		line = strings.Trim(line, "\r")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		*dst = append(*dst, line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func init()  {

}
func Generatemmdb(pathbase string)  {

	readFiles(pathbase)

	var err error
	writer, err = mmdbwriter.New(
		mmdbwriter.Options{
			DatabaseType: "GeoLite2-City",
			RecordSize:   28,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range cityIpBlocks {
		_, ipnet, err := net.ParseCIDR(k)
		if err != nil {
			log.Panic(err)
		}
		citylocation := getCityLocation(v, k)

		insertData(citylocation, v, ipnet)
	}
	fh, err := os.Create("../GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	_, err = writer.WriteTo(fh)
	if err != nil {
		log.Fatal(err)
	}
}

func readFiles(pathbase string) {
	ReadCityIpBlockCsvFile(filepath.Join(pathbase, "GeoLite2-City-Blocks-IPv4.csv"))
	ReadCityIpBlockCsvFile(filepath.Join(pathbase, "GeoLite2-City-Blocks-IPv6.csv"))
	ReadCityLocationCsvFile(filepath.Join(pathbase, "GeoLite2-City-Locations-zh-CN.csv"))
}

func getCityLocation(v *CityIpBlock, k string) *CityLocation {
	citylocation, ok := CityLocations[v.GeoNameId]
	if !ok {
		log.Printf("city location not found %s", k)
		citylocation = &CityLocation{
			GeoNameId:           "",
			LocaleCode:          "",
			ContinentCode:       "",
			ContinentName:       "",
			CountryIsoCode:      "",
			CountryName:         "",
			Subdivision1IsoCode: "",
			Subdivision1Name:    "",
			Subdivision2IsoCode: "",
			Subdivision2Name:    "",
			CityName:            "",
			MetroCode:           "",
			TimeZone:            "",
			IsInEuropeanUnion:   "",
		}
	}
	return citylocation
}

func insertData(citylocation *CityLocation, v *CityIpBlock, ipnet *net.IPNet) {
	data := mmdbtype.Map{
		"name":             mmdbtype.String(citylocation.CityName),
		"continent_name":   mmdbtype.String(citylocation.ContinentName),
		"country_name":     mmdbtype.String(citylocation.CountryName),
		"country_iso_code": mmdbtype.String(citylocation.CountryIsoCode),
		"location": mmdbtype.Map{
			"accuracy_radius": mmdbtype.Uint16(str2int(v.AccuracyRadius)),
			"latitude":        mmdbtype.Float64(str2float64(v.Latitude)),
			"longitude":       mmdbtype.Float64(str2float64(v.Longitude)),
		},
		"subdivision_1_name":             mmdbtype.String(citylocation.Subdivision1Name),
		"subdivision_2_name":             mmdbtype.String(citylocation.Subdivision2Name),
	}
	err := writer.Insert(ipnet, data)
	if err != nil {
		log.Fatalf("fail to insert to writer %v\n", err)
	}
}

func str2int(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func str2float64(str string) float64 {
	f ,err := strconv.ParseFloat(str,65)
	if err != nil {
		return 0.0
	}
	return f
}

func getaIp(cidr string) string {
	idx := strings.LastIndex(cidr,"/")
	return cidr[:idx]
}
