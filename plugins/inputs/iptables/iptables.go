// +build linux

package iptables

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

// Iptables is a telex plugin to gather packets and bytes throughput from Linux's iptables packet filter.
type Iptables struct {
	UseSudo bool
	UseLock bool
	Binary  string
	Table   string
	Chains  []string
	lister  chainLister
}

// Gather gathers iptables packets and bytes throughput from the configured tables and chains.
func (ipt *Iptables) Gather(acc telex.Accumulator) error {
	if ipt.Table == "" || len(ipt.Chains) == 0 {
		return nil
	}
	// best effort : we continue through the chains even if an error is encountered,
	// but we keep track of the last error.
	for _, chain := range ipt.Chains {
		data, e := ipt.lister(ipt.Table, chain)
		if e != nil {
			acc.AddError(e)
			continue
		}
		e = ipt.parseAndGather(data, acc)
		if e != nil {
			acc.AddError(e)
			continue
		}
	}
	return nil
}

func (ipt *Iptables) chainList(table, chain string) (string, error) {
	var binary string
	if ipt.Binary != "" {
		binary = ipt.Binary
	} else {
		binary = "iptables"
	}
	iptablePath, err := exec.LookPath(binary)
	if err != nil {
		return "", err
	}
	var args []string
	name := iptablePath
	if ipt.UseSudo {
		name = "sudo"
		args = append(args, iptablePath)
	}
	iptablesBaseArgs := "-nvL"
	if ipt.UseLock {
		iptablesBaseArgs = "-wnvL"
	}
	args = append(args, iptablesBaseArgs, chain, "-t", table, "-x")
	c := exec.Command(name, args...)
	out, err := c.Output()
	return string(out), err
}

const measurement = "iptables"

var errParse = errors.New("Cannot parse iptables list information")
var chainNameRe = regexp.MustCompile(`^Chain\s+(\S+)`)
var fieldsHeaderRe = regexp.MustCompile(`^\s*pkts\s+bytes\s+`)
var valuesRe = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+.*?/\*\s*(.+?)\s*\*/\s*`)

func (ipt *Iptables) parseAndGather(data string, acc telex.Accumulator) error {
	lines := strings.Split(data, "\n")
	if len(lines) < 3 {
		return nil
	}
	mchain := chainNameRe.FindStringSubmatch(lines[0])
	if mchain == nil {
		return errParse
	}
	if !fieldsHeaderRe.MatchString(lines[1]) {
		return errParse
	}
	for _, line := range lines[2:] {
		matches := valuesRe.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		pkts := matches[1]
		bytes := matches[2]
		comment := matches[3]

		tags := map[string]string{"table": ipt.Table, "chain": mchain[1], "ruleid": comment}
		fields := make(map[string]interface{})

		var err error
		fields["pkts"], err = strconv.ParseUint(pkts, 10, 64)
		if err != nil {
			continue
		}
		fields["bytes"], err = strconv.ParseUint(bytes, 10, 64)
		if err != nil {
			continue
		}
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}

type chainLister func(table, chain string) (string, error)

func init() {
	inputs.Add("iptables", func() telex.Input {
		ipt := new(Iptables)
		ipt.lister = ipt.chainList
		return ipt
	})
}
