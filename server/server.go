package server

import (
	"context"
	"encoding/json"
	"geoip-mmdb/search"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func Start(addr string)  {
	http.HandleFunc("/domains", func(writer http.ResponseWriter, request *http.Request) {
		header := writer.Header()
		header.Add("Content-Type", "application/json;charset=UTF-8")

		var domains []string
		bytes, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(400)
			return
		}
		err = json.Unmarshal(bytes, &domains)
		if err != nil {
			log.Printf("%v", err)
			log.Printf("%s", string(bytes))
			writer.WriteHeader(400)
			return
		}
		result := make(map[string] *search.Res)
		for _, domain := range domains {
			res := &search.Res{}
			ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", domain)
			if err != nil || len(ips) == 0 {
				res = &search.Res{
					Ip:                           "",
					ContinentName:                "",
					CountryName:                  "",
					Subdivision1Name:             "",
					Name:                         "",
					Subdivision2Name:             "",
					AutonomousSystemOrganization: "",
					CountryIsoCode:               "",
					Location: struct {
						Longitude      float64 `json:"longitude"`
						Latitude       float64 `json:"latitude"`
						AccuracyRadius uint16  `json:"accuracy_radius"`
					}{},
				}
			} else {
				res = search.Search(ips[0].String())
			}
			result[domain] = res
		}
		bytes, err = json.Marshal(result)
		if err != nil {
			writer.WriteHeader(400)
			return
		}
		writer.Write(bytes)

	})
	http.HandleFunc("/ip", func(writer http.ResponseWriter, request *http.Request) {
		header := writer.Header()
		header.Add("Content-Type", "application/json;charset=UTF-8")
		ipStr := ""
		rawQuery := request.URL.RawQuery
		rawQuery = strings.TrimSpace(rawQuery)
		if rawQuery != "" {
			ipStr = rawQuery
		} else {
			remoteAddr := request.RemoteAddr
			fwdAddress := request.Header.Get("X-Forwarded-For")
			if fwdAddress != "" {
				ipAddress := fwdAddress
				ips := strings.Split(fwdAddress, ", ")
				if len(ips) > 1 {
					ipAddress = ips[0]
				}
				remoteAddr = ipAddress
			}
			idx := strings.Index(remoteAddr, ":")
			if idx != -1 {
				ipStr = remoteAddr[0: idx]
			} else {
				ipStr = remoteAddr
			}

		}
		result := search.Search(ipStr)
		bytes, err := json.Marshal(result)
		if err != nil {
			writer.WriteHeader(400)
			return
		}
		writer.Write(bytes)

	})

	http.ListenAndServe(addr, nil)
}