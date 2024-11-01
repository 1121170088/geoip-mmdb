package merge_ip2region

import (
	"encoding/binary"
	"encoding/csv"
	"github.com/orestonce/Ip2regionTool"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func ConvertDbToTxt(req Ip2regionTool.ConvertDbToTxt_Req) (errMsg string) {
	if req.DbVersion == 0 {
		req.DbVersion = Ip2regionTool.GetDbVersionByName(req.DbFileName)
	}
	stat, err := os.Stat(req.DbFileName)
	if err != nil {
		return "文件状态错误: " + req.DbFileName + "," + err.Error()
	}
	if stat.Size() > 1000*1024*1024 {
		return "不支持超过1000MB的db文件: " + strconv.Itoa(int(stat.Size()))
	}
	dbFileContent, err := ioutil.ReadFile(req.DbFileName)
	if err != nil {
		return "读取db文件失败: " + req.DbFileName + ", " + err.Error()
	}
	var list []Ip2regionTool.IpRangeItem
	if req.DbVersion == 1 {
		list, errMsg = Ip2regionTool.ReadV1DataBlob(dbFileContent)
	} else {
		list, errMsg = Ip2regionTool.ReadV2DataBlob(dbFileContent)
	}
	if errMsg != `` {
		return "文件数据错误: " + errMsg
	}
	if req.Merge {
		list = Ip2regionTool.MergeIpRangeList(list)
	}
	errMsg = Ip2regionTool.VerifyIpRangeList(Ip2regionTool.VerifyIpRangeListRequest{
		DataInfoList:     list,
		VerifyFullUint32: true,
		VerifyFiled7:     req.DbVersion == 1, // 只有版本1才需要验证字段数为7
	})
	if errMsg != `` {
		return "验证文件数据失败: " + errMsg
	}

	csf, err := os.Create("../ip2region.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csf.Close()
	csvWriter := csv.NewWriter(csf)
	defer csvWriter.Flush()
	WriteCsv(list, csvWriter)
	return ""
}

type IpRangeItem struct {
	Origin  string
	LowU32  uint32
	HighU32 uint32
	Attach  string
	CityId  uint32
}

func WriteCsv(list []Ip2regionTool.IpRangeItem, csrwriter *csv.Writer) {
	for _, one := range list {
		ss := strings.Split(one.Attach, "|")
		//国家|区域|省份|城市|ISP
		if ss[0] != "中国" {
			continue
		}
		if ss[1] == "0" {
			ss[1] = ""
		}
		if ss[2] == "0" {
			ss[2] = ""
		}
		if ss[3] == "0" {
			ss[3] = ""
		}
		if ss[3] == "" {
			continue
		}
		pr := []rune(ss[2])
		prl := len(pr)
		if prl > 0 {
			if string(pr[prl-1]) == "省" {
				ss[2] = string(pr[0 : prl-1])
			}
			if string(pr[prl-1]) == "市" {
				ss[2] = string(pr[0 : prl-1])
			}
			if prl > 3 {
				if string(pr[prl-3:]) == "自治区" {
					ss[2] = string(pr[0 : prl-3])
				}
			}
		}
		record := []string{uint32ToIp(one.LowU32).String(), uint32ToIp(one.HighU32).String(),
			ss[2], ss[3], ss[1], ""}
		csrwriter.Write(record)
	}
}

func uint32ToIp(ip uint32) net.IP {
	var tmp = make([]byte, 4)
	binary.BigEndian.PutUint32(tmp, ip)
	return net.IPv4(tmp[0], tmp[1], tmp[2], tmp[3])
}
