package generate_test

import (
	"geoip-mmdb/generate_asn_mmdb"
	"geoip-mmdb/generate_city_mmdb"
	"geoip-mmdb/reader"
	"github.com/stretchr/testify/require"
	"log"
	"net"
	"testing"
)

func Test_Generate_Asn(t *testing.T) {
	generate_asn_mmdb.Generatemmdb("../generate_asn_mmdb")
}

func Test_Generate_City(t *testing.T) {
	generate_city_mmdb.Generatemmdb("../generate_city_mmdb")
}

func TestReader(t *testing.T) {
	cityReader, err := reader.Open("../GeoLite2-City.mmdb")
	require.NoError(t, err)
	defer cityReader.Close()
	asnReader, err := reader.Open("../GeoLite2-ASN.mmdb")
	require.NoError(t, err)
	defer asnReader.Close()


	cidr := "180.149.130.16"
	cityRecord, err := cityReader.City(net.ParseIP(cidr))
	require.NoError(t, err)

	asnRecord, err := asnReader.ASN(net.ParseIP(cidr))
	require.NoError(t, err)

	log.Printf("%v %v", cityRecord, asnRecord)
}
