package myNet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/wyzzgzhdcxy/wcj-go-common/core"
)

var icmpPacket ICMP

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

func Ping(ip string) bool {
	//开始填充数据包
	icmpPacket.Type = 8 //8->echo message  0->reply message
	icmpPacket.Code = 0
	icmpPacket.Checksum = 0
	icmpPacket.Identifier = 0
	icmpPacket.SequenceNum = 0

	recvBuf := make([]byte, 32)
	var buffer bytes.Buffer

	//先在buffer中写入icmp数据报求去校验和
	err := binary.Write(&buffer, binary.BigEndian, icmpPacket)
	if err != nil {
		return false
	}
	icmpPacket.Checksum = CheckSum(buffer.Bytes())
	//然后清空buffer并把求完校验和的icmp数据报写入其中准备发送
	buffer.Reset()
	err = binary.Write(&buffer, binary.BigEndian, icmpPacket)
	if err != nil {
		return false
	}
	Time, _ := time.ParseDuration("5s")
	conn, err := net.DialTimeout("ip4:icmp", ip, Time)
	if err != nil {
		return false
	}
	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		//log.Println("conn.Write error:", err)
		return false
	}
	err = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	if err != nil {
		return false
	}
	num, err := conn.Read(recvBuf)
	if err != nil {
		//log.Println("conn.Read error:", err)
		return false
	}
	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		return false
	}
	if string(recvBuf[0:num]) != "" {
		return true
	}
	return false

}

func CheckSum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += sum >> 16
	return uint16(^sum)
}

// PortParse 端口遍历
func PortParse(port string) []int {
	var resultPorts []int
	ports := strings.Split(port, "-")
	if len(ports) < 1 {
		return resultPorts
	}
	portOne, _ := strconv.Atoi(ports[0])
	if len(ports) == 1 {
		resultPorts = append(resultPorts, portOne)
		return resultPorts
	}
	portTwo, _ := strconv.Atoi(ports[1])
	if portOne > portTwo {
		tmp := portOne
		portOne = portTwo
		portTwo = tmp
	}
	for i := portOne; i <= portTwo; i++ {
		resultPorts = append(resultPorts, i)
	}
	return resultPorts
}

// IpParse ip地址解析
func IpParse(ip string) []string {
	var resultIps []string
	if !Ipv4Valid(ip) {
		return resultIps
	}
	ipOne := strings.Split(ip, ".")
	ipOneLast, _ := strconv.Atoi(ipOne[3])
	for i := ipOneLast; i <= 255; i++ {
		rip := fmt.Sprintf("%s.%s.%s.%d", ipOne[0], ipOne[1], ipOne[2], i)
		resultIps = append(resultIps, rip)
	}
	return resultIps
}

// Ipv4Valid ipv4验证
func Ipv4Valid(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])$`
	match, _ := regexp.MatchString(ipReg, ip)
	if match {
		return true
	}
	return false
}

var pool *core.Pool

var total, success, fail int64

// PingIps ping ip地址
func PingIps(ips []string, queueNum int, callback func(msg []byte)) {
	if len(ips) < 1 {
		return
	}
	if queueNum < 1 {
		queueNum = 10
	}
	pool = core.NewPool(queueNum)
	for _, ip := range ips {
		pool.Add(1)
		go ping(ip, callback)
	}
	pool.Wait()
	fmt.Println("ip总数: ", total)
	fmt.Println("ip可访问: ", success)
	fmt.Println("ip不可访问: ", fail)
}

type IpConnInfo struct {
	Time string
	Ip   string
	Pass bool
}

type ConnState struct {
	Message string
	Status  int
}

func ping(ip string, callback func(msg []byte)) {
	total++
	b := Ping(ip)
	// 格式化时间
	formattedTime := core.TimeFormat(time.Now())
	connInfo := IpConnInfo{Ip: ip, Pass: b, Time: formattedTime}
	jsonBy, _ := json.Marshal(connInfo)
	callback(jsonBy)
	if b == true {
		success++
	} else {
		fail++
	}
	pool.Done()
}
