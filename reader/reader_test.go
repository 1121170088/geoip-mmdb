package reader

import (
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReader(t *testing.T) {
	reader, err := Open("../GeoLite2-City.mmdb")
	require.NoError(t, err)

	defer reader.Close()

	record, err := reader.City(net.ParseIP("2409:8a20:857:e0b1:14ed::70f"))
	require.NoError(t, err)

	m := reader.Metadata()
	assert.Equal(t, uint(2), m.BinaryFormatMajorVersion)
	assert.Equal(t, uint(0), m.BinaryFormatMinorVersion)
	assert.NotZero(t, m.BuildEpoch)
	assert.Equal(t, "GeoLite2-City", m.DatabaseType)

	log.Printf("%v", record)
}
