package server

import (
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
			result[domain] = nil
			ips, err := net.LookupIP(domain)
			if err != nil || len(ips) == 0 {
				continue
			}
			res := search.Search(ips[0].String())
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
			idx := strings.Index(remoteAddr, ":")
			ipStr = remoteAddr[0: idx]
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