package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"geoip-mmdb/search"
	"geoip-mmdb/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	asnFile string
	cityFile string
	ip string
	serverMode bool
	addr string
	tldFile string

)

func init() {
	flag.StringVar(&asnFile, "asn", "./GeoLite2-ASN.mmdb", "asn mmdb file default ./GeoLite2-ASN.mmdb")
	flag.StringVar(&cityFile, "city", "./GeoLite2-City.mmdb", "city mmdb file default ./GeoLite2-City.mmdb")
	flag.StringVar(&ip, "ip", "", "ip")
	flag.BoolVar(&serverMode, "s", false, "http server mode")
	flag.StringVar(&addr, "addr", "127.0.0.1:9080", "server addr, default 127.0.0.1:9080")
	flag.StringVar(&tldFile, "tld", "", "default empty string, may download at https://publicsuffix.org/list/public_suffix_list.dat, using for getting sub domain from domain submitted via http request")

	flag.Parse()
	serverMode = true

}

func main()  {

	search.Init(asnFile, cityFile)


	if !serverMode {
		result := search.Search(ip)
		bytes, err := json.Marshal(result)
		if err != nil {
			log.Printf("json %s", err.Error())
			return
		}
		fmt.Fprintln(os.Stdout, string(bytes))

	} else {

		 go server.Start(addr, tldFile)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
	}

	search.UnInit()
}


