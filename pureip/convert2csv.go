package pureip

import (
	"bufio"
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//var (
//	Areas = make([]*Area, 0)
//)

type Area struct {
	Province string
	City     string
	District string
}

//func ReadAreaCsv(filename string) {
//	csvFile, err := os.Open(filename)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer csvFile.Close()
//	reader := csv.NewReader(csvFile)
//	reader.Read()
//	for {
//		strs, err := reader.Read()
//		if err == io.EOF {
//			break
//		}
//		area := &Area{
//			District: strs[0],
//			City:     strs[1],
//			Province: strs[2],
//		}
//		Areas = append(Areas, area)
//	}
//}

func ConvertTxt2Csv(pathbase string) {
	purecsv, err := os.Create("../pure.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer purecsv.Close()
	purecsvWriter := csv.NewWriter(purecsv)

	//ReadAreaCsv("../area.csv")
	//cityReader, err := reader.Open(filepath.Join(pathbase, "GeoLite2-City.mmdb"))
	//if err != nil {
	//	log.Panic(err)
	//}
	//defer cityReader.Close()
	f, err := os.Open(filepath.Join(pathbase, "pure/pure.txt"))
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var rexp = regexp.MustCompile(`^(\d+\.\d+\.\d+\.\d+)\s+(\d+\.\d+\.\d+\.\d+)\s+(.+)$`)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, "\n")
		line = strings.Trim(line, "\r")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		strs := rexp.FindAllStringSubmatch(line, -1)

		if len(strs) == 1 && len(strs[0]) == 4 {
			sipstr := strs[0][1]
			eipstr := strs[0][2]
			desc := strs[0][3]
			desc = strings.TrimSpace(desc)
			//city, err := cityReader.City(net.ParseIP(sipstr))
			//if err == nil && city.CountryIsoCode == "CN" {
			if !strings.Contains(desc, "中国") {
				continue
			}
			var province = ""
			var city = ""
			var district = ""
			ss := strings.Split(desc, " ")
			ss = strings.Split(ss[0], "–")
			slen := len(ss)

			//province, city, district, desc
			if slen == 2 {
				province = ss[1]
			} else if slen == 3 {
				province = ss[1]
				city = ss[2]
			} else if slen > 3 {
				province = ss[1]
				city = ss[2]
				district = ss[3]
			}

			if province != "" {
				//log.Printf("sip %s eip %s desc %s ||| %s %s %s", sipstr, eipstr, desc, province, city, district)
				record := []string{sipstr, eipstr, province, city, district, desc}
				if err := purecsvWriter.Write(record); err != nil {
					log.Fatalln("error writing record to file", err)
				}
				purecsvWriter.Flush()
				if purecsvWriter.Error() != nil {
					log.Panic(purecsvWriter.Error())
				}
				count++
			} else {
				log.Printf("sip %s eip %s desc %s ||| %s %s %s", sipstr, eipstr, desc, province, city, district)
			}
			//}

		} else {
			log.Println(line)
			continue
		}
	}
	log.Printf("%d", count)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
