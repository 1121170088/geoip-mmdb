package util

import (
	"math/big"
	"net"
	"strconv"
	"strings"
)

func IpCidr2Uint32(cidr string) (min *big.Int, max *big.Int, hostNum *big.Int, err error) {
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
	max = big.NewInt(0).Add(min, hostNum)
	return
}

var maxIpv6 = big.NewInt(0).Sub(big.NewInt(0).Exp(big.NewInt(2), big.NewInt(128), nil), big.NewInt(1))

func IpCidr62Int(cidr6 string) (min *big.Int, max *big.Int, hostNum *big.Int, err error) {
	_, ipnet, err := net.ParseCIDR(cidr6)
	if err != nil {
		return
	}
	bigint := big.NewInt(0)
	bigint.SetBytes(ipnet.IP)
	mask := big.NewInt(0)
	mask.SetBytes(ipnet.Mask)
	min = big.NewInt(0).And(bigint, mask)
	ones, _ := ipnet.Mask.Size()
	hostNum = big.NewInt(0).Rsh(maxIpv6, uint(ones))
	max = min.Add(min, hostNum)
	return
}