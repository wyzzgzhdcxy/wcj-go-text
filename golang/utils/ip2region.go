package utils

import (
	"fmt"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

func IpParse(ip string) string {
	var dbPath = core.GetTempDir() + "/app_data/ip2region.xdb"
	searcher, err := xdb.NewWithFileOnly(xdb.IPvx, dbPath)
	if err != nil {
		fmt.Printf("failed to create searcher: %s\n", err.Error())
		return ""
	}

	defer searcher.Close()

	// do the search
	region, err := searcher.SearchByStr(ip)
	if err != nil {
		fmt.Printf("failed to SearchIP(%s): %s\n", ip, err)
		return ""
	}
	return region
}
