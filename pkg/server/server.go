package server

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
)

//HandlerFunc ..
type HandlerFunc func(req *Request)

var (
	//ErrBadRequest ...
	ErrBadRequest = errors.New("Bad Request")
	//ErrMethodNotAlowed ..
	ErrMethodNotAlowed = errors.New("Method not Alowed")
	//ErrHTTPVersionNotValid ..
	ErrHTTPVersionNotValid = errors.New("Http version not valid")
)

//Server ..
type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

//Request ...
type Request struct {
	Conn        net.Conn
	QueryParams url.Values
}

//NewServer ...
func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
}

//Register ...
func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

//Start ...
func (s *Server) Start() (err error) {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			err = cerr
			return
		}
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go s.handle(conn)

	}

	//return nil
}

func (s *Server) handle(conn net.Conn) {

	defer conn.Close()

	buf := make([]byte, (1024 * 8))
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			log.Printf("%s", buf[:n])
		}
		if err != nil {
			log.Println(err)
			return
		}

		var req Request
		data := buf[:n]
		rLD := []byte{'\r', '\n'}
		rLE := bytes.Index(data, rLD)
		if rLE == -1 {
			log.Printf("Bad Request")
			return
		}

		reqLine := string(data[:rLE])
		parts := strings.Split(reqLine, " ")

		if len(parts) != 3 {
			log.Println(ErrBadRequest)
			return
		}
		//method, path, version := parts[0], parts[1], parts[2]
		path, version := parts[1], parts[2]
		if version != "HTTP/1.1" {
			log.Println(ErrHTTPVersionNotValid)
			return
		}

		decode, err := url.PathUnescape(path)
		if err != nil {
			log.Println(err)
			return
		}

		uri, err := url.ParseRequestURI(decode)
		if err != nil {
			log.Println(err)
			return
		}

		req.Conn = conn
		req.QueryParams = uri.Query()

		var handler = func(req *Request) { conn.Close() }

		s.mu.RLock()
		for i := 0; i < len(s.handlers); i++ {
			if hr, found := s.handlers[uri.Path]; found {
				handler = hr
				break
			}
		}
		s.mu.RUnlock()

		handler(&req)

	}

}
