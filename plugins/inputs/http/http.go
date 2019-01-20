package http

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/internal"
	"github.com/lavaorg/telex/internal/tls"
	"github.com/lavaorg/telex/plugins/inputs"
	"github.com/lavaorg/telex/plugins/parsers"
)

type HTTP struct {
	URLs            []string `toml:"urls"`
	Method          string   `toml:"method"`
	Body            string   `toml:"body"`
	ContentEncoding string   `toml:"content_encoding"`

	Headers map[string]string `toml:"headers"`

	// HTTP Basic Auth Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`
	tls.ClientConfig

	Timeout internal.Duration `toml:"timeout"`

	client *http.Client

	// The parser will automatically be set by telex core code because
	// this plugin implements the ParserInput interface (i.e. the SetParser method)
	parser parsers.Parser
}



// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (h *HTTP) Gather(acc telex.Accumulator) error {
	if h.parser == nil {
		return errors.New("Parser is not set")
	}

	if h.client == nil {
		tlsCfg, err := h.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		h.client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
				Proxy:           http.ProxyFromEnvironment,
			},
			Timeout: h.Timeout.Duration,
		}
	}

	var wg sync.WaitGroup
	for _, u := range h.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := h.gatherURL(acc, url); err != nil {
				acc.AddError(fmt.Errorf("[url=%s]: %s", url, err))
			}
		}(u)
	}

	wg.Wait()

	return nil
}

// SetParser takes the data_format from the config and finds the right parser for that format
func (h *HTTP) SetParser(parser parsers.Parser) {
	h.parser = parser
}

// Gathers data from a particular URL
// Parameters:
//     acc    : The telex Accumulator to use
//     url    : endpoint to send request to
//
// Returns:
//     error: Any error that may have occurred
func (h *HTTP) gatherURL(
	acc telex.Accumulator,
	url string,
) error {
	body, err := makeRequestBodyReader(h.ContentEncoding, h.Body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(h.Method, url, body)
	if err != nil {
		return err
	}

	if h.ContentEncoding == "gzip" {
		request.Header.Set("Content-Encoding", "gzip")
	}

	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			request.Host = v
		} else {
			request.Header.Add(k, v)
		}
	}

	if h.Username != "" || h.Password != "" {
		request.SetBasicAuth(h.Username, h.Password)
	}

	resp, err := h.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received status code %d (%s), expected %d (%s)",
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	metrics, err := h.parser.Parse(b)
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		if !metric.HasTag("url") {
			metric.AddTag("url", url)
		}
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}

	return nil
}

func makeRequestBodyReader(contentEncoding, body string) (io.Reader, error) {
	var err error
	var reader io.Reader = strings.NewReader(body)
	if contentEncoding == "gzip" {
		reader, err = internal.CompressWithGzip(reader)
		if err != nil {
			return nil, err
		}
	}
	return reader, nil
}

func init() {
	inputs.Add("http", func() telex.Input {
		return &HTTP{
			Timeout: internal.Duration{Duration: time.Second * 5},
			Method:  "GET",
		}
	})
}
