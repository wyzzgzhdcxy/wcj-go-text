package myNet

import (
	"fmt"
	"log"
	"net"

	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

// GetMacAdders 获取所有的mac地址
func GetMacAdders() (macAdders []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("获取所有的mac地址异常: %v", err)
		return macAdders
	}
	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		log.Printf("netInterface,%v", netInterface)
		if len(macAddr) == 0 {
			continue
		}
		macAdders = append(macAdders, macAddr)
	}
	return macAdders
}

func GetMacAdder() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("获取所有的mac地址异常: %v", err)
		return ""
	}

	for _, iface := range interfaces {
		// 只获取物理网卡的MAC地址
		if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagUp != 0 {
			if iface.Name != "以太网" {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				fmt.Println("Failed to retrieve addresses for interface", iface.Name, ":", err)
				continue
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					hwAddr := iface.HardwareAddr
					log.Printf("Interface:%s,MAC address:%s", iface.Name, hwAddr)
					return hwAddr.String()
				}
			}
		}
	}
	return ""
}

func GetIpList() {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Failed to retrieve network interfaces:", err)
		return
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Failed to retrieve addresses for interface", iface.Name, ":", err)
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					fmt.Println("Interface:", iface.Name)
					fmt.Println("IP Address:", ipNet.IP)
				}
			}
		}
	}
}

func GetMacMd5() string {
	return core.Md5Byte([]byte(GetMacAdder()))
}
