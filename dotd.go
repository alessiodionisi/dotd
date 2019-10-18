package dotd

import (
	"bytes"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

type Dotd struct {
	Config  *Config
	udpConn *net.UDPConn
}

type Config struct {
	Addr     string
	Upstream string
	Logs     bool
}

func New(cfg *Config) *Dotd {
	return &Dotd{
		Config: cfg,
	}
}

func (dd *Dotd) Listen() {
	udpAddr, err := net.ResolveUDPAddr("udp", dd.Config.Addr)
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	dd.udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}
	defer dd.udpConn.Close()

	buf := make([]byte, 512)

	log.Printf("server: listening on %s\n", udpAddr)
	for {
		bufLen, peerAddr, err := dd.udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}

		go dd.answerMessage(buf[:bufLen], peerAddr)
	}
}

func (dd *Dotd) answerMessage(bt []byte, addr *net.UDPAddr) {
	if dd.Config.Logs {
		msg := &dnsmessage.Message{}
		err := msg.Unpack(bt)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
		log.Printf("dns: <- %s ID: %d Q: %+v\n", addr, msg.ID, msg.Questions)
	}

	reqRdr := bytes.NewReader(bt)
	req, err := http.NewRequest(http.MethodPost, dd.Config.Upstream, reqRdr)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	req.Header.Add("content-type", "application/dns-message")
	req.Header.Add("accept", "application/dns-message")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	_, err = dd.udpConn.WriteToUDP(resBody, addr)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	if dd.Config.Logs {
		resMsg := &dnsmessage.Message{}
		err = resMsg.Unpack(resBody)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
		log.Printf("dns: -> %s ID: %d A: %+v\n", addr, resMsg.ID, resMsg.Answers)
	}
}
