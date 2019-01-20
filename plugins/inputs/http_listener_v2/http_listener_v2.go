package http_listener_v2

import (
	"compress/gzip"
	"crypto/subtle"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/internal"
	tlsint "github.com/lavaorg/telex/internal/tls"
	"github.com/lavaorg/telex/plugins/inputs"
	"github.com/lavaorg/telex/plugins/parsers"
)

// defaultMaxBodySize is the default maximum request body size, in bytes.
// if the request body is over this size, we will return an HTTP 413 error.
// 500 MB
const defaultMaxBodySize = 500 * 1024 * 1024

type TimeFunc func() time.Time

type HTTPListenerV2 struct {
	ServiceAddress string
	Path           string
	Methods        []string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration
	MaxBodySize    internal.Size
	Port           int

	tlsint.ServerConfig

	BasicUsername string
	BasicPassword string

	TimeFunc

	wg sync.WaitGroup

	listener net.Listener

	parsers.Parser
	acc telex.Accumulator
}

func (h *HTTPListenerV2) Gather(_ telex.Accumulator) error {
	return nil
}

func (h *HTTPListenerV2) SetParser(parser parsers.Parser) {
	h.Parser = parser
}

// Start starts the http listener service.
func (h *HTTPListenerV2) Start(acc telex.Accumulator) error {
	if h.MaxBodySize.Size == 0 {
		h.MaxBodySize.Size = defaultMaxBodySize
	}

	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second * 10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}

	h.acc = acc

	tlsConf, err := h.ServerConfig.TLSConfig()
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:         h.ServiceAddress,
		Handler:      h,
		ReadTimeout:  h.ReadTimeout.Duration,
		WriteTimeout: h.WriteTimeout.Duration,
		TLSConfig:    tlsConf,
	}

	var listener net.Listener
	if tlsConf != nil {
		listener, err = tls.Listen("tcp", h.ServiceAddress, tlsConf)
	} else {
		listener, err = net.Listen("tcp", h.ServiceAddress)
	}
	if err != nil {
		return err
	}
	h.listener = listener
	h.Port = listener.Addr().(*net.TCPAddr).Port

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		server.Serve(h.listener)
	}()

	log.Printf("I! Started HTTP listener V2 service on %s\n", h.ServiceAddress)

	return nil
}

// Stop cleans up all resources
func (h *HTTPListenerV2) Stop() {
	h.listener.Close()
	h.wg.Wait()

	log.Println("I! Stopped HTTP listener V2 service on ", h.ServiceAddress)
}

func (h *HTTPListenerV2) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path == h.Path {
		h.AuthenticateIfSet(h.serveWrite, res, req)
	} else {
		h.AuthenticateIfSet(http.NotFound, res, req)
	}
}

func (h *HTTPListenerV2) serveWrite(res http.ResponseWriter, req *http.Request) {
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize.Size {
		tooLarge(res)
		return
	}

	// Check if the requested HTTP method was specified in config.
	isAcceptedMethod := false
	for _, method := range h.Methods {
		if req.Method == method {
			isAcceptedMethod = true
			break
		}
	}
	if !isAcceptedMethod {
		methodNotAllowed(res)
		return
	}

	// Handle gzip request bodies
	body := req.Body
	if req.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(req.Body)
		if err != nil {
			log.Println("D! " + err.Error())
			badRequest(res)
			return
		}
		defer body.Close()
	}

	body = http.MaxBytesReader(res, body, h.MaxBodySize.Size)
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		tooLarge(res)
		return
	}

	metrics, err := h.Parse(bytes)
	if err != nil {
		log.Println("D! " + err.Error())
		badRequest(res)
		return
	}
	for _, m := range metrics {
		h.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
	res.WriteHeader(http.StatusNoContent)
}

func tooLarge(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	res.Write([]byte(`{"error":"http: request body too large"}`))
}

func methodNotAllowed(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusMethodNotAllowed)
	res.Write([]byte(`{"error":"http: method not allowed"}`))
}

func internalServerError(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusInternalServerError)
}

func badRequest(res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(`{"error":"http: bad request"}`))
}

func (h *HTTPListenerV2) AuthenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if h.BasicUsername != "" && h.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {

			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
		handler(res, req)
	} else {
		handler(res, req)
	}
}

func init() {
	inputs.Add("http_listener_v2", func() telex.Input {
		return &HTTPListenerV2{
			ServiceAddress: ":8080",
			TimeFunc:       time.Now,
			Path:           "/telex",
			Methods:        []string{"POST", "PUT"},
		}
	})
}
