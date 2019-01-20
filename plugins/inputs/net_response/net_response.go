package net_response

import (
	"bufio"
	"errors"
	"net"
	"net/textproto"
	"regexp"
	"time"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/internal"
	"github.com/lavaorg/telex/plugins/inputs"
)

type ResultType uint64

const (
	Success          ResultType = 0
	Timeout                     = 1
	ConnectionFailed            = 2
	ReadFailed                  = 3
	StringMismatch              = 4
)

// NetResponse struct
type NetResponse struct {
	Address     string
	Timeout     internal.Duration
	ReadTimeout internal.Duration
	Send        string
	Expect      string
	Protocol    string
}

// TCPGather will execute if there are TCP tests defined in the configuration.
// It will return a map[string]interface{} for fields and a map[string]string for tags
func (n *NetResponse) TCPGather() (tags map[string]string, fields map[string]interface{}) {
	// Prepare returns
	tags = make(map[string]string)
	fields = make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Connecting
	conn, err := net.DialTimeout("tcp", n.Address, n.Timeout.Duration)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			setResult(Timeout, fields, tags, n.Expect)
		} else {
			setResult(ConnectionFailed, fields, tags, n.Expect)
		}
		return tags, fields
	}
	defer conn.Close()
	// Send string if needed
	if n.Send != "" {
		msg := []byte(n.Send)
		conn.Write(msg)
		// Stop timer
		responseTime = time.Since(start).Seconds()
	}
	// Read string if needed
	if n.Expect != "" {
		// Set read timeout
		conn.SetReadDeadline(time.Now().Add(n.ReadTimeout.Duration))
		// Prepare reader
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		// Read
		data, err := tp.ReadLine()
		// Stop timer
		responseTime = time.Since(start).Seconds()
		// Handle error
		if err != nil {
			setResult(ReadFailed, fields, tags, n.Expect)
		} else {
			// Looking for string in answer
			RegEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
			find := RegEx.FindString(string(data))
			if find != "" {
				setResult(Success, fields, tags, n.Expect)
			} else {
				setResult(StringMismatch, fields, tags, n.Expect)
			}
		}
	} else {
		setResult(Success, fields, tags, n.Expect)
	}
	fields["response_time"] = responseTime
	return tags, fields
}

// UDPGather will execute if there are UDP tests defined in the configuration.
// It will return a map[string]interface{} for fields and a map[string]string for tags
func (n *NetResponse) UDPGather() (tags map[string]string, fields map[string]interface{}) {
	// Prepare returns
	tags = make(map[string]string)
	fields = make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Resolving
	udpAddr, err := net.ResolveUDPAddr("udp", n.Address)
	// Connecting
	conn, err := net.DialUDP("udp", nil, udpAddr)
	// Handle error
	if err != nil {
		setResult(ConnectionFailed, fields, tags, n.Expect)
		return tags, fields
	}
	defer conn.Close()
	// Send string
	msg := []byte(n.Send)
	conn.Write(msg)
	// Read string
	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(n.ReadTimeout.Duration))
	// Read
	buf := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buf)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		setResult(ReadFailed, fields, tags, n.Expect)
		return tags, fields
	}

	// Looking for string in answer
	RegEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
	find := RegEx.FindString(string(buf))
	if find != "" {
		setResult(Success, fields, tags, n.Expect)
	} else {
		setResult(StringMismatch, fields, tags, n.Expect)
	}

	fields["response_time"] = responseTime

	return tags, fields
}

// Gather is called by telex when the plugin is executed on its interval.
// It will call either UDPGather or TCPGather based on the configuration and
// also fill an Accumulator that is supplied.
func (n *NetResponse) Gather(acc telex.Accumulator) error {
	// Set default values
	if n.Timeout.Duration == 0 {
		n.Timeout.Duration = time.Second
	}
	if n.ReadTimeout.Duration == 0 {
		n.ReadTimeout.Duration = time.Second
	}
	// Check send and expected string
	if n.Protocol == "udp" && n.Send == "" {
		return errors.New("Send string cannot be empty")
	}
	if n.Protocol == "udp" && n.Expect == "" {
		return errors.New("Expected string cannot be empty")
	}
	// Prepare host and port
	host, port, err := net.SplitHostPort(n.Address)
	if err != nil {
		return err
	}
	if host == "" {
		n.Address = "localhost:" + port
	}
	if port == "" {
		return errors.New("Bad port")
	}
	// Prepare data
	tags := map[string]string{"server": host, "port": port}
	var fields map[string]interface{}
	var returnTags map[string]string
	// Gather data
	if n.Protocol == "tcp" {
		returnTags, fields = n.TCPGather()
		tags["protocol"] = "tcp"
	} else if n.Protocol == "udp" {
		returnTags, fields = n.UDPGather()
		tags["protocol"] = "udp"
	} else {
		return errors.New("Bad protocol")
	}
	for key, value := range returnTags {
		tags[key] = value
	}
	// Merge the tags
	for k, v := range returnTags {
		tags[k] = v
	}
	// Add metrics
	acc.AddFields("net_response", fields, tags)
	return nil
}

func setResult(result ResultType, fields map[string]interface{}, tags map[string]string, expect string) {
	var tag string
	switch result {
	case Success:
		tag = "success"
	case Timeout:
		tag = "timeout"
	case ConnectionFailed:
		tag = "connection_failed"
	case ReadFailed:
		tag = "read_failed"
	case StringMismatch:
		tag = "string_mismatch"
	}

	tags["result"] = tag
	fields["result_code"] = uint64(result)

	// deprecated in 1.7; use result tag
	fields["result_type"] = tag

	// deprecated in 1.4; use result tag
	if expect != "" {
		fields["string_found"] = result == Success
	}
}

func init() {
	inputs.Add("net_response", func() telex.Input {
		return &NetResponse{}
	})
}
