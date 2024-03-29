package merge_pure_city_mmdb

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"geoip-mmdb/reader"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var (
	cityIpBlocks = make(map[string]*CityIpBlock)
	CityLocations = make(map[string]*CityLocation)

	PureAreas = make([][]string, 0)
	MdbAreas = make([]*MdbCN, 0)

	writer *mmdbwriter.Tree

	done = make(chan struct{}, 0)
	pureCityChannel = make(chan []string, 1000)
	mdbCityChannel = make(chan *MdbCN, 1000)
	mergerChannel = make(chan *MdbCN, 1000)
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
	NetWork string
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

type MdbCN struct {
	NetWork string
	City *reader.City
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
func ReadPureCsvFile(filename string)  {
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

		PureAreas = append(PureAreas, strs)
	}
}

func Generatemmdb2(pathbase string)  {

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

		citylocation := getCityLocation(v, k)

		city := &reader.City{
			Name:           citylocation.CityName,
			ContinentName:  citylocation.ContinentName,
			CountryName:    citylocation.CountryName,
			CountryIsoCode: citylocation.CountryIsoCode,
			Location: struct {
				AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
				Latitude       float64 `maxminddb:"latitude"`
				Longitude      float64 `maxminddb:"longitude"`
			}{
				AccuracyRadius: uint16(str2int(v.AccuracyRadius)),
				Latitude: str2float64(v.Latitude),
				Longitude: str2float64(v.Longitude),
			},
			Subdivision1Name: citylocation.Subdivision1Name,
			Subdivision2Name: citylocation.Subdivision2Name,
		}
		_, ipnet, err := net.ParseCIDR(k)
		if err != nil {
			log.Panic(err)
		}
		insertData(ipnet, city)
	}

	for _, pureCity := range PureAreas {
		if pureCity[4] != "" {
			startInt,err := ip2Uint32(pureCity[0])
			if err != nil {
				log.Panic(err)
			}
			endInt, err := ip2Uint32(pureCity[1])
			if err != nil {
				log.Panic(err)
			}
			cidrstr := getCidrStr(startInt, endInt)
			_, ipnet, err := net.ParseCIDR(cidrstr)
			if err != nil {
				log.Panic(err)
			}
			city := reader.City{
				Name:           pureCity[3],
				ContinentName:  "亚洲",
				CountryName:    "中国",
				CountryIsoCode: "CN",
				Location: struct {
					AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
					Latitude       float64 `maxminddb:"latitude"`
					Longitude      float64 `maxminddb:"longitude"`
				}{},
				Subdivision1Name: pureCity[2],
				Subdivision2Name: pureCity[4],
			}
			insertData(ipnet, &city)
		}

	}

	fh, err := os.Create("../GeoLite2-merge-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	_, err = writer.WriteTo(fh)
	if err != nil {
		log.Fatal(err)
	}
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

		citylocation := getCityLocation(v, k)

		city := &reader.City{
			Name:           citylocation.CityName,
			ContinentName:  citylocation.ContinentName,
			CountryName:    citylocation.CountryName,
			CountryIsoCode: citylocation.CountryIsoCode,
			Location: struct {
				AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
				Latitude       float64 `maxminddb:"latitude"`
				Longitude      float64 `maxminddb:"longitude"`
			}{
				AccuracyRadius: uint16(str2int(v.AccuracyRadius)),
				Latitude: str2float64(v.Latitude),
				Longitude: str2float64(v.Longitude),
			},
			Subdivision1Name: citylocation.Subdivision1Name,
			Subdivision2Name: citylocation.Subdivision2Name,
		}
		if city.CountryIsoCode != "CN" || strings.Index(k, ":" ) != -1 {
		//if strings.Index(k, ":" ) != -1 {
			// 不是cn的ip范围 或者 ipv6范围 先插入
			_, ipnet, err := net.ParseCIDR(k)
			if err != nil {
				log.Panic(err)
			}
			insertData(ipnet, city)
		} else {
			mdbCN := MdbCN{
				NetWork: k,
				City:    city,
			}
			MdbAreas = append(MdbAreas, &mdbCN)
		}
	}


	SortAscMdbAreaByIp()
	log.Println("sorted mdb cities")
	SortAscPureArea()
	log.Println("sorted pure cities")

	go SendMdbCity()
	go SendPureCity()
	go MergeCity()
	go HandMergeChannel()

	<- done

	fh, err := os.Create("../GeoLite2-merge-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()
	_, err = writer.WriteTo(fh)
	if err != nil {
		log.Fatal(err)
	}
}

func SortAscMdbAreaByIp()  {
	sort.Slice(MdbAreas, func(i, j int) bool {
		later := MdbAreas[i]
		former := MdbAreas[j]

		//ipCidr2Uint32(cidr string) (min uint32, max uint32, hostNum uint32, err error)
		laterIpMin, _, _, err := ipCidr2Uint32(later.NetWork)
		if err != nil {
			log.Panic(err)
		}
		formerIpMin, _, _, err := ipCidr2Uint32(former.NetWork)
		if err != nil {
			log.Panic(err)
		}
		return laterIpMin < formerIpMin
	})
}

func SortAscPureArea()  {
	sort.Slice(PureAreas, func(i, j int) bool {
		later := PureAreas[i]
		former := PureAreas[j]
		laterIp, err := ip2Uint32(later[0])
		if err != nil {
			log.Panic(err)
		}
		formerIp, err := ip2Uint32(former[0])
		if err != nil {
			log.Panic(err)
		}
		return laterIp < formerIp
	})
}

func SendMdbCity()  {
	for _, city:= range MdbAreas {
		mmin, mmax, _,  err := ipCidr2Uint32(city.NetWork)
		if err != nil {
			log.Panic(err)
		}
		for i:=mmin; i <= mmax; i++ {
			mdbCN := MdbCN{
				NetWork: int2ip(i).String(),
				City:    city.City,
			}
			if city.City.Location.Longitude == 0 {
				log.Printf("dddddddd")
			}
			mdbCityChannel <- &mdbCN
		}
	}
	close(mdbCityChannel)
	log.Printf("send mdc city finished")
}
func SendPureCity()  {
	for _, city := range PureAreas {
		pIpStartStr := city[0]
		pIpEndStr := city[1]

		pMin, err := ip2Uint32(pIpStartStr)
		if err != nil {
			log.Panic(err)
		}

		pMax, err := ip2Uint32(pIpEndStr)
		if err != nil {
			log.Panic(err)
		}
		for i:=pMin; i <= pMax; i++ {

			pureCity := []string{
				int2ip(i).String(),
				city[2],
				city[3],
				city[4],
			}
			pureCityChannel <- pureCity
		}
	}
	close(pureCityChannel)
	log.Printf("send pure city finished")
}
func MergeCity()  {
	var pureCity []string
	var pureChannelDone = false
	var pureIpint uint32 = 0
	var getPureCity = true

	var mdbCity *MdbCN
	var mdbCityChannelDone = false
	var mdbIpInt uint32 = 0
	var getMdbCity = true
	for {
		if getPureCity {
			pureCity = <- pureCityChannel
		}
		if pureCity == nil {
			pureChannelDone = true
		}
		if getMdbCity {
			mdbCity = <- mdbCityChannel
		}
		if mdbCity == nil {
			mdbCityChannelDone = true
		}

		if pureChannelDone && mdbCityChannelDone {
			break
		}

		if !pureChannelDone && !mdbCityChannelDone {
			pureIpStr := pureCity[0]
			pureProvince := pureCity[1]
			pureCit := pureCity[2]
			pureDistrict := pureCity[3]

			var err error
			if getPureCity {
				pureIpint, err = ip2Uint32(pureIpStr)

				if err != nil {
					log.Panic(err)
				}
			}
			if getMdbCity {
				mdbIpInt, err = ip2Uint32(mdbCity.NetWork)
				if err != nil {
					log.Panic(err)
				}
			}
			if pureIpint > mdbIpInt {
				getPureCity = false
				getMdbCity = true
				mergerChannel <- mdbCity
			} else if pureIpint == mdbIpInt {

				getMdbCity = true
				getPureCity = true
				newCity := &reader.City{
					Name:           mdbCity.City.Name,
					ContinentName:  mdbCity.City.ContinentName,
					CountryName:    mdbCity.City.CountryName,
					CountryIsoCode: mdbCity.City.CountryIsoCode,
					Location: struct {
						AccuracyRadius uint16  `maxminddb:"accuracy_radius"`
						Latitude       float64 `maxminddb:"latitude"`
						Longitude      float64 `maxminddb:"longitude"`
					}{
						AccuracyRadius: mdbCity.City.Location.AccuracyRadius,
						Latitude:       mdbCity.City.Location.Latitude,
						Longitude:      mdbCity.City.Location.Longitude,
					},
					Subdivision1Name: mdbCity.City.Subdivision1Name,
					Subdivision2Name: mdbCity.City.Subdivision2Name,
				}
				city := &MdbCN{
					NetWork: mdbCity.NetWork,
					City: newCity,
				}
				//if  pureDistrict != "" || newCity.Name == "" || newCity.Subdivision1Name == "" {
				//	newCity.Subdivision2Name = pureDistrict
				//	newCity.Name = pureCit
				//	newCity.Subdivision1Name = pureProvince
				//}
				//if  pureProvince != "" {
				//	newCity.Subdivision2Name = pureDistrict
				//	newCity.Name = pureCit
				//	newCity.Subdivision1Name = pureProvince
				//}
				if pureDistrict != "" {
					newCity.Subdivision2Name = pureDistrict
				}
				if pureCit != "" {
					newCity.Name = pureCit
				}
				if pureProvince != "" {
					newCity.Subdivision1Name = pureProvince
				}

				mergerChannel <- city
			} else {
				// mdbIpInt > pureIpint
				getMdbCity = false
				getPureCity = true
			}
		} else if mdbCityChannelDone && !pureChannelDone {
			getPureCity = true
			<- pureCityChannel
		} else if !mdbCityChannelDone && pureChannelDone {
			getMdbCity = true
			mergerChannel <- mdbCity
		}
	}
	close(mergerChannel)
	log.Printf("merge city finished")
}
func HandMergeChannel()  {
	var theCity *reader.City
	var startIp uint32
	var preIp uint32
	var lastCity *reader.City
	var lastIp uint32
	var pre4thByte uint8


	for city:= range mergerChannel {
		if city == nil {
			break
		}
		lastCity = city.City
		intIp, err := ip2Uint32(city.NetWork)
		if err != nil {
			log.Panic(err)
		}
		lastIp = intIp
		current4thByte := uint8(lastIp)

		if theCity == nil {
			intIp, err := ip2Uint32(city.NetWork)
			if err != nil {
				log.Panic(err)
			}
			startIp = intIp
			preIp = intIp
			theCity = city.City
			pre4thByte = current4thByte
		} else {
			if (theCity.Subdivision1Name != city.City.Subdivision1Name &&
				theCity.Name != city.City.Name &&
				theCity.Subdivision2Name != city.City.Subdivision2Name) ||
				preIp + 1 != lastIp || current4thByte <= pre4thByte {

				cidr := getCidrStr(startIp, preIp)
				_, ipnet,  err := net.ParseCIDR(cidr)
				if err != nil {
					log.Panic(err)
				}
				insertData(ipnet, theCity)
				log.Printf("%s,%s,%s,%s", ipnet.String(), theCity.Subdivision1Name, theCity.Name, theCity.Subdivision2Name)
				theCity = lastCity
				preIp = lastIp
				startIp = lastIp
				pre4thByte = current4thByte
			} else {
				preIp = lastIp
				pre4thByte = current4thByte
			}
		}
	}
	if theCity != nil {
		cidr := getCidrStr(startIp, preIp)
		_, ipnet,  err := net.ParseCIDR(cidr)
		if err != nil {
			log.Panic(err)
		}
		insertData(ipnet, theCity)
	}
	done<- struct{}{}
	log.Printf("HandMergeChannel finished")
}
func readFiles(pathbase string) {
	ReadCityIpBlockCsvFile(filepath.Join(pathbase, "GeoLite2-City-Blocks-IPv4.csv"))
	ReadCityIpBlockCsvFile(filepath.Join(pathbase, "GeoLite2-City-Blocks-IPv6.csv"))
	ReadCityLocationCsvFile(filepath.Join(pathbase, "GeoLite2-City-Locations-zh-CN.csv"))
	ReadPureCsvFile(filepath.Join(pathbase, "pure.csv"))
}

func getCityLocation(v *CityIpBlock, k string) *CityLocation {
	citylocation, ok := CityLocations[v.GeoNameId]
	if !ok {
		log.Printf("city location not found %s", k)
		citylocation = &CityLocation{
			NetWork: k,
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

	citylocation.NetWork = k
	return citylocation
}


func insertData(ipnet *net.IPNet, city *reader.City) {
	data := mmdbtype.Map{
		"name":             mmdbtype.String(city.Name),
		"continent_name":   mmdbtype.String(city.ContinentName),
		"country_name":     mmdbtype.String(city.CountryName),
		"country_iso_code": mmdbtype.String(city.CountryIsoCode),
		"location": mmdbtype.Map{
			"accuracy_radius": mmdbtype.Uint16(city.Location.AccuracyRadius),
			"latitude":        mmdbtype.Float64(city.Location.Latitude),
			"longitude":       mmdbtype.Float64(city.Location.Longitude),
		},
		"subdivision_1_name":             mmdbtype.String(city.Subdivision1Name),
		"subdivision_2_name":             mmdbtype.String(city.Subdivision2Name),
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

func getCidrStr(start, end uint32) string  {
	ip1 := int2ip(start)
	ip2 := int2ip(end)
	maxLen := 32
	for l := maxLen; l >= 0; l-- {
		mask := net.CIDRMask(l, maxLen)
		na := ip1.Mask(mask)
		n := net.IPNet{IP: na, Mask: mask}

		if n.Contains(ip2) {
			return fmt.Sprintf("%v/%v", na, l)
		}
	}
	return ""
}

func ipCidr2Uint32(cidr string) (min uint32, max uint32, hostNum uint32, err error) {
	strs := strings.Split(cidr, "/")
	ipstr := strs[0]
	maskstr := strs[1]
	bs:= strings.Split(ipstr, ".")
	var b1 int
	var b2 int
	var b3 int
	var b4 int
	var mask int
	mask, err= strconv.Atoi(maskstr)
	if err != nil {
		return
	}
	b1, err= strconv.Atoi(bs[0])
	if err != nil {
		return
	}
	b2, err= strconv.Atoi(bs[1])
	if err != nil {
		return
	}
	b3, err= strconv.Atoi(bs[2])
	if err != nil {
		return
	}
	b4, err= strconv.Atoi(bs[3])
	if err != nil {
		return
	}
	min = uint32(b1)<<24 |
		uint32(b2) << 16 |
		uint32(b3) << 8 |
		uint32(b4) & (0xffffffff << (32 - mask))
	hostNum = 0xffffffff >> mask
	max = min + hostNum
	return
}

func ip2Uint32(ipstr string) (ipint uint32, err error) {

	bs:= strings.Split(ipstr, ".")
	var b1 int
	var b2 int
	var b3 int
	var b4 int
	b1, err= strconv.Atoi(bs[0])
	if err != nil {
		return
	}
	b2, err= strconv.Atoi(bs[1])
	if err != nil {
		return
	}
	b3, err= strconv.Atoi(bs[2])
	if err != nil {
		return
	}
	b4, err= strconv.Atoi(bs[3])
	if err != nil {
		return
	}
	ipint = uint32(b1)<<24 |
		uint32(b2) << 16 |
		uint32(b3) << 8 |
		uint32(b4)
	return
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}