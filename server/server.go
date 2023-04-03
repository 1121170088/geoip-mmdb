package server

import (
	"context"
	"encoding/json"
	"geoip-mmdb/search"
	ds "github.com/1121170088/find-domain/search"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func Start(addr, tldFile string)  {
	if tldFile != "" {
		ds.Init(tldFile)
	}

	http.HandleFunc("/domains", func(writer http.ResponseWriter, request *http.Request) {
		header := writer.Header()
		header.Add("Content-Type", "application/json;charset=UTF-8")
		level := request.URL.Query().Get("level")

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
		if tldFile != "" && level == "2" {
			tempMap := make(map[string] struct{})
			for _, domain := range domains {
				dm := ds.Search(domain)
				if dm != "" {
					tempMap[dm] = struct{}{}
				}
			}
			domains = make([]string, 0)
			for k, _ := range tempMap {
				domains = append(domains, k)
			}
		}
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
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		header := writer.Header()
		header.Add("Content-Type", "application/json;charset=UTF-8")
		ipStr := ""
		uri := request.RequestURI
		uri = strings.TrimSpace(uri)
		uri = uri[1:]

		if uri != "" {
			ipStr = uri
		} else {
			remoteAddr := request.RemoteAddr
			fwdAddress := request.Header.Get("X-Forwarded-For")
			if fwdAddress != "" {
				ipAddress := fwdAddress
				ips := strings.Split(fwdAddress, ",")
				if len(ips) > 1 {
					ipAddress = ips[0]
				}
				remoteAddr = ipAddress
			}
			strs := strings.Split(remoteAddr, ":")
			if len(strs) > 1 {
				// should be ipv6
				ipStr = remoteAddr
			} else {
				// should be ip:port
				ipStr = strs[0]
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