package merge_ip2region

import (
	"fmt"
	"github.com/orestonce/Ip2regionTool"
	"os"
	"testing"
)

func TestConvertDbToTxt(t *testing.T) {
	var dbPath = "ip2region.xdb"
	var txtFileName = "1.txt"
	errMsg := ConvertDbToTxt(Ip2regionTool.ConvertDbToTxt_Req{
		DbFileName:  dbPath,
		TxtFileName: txtFileName,
		Merge:       true,
		DbVersion:   2,
	})
	if errMsg != `` {
		fmt.Println(errMsg)
		os.Exit(-1)
	}
}
