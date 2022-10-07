package pureip

import (
	"bufio"
	"encoding/csv"
	"geoip-mmdb/reader"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	Areas = make([]*Area, 0)
)

type Area struct {
	Province string
	City string
	District string
}

func ReadAreaCsv(filename string) {
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
		area := &Area{
			District: strs[0],
			City:     strs[1],
			Province: strs[2],
		}
		Areas = append(Areas, area)
	}
}

func ConvertTxt2Csv(pathbase string)  {
	purecsv, err :=  os.Create("../pure.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer purecsv.Close()
	purecsvWriter := csv.NewWriter(purecsv)

	ReadAreaCsv("../area.csv")
	cityReader, err := reader.Open(filepath.Join(pathbase, "GeoLite2-City.mmdb"))
	if err != nil {
		log.Panic(err)
	}
	defer cityReader.Close()
	f, err := os.Open(filepath.Join(pathbase,"pure.txt"))
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
			city, err := cityReader.City(net.ParseIP(sipstr))
			if err == nil && city.CountryIsoCode == "CN" {


				var province = ""
				var city = ""
				var district = ""
				var maxDistct = 0
				for _, area := range Areas {
					// 省，市 只找desc中连续的2个字，先找省，找到后继续找市，区县取最深度连续的字
					descR := []rune(desc)
					descLen := len(descR)
					prov := []rune(area.Province)
					provLen := len(prov)
					cit := []rune(area.City)
					citLen := len(cit)
					dis := []rune(area.District)
					disLen := len(dis)



					if descLen < 2 || provLen < 2 {
						continue
					}

					used := 0
					if descR[0] == prov[0] && descR[1] == prov[1] {
						province = area.Province
						if descLen < 3 {
							continue
						}
						used = 2
						cityMactch := false
						for i := 0; i < citLen; {
							if used >= descLen {
								break
							}
							if cit[i] == descR[used] {
								i++
								if cityMactch {
									city = area.City
									break
								} else {
									cityMactch = true
								}
							} else {
								if cityMactch {
									break
									cityMactch = false
								}
							}
							used++
						}
						if city != "" {
							disMactched := false
							for i := 0; i < disLen; {
								if used >= descLen {
									break
								}
								if dis[i] == descR[used] {
									i++
									if disMactched {

										if i > maxDistct {
											district = area.District
											maxDistct = i
										}
									} else {
										disMactched = true
									}
								} else {
									if disMactched {
										break
										disMactched = false
									}
								}
								used++
							}
						}
					}

				}

				if province == "" {
					for _, area := range Areas {
						// 市 只找desc中连续的2个字，区县取最深度连续的字
						descR := []rune(desc)
						descLen := len(descR)
						cit := []rune(area.City)
						citLen := len(cit)
						dis := []rune(area.District)
						disLen := len(dis)



						if descLen < 2 || citLen < 2 {
							continue
						}

						used := 0
						if descR[0] == cit[0] && descR[1] == cit[1] {
							province = area.Province
							city = area.City
							if descLen < 3 {
								continue
							}
							used = 2
							disMactched := false
							for i := 0; i < disLen; {
								if used >= descLen {
									break
								}
								if dis[i] == descR[used] {
									i++
									if disMactched {

										if i > maxDistct {
											district = area.District
											maxDistct = i
										}
									} else {
										disMactched = true
									}
								} else {
									if disMactched {
										break
										disMactched = false
									}
								}
								used++
							}

						}

					}
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
			}

		} else {
			log.Println(line)
			continue
		}
	}
	log.Printf("%d" , count)
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
