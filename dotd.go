package dotd

import (
	"bytes"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

type fnLog func(format string, v ...interface{}) //log function

type Dotd struct {
	Config  *Config
	udpConn *net.UDPConn
}

type Config struct {
	Addr     string
	Upstream string
	Verbose  bool
	FileLog  string
}

func New(cfg *Config) *Dotd {
	return &Dotd{
		Config: cfg,
	}
}

func (dd *Dotd) Listen() {
	//Logging to a file
	if dd.Config.FileLog != "" {
		f, err := os.OpenFile(dd.Config.FileLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
		dd.Log(log.Printf, "server: starting to log\n")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", dd.Config.Addr)
	if err != nil {
		dd.Log(log.Fatalf, "error: %s\n", err)
	}

	dd.udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		dd.Log(log.Fatalf, "error: %s\n", err)
	}
	defer dd.udpConn.Close()

	buf := make([]byte, 512)

	dd.Log(log.Printf, "server: listening on %s\n", udpAddr)
	for {
		bufLen, peerAddr, err := dd.udpConn.ReadFromUDP(buf)
		if err != nil {
			dd.Log(log.Printf, "error: %s\n", err)
		}

		go dd.answerMessage(buf[:bufLen], peerAddr)
	}
}

func (dd *Dotd) answerMessage(bt []byte, addr *net.UDPAddr) {
	if dd.Config.Verbose {
		msg := &dnsmessage.Message{}
		err := msg.Unpack(bt)

		if err != nil {
			dd.Log(log.Printf, "error: %s\n", err)
		} else {
			dd.Log(log.Printf, "dns: <- %s ID: %d Q: %+v\n", addr, msg.ID, msg.Questions)
		}
	}

	reqRdr := bytes.NewReader(bt)
	req, err := http.NewRequest(http.MethodPost, dd.Config.Upstream, reqRdr)
	if err != nil {
		dd.Log(log.Printf, "error: %s\n", err)
	}

	req.Header.Add("content-type", "application/dns-message")
	req.Header.Add("accept", "application/dns-message")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		dd.Log(log.Printf, "error: %s\n", err)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		dd.Log(log.Printf, "error: %s\n", err)
	}

	_, err = dd.udpConn.WriteToUDP(resBody, addr)
	if err != nil {
		dd.Log(log.Printf, "error: %s\n", err)
	}

	if dd.Config.Verbose {
		resMsg := &dnsmessage.Message{}
		err = resMsg.Unpack(resBody)
		if err != nil {
			dd.Log(log.Printf, "error: %s\n", err)
		}
		dd.Log(log.Printf, "dns: -> %s ID: %d A: %+v\n", addr, resMsg.ID, resMsg.Answers)
	}
}

func (dd *Dotd) Log(wLog fnLog, format string, v ...interface{}) {
	if dd.Config.Verbose {
		fmt.Printf(format, v...)
	}

	if dd.Config.FileLog != "" {
		wLog(format, v...)
	}
}
