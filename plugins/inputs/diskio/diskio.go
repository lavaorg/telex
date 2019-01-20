package diskio

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/filter"
	"github.com/lavaorg/telex/plugins/inputs"
	"github.com/lavaorg/telex/plugins/inputs/system"
)

var (
	varRegex = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)
)

type DiskIO struct {
	ps system.PS

	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool

	infoCache    map[string]diskInfoCache
	deviceFilter filter.Filter
	initialized  bool
}

// hasMeta reports whether s contains any special glob characters.
func hasMeta(s string) bool {
	return strings.IndexAny(s, "*?[") >= 0
}

func (s *DiskIO) init() error {
	for _, device := range s.Devices {
		if hasMeta(device) {
			filter, err := filter.Compile(s.Devices)
			if err != nil {
				return fmt.Errorf("error compiling device pattern: %v", err)
			}
			s.deviceFilter = filter
		}
	}
	s.initialized = true
	return nil
}

func (s *DiskIO) Gather(acc telex.Accumulator) error {
	if !s.initialized {
		err := s.init()
		if err != nil {
			return err
		}
	}

	devices := []string{}
	if s.deviceFilter == nil {
		devices = s.Devices
	}

	diskio, err := s.ps.DiskIO(devices)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err)
	}

	for _, io := range diskio {
		if s.deviceFilter != nil && !s.deviceFilter.Match(io.Name) {
			continue
		}

		tags := map[string]string{}
		tags["name"] = s.diskName(io.Name)
		for t, v := range s.diskTags(io.Name) {
			tags[t] = v
		}
		if !s.SkipSerialNumber {
			if len(io.SerialNumber) != 0 {
				tags["serial"] = io.SerialNumber
			} else {
				tags["serial"] = "unknown"
			}
		}

		fields := map[string]interface{}{
			"reads":            io.ReadCount,
			"writes":           io.WriteCount,
			"read_bytes":       io.ReadBytes,
			"write_bytes":      io.WriteBytes,
			"read_time":        io.ReadTime,
			"write_time":       io.WriteTime,
			"io_time":          io.IoTime,
			"weighted_io_time": io.WeightedIO,
			"iops_in_progress": io.IopsInProgress,
		}
		acc.AddCounter("diskio", fields, tags)
	}

	return nil
}

func (s *DiskIO) diskName(devName string) string {
	if len(s.NameTemplates) == 0 {
		return devName
	}

	di, err := s.diskInfo(devName)
	if err != nil {
		log.Printf("W! Error gathering disk info: %s", err)
		return devName
	}

	for _, nt := range s.NameTemplates {
		miss := false
		name := varRegex.ReplaceAllStringFunc(nt, func(sub string) string {
			sub = sub[1:] // strip leading '$'
			if sub[0] == '{' {
				sub = sub[1 : len(sub)-1] // strip leading & trailing '{' '}'
			}
			if v, ok := di[sub]; ok {
				return v
			}
			miss = true
			return ""
		})

		if !miss {
			return name
		}
	}

	return devName
}

func (s *DiskIO) diskTags(devName string) map[string]string {
	if len(s.DeviceTags) == 0 {
		return nil
	}

	di, err := s.diskInfo(devName)
	if err != nil {
		log.Printf("W! Error gathering disk info: %s", err)
		return nil
	}

	tags := map[string]string{}
	for _, dt := range s.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("diskio", func() telex.Input {
		return &DiskIO{ps: ps, SkipSerialNumber: true}
	})
}
