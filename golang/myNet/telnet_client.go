package myNet

import (
	"errors"
	"io"
	"log"
	"net"
	"time"
)

const defaultBufferSize = 4096
const readTimeout = 1000

// TelnetClient represents a TCP client which is responsible for writing input data and printing response.
type TelnetClient struct {
	destination     *net.TCPAddr
	responseTimeout time.Duration
}

func NewTelnetClient(tcpAddr string) *TelnetClient {
	resolved := resolveTCPAddr(tcpAddr)

	return &TelnetClient{
		destination:     resolved,
		responseTimeout: 10 * time.Second, //defult timeout
	}
}

func resolveTCPAddr(addr string) *net.TCPAddr {
	resolved, error := net.ResolveTCPAddr("tcp", addr)
	if nil != error {
		log.Printf("解析连接地址异常！ \"%v\": %v\n", addr, error)
	}
	return resolved
}

// ProcessData method processes data: reads from input and writes to output.
func (t *TelnetClient) ProcessData(inputData string, outputData io.Writer) error {

	connection, error := net.DialTCP("tcp", nil, t.destination)
	if nil != error {
		log.Printf("建立连接失败！%s", t.destination.String())
		return errors.New("tcp建立连接失败")
	}
	log.Printf("连接创建成功！")

	defer connection.Close()

	requestDataChannel := make(chan []byte)
	doneChannel := make(chan bool)
	responseDataChannel := make(chan []byte)

	go t.readInputData(inputData, requestDataChannel, doneChannel)
	go t.readServerData(connection, responseDataChannel)

	var afterEOFResponseTicker = new(time.Ticker)

	for {
		select {
		case request := <-requestDataChannel:
			if _, error := connection.Write(request); nil != error {
				log.Printf("往连接写入数据异常: %v\n", error)
			}
		case <-doneChannel:
			afterEOFResponseTicker = time.NewTicker(readTimeout * time.Millisecond)
		case response := <-responseDataChannel:
			outputData.Write(response)
			afterEOFResponseTicker.Stop()
			return nil
		case <-afterEOFResponseTicker.C:
			log.Printf("读取超时！超时时间：%d\n", readTimeout)
			return nil
		}
	}
}

func (t *TelnetClient) readInputData(inputData string, toSent chan<- []byte, doneChannel chan<- bool) {
	toSent <- []byte(inputData)
	//t.assertEOF(error)
	doneChannel <- true
}

func (t *TelnetClient) readServerData(connection *net.TCPConn, received chan<- []byte) {
	buffer := make([]byte, defaultBufferSize)
	var error error
	var n int

	for nil == error {
		n, error = connection.Read(buffer)
		received <- buffer[:n]
	}

	t.assertEOF(error)
}

func (t *TelnetClient) assertEOF(error error) {
	if "EOF" != error.Error() {
		log.Printf("Error occured while operating on TCP socket: %v\n\n", error)
	}
}
