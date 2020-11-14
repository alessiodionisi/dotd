package server

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strconv"

	"github.com/adnsio/dotd/pkg/roundrobin"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/dns/dnsmessage"
)

type Server struct {
	address *net.UDPAddr
	roundr  *roundrobin.RoundRobin
	conn    *net.UDPConn
	resolve map[string]string
}

func (s *Server) ListenAndServe() error {
	var err error
	s.conn, err = net.ListenUDP("udp", s.address)
	if err != nil {
		return err
	}
	defer s.conn.Close()

	log.Info().Msgf("listening on %s", s.address.String())

	exit := make(chan bool)
	for i := 0; i < runtime.NumCPU(); i++ {
		go s.listen()
	}
	<-exit

	return nil
}

func (s *Server) listen() {
	data := make([]byte, 1024)
	for {
		length, addr, err := s.conn.ReadFromUDP(data)
		if err != nil {
			log.Err(err).Stack().Caller().Send()
			continue
		}

		log.Debug().
			Int("length", length).
			Msgf("received data from %s", addr.String())

		go func() {
			err := s.answer(addr, data[:length])
			if err != nil {
				log.Err(err).Stack().Caller().Send()
			}
		}()
	}
}

func (s *Server) answer(addr *net.UDPAddr, data []byte) error {
	msg := &dnsmessage.Message{}

	err := msg.Unpack(data)
	if err != nil {
		return err
	}

	firstQuestion := msg.Questions[0]

	log.Debug().
		Uint16("id", msg.ID).
		Str("name", firstQuestion.Name.String()).
		Str("type", firstQuestion.Type.String()).
		Msgf("dns question from %s", addr.String())

	name := firstQuestion.Name.Data[:firstQuestion.Name.Length-1]
	resolved, ok := s.resolve[string(name)]
	if ok && resolved != "" {
		log.Debug().
			Uint16("id", msg.ID).
			Msgf("resolving %s", firstQuestion.Name.String())

		resolvedIP := net.ParseIP(resolved)
		if resolvedIP == nil {
			return fmt.Errorf(`invalid ip address "%s"`, resolved)
		}

		msg.Header.Response = true
		msg.Header.RecursionAvailable = true

		switch firstQuestion.Type {
		case dnsmessage.TypeA:
			ip4 := resolvedIP.To4()
			if ip4 == nil {
				if err := s.writeDNSMessage(addr, msg); err != nil {
					return err
				}

				break
			}

			var a [4]byte
			copy(a[:], ip4)

			msg.Answers = []dnsmessage.Resource{
				{
					Header: dnsmessage.ResourceHeader{
						Name:  firstQuestion.Name,
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.AResource{
						A: a,
					},
				},
			}
		case dnsmessage.TypeAAAA:
			if resolvedIP.To4() != nil {
				if err := s.writeDNSMessage(addr, msg); err != nil {
					return err
				}

				break
			}

			var aaaa [16]byte
			copy(aaaa[:], resolvedIP)

			msg.Answers = []dnsmessage.Resource{
				{
					Header: dnsmessage.ResourceHeader{
						Name:  firstQuestion.Name,
						Type:  dnsmessage.TypeAAAA,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.AAAAResource{
						AAAA: aaaa,
					},
				},
			}
		default:
			if err := s.writeDNSMessage(addr, msg); err != nil {
				return err
			}
		}

		if err := s.writeDNSMessage(addr, msg); err != nil {
			return err
		}

		log.Debug().
			Uint16("id", msg.ID).
			Msg("dns question answered")

	} else {
		if err := s.forwardToUpstream(addr, data, msg.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) writeDNSMessage(addr *net.UDPAddr, msg *dnsmessage.Message) error {
	resData, err := msg.Pack()
	if err != nil {
		return err
	}

	_, err = s.conn.WriteToUDP(resData, addr)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) forwardToUpstream(addr *net.UDPAddr, data []byte, id uint16) error {
	for i := 0; i < s.roundr.Length(); i++ {
		upstream, err := s.roundr.Pick()
		if err != nil {
			return err
		}

		log.Debug().
			Uint16("id", id).
			Int("attempt", i+1).
			Int("upstreams", s.roundr.Length()).
			Msgf("forwarding request to %s", upstream.String())

		dataReader := bytes.NewReader(data)
		req, err := http.NewRequest(http.MethodPost, upstream.String(), dataReader)
		if err != nil {
			return err
		}

		req.Header.Add("content-type", "application/dns-message")
		req.Header.Add("accept", "application/dns-message")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Err(err).Stack().Caller().Send()
			continue
		}

		if res.StatusCode != 200 {
			log.Err(errors.New("invalid status code")).Stack().Caller().Send()
			continue
		}

		resData, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		_, err = s.conn.WriteToUDP(resData, addr)
		if err != nil {
			return err
		}

		log.Debug().
			Uint16("id", id).
			Msg("dns question answered")

		return nil
	}

	return errors.New("max attempts reached")
}

func parseAddress(address string) (*net.UDPAddr, error) {
	host, stringPort, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(stringPort)
	if err != nil {
		return nil, err
	}

	parsedIP := net.ParseIP(host)
	if parsedIP == nil {
		return nil, fmt.Errorf(`"%s" is not a valid ip address`, host)
	}

	return &net.UDPAddr{
		IP:   parsedIP,
		Port: port,
	}, nil
}

func New(address string, upstreams []string, resolve map[string]string) (*Server, error) {
	udpAddress, err := parseAddress(address)
	if err != nil {
		return nil, err
	}

	upstreamURLs := make([]*url.URL, 0, len(upstreams))
	for _, upstream := range upstreams {
		upstreamURL, err := url.Parse(upstream)
		if err != nil {
			return nil, err
		}

		upstreamURLs = append(upstreamURLs, upstreamURL)
	}

	upstreamRR := roundrobin.New(upstreamURLs)

	return &Server{
		address: udpAddress,
		roundr:  upstreamRR,
		resolve: resolve,
	}, nil
}
