package interrupts

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
)

type Interrupts struct {
	CpuAsTag bool `toml:"cpu_as_tag"`
}

type IRQ struct {
	ID     string
	Type   string
	Device string
	Total  int64
	Cpus   []int64
}

func NewIRQ(id string) *IRQ {
	return &IRQ{ID: id, Cpus: []int64{}}
}

func parseInterrupts(r io.Reader) ([]IRQ, error) {
	var irqs []IRQ
	var cpucount int
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		cpus := strings.Fields(scanner.Text())
		if cpus[0] != "CPU0" {
			return nil, fmt.Errorf("Expected first line to start with CPU0, but was %s", scanner.Text())
		}
		cpucount = len(cpus)
	}

scan:
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if !strings.HasSuffix(fields[0], ":") {
			continue
		}
		irqid := strings.TrimRight(fields[0], ":")
		irq := NewIRQ(irqid)
		irqvals := fields[1:]
		for i := 0; i < cpucount; i++ {
			if i < len(irqvals) {
				irqval, err := strconv.ParseInt(irqvals[i], 10, 64)
				if err != nil {
					continue scan
				}
				irq.Cpus = append(irq.Cpus, irqval)
			}
		}
		for _, irqval := range irq.Cpus {
			irq.Total += irqval
		}
		_, err := strconv.ParseInt(irqid, 10, 64)
		if err == nil && len(fields) >= cpucount+2 {
			irq.Type = fields[cpucount+1]
			irq.Device = strings.Join(fields[cpucount+2:], " ")
		} else if len(fields) > cpucount {
			irq.Type = strings.Join(fields[cpucount+1:], " ")
		}
		irqs = append(irqs, *irq)
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("Error scanning file: %s", scanner.Err())
	}
	return irqs, nil
}

func gatherTagsFields(irq IRQ) (map[string]string, map[string]interface{}) {
	tags := map[string]string{"irq": irq.ID, "type": irq.Type, "device": irq.Device}
	fields := map[string]interface{}{"total": irq.Total}
	for i := 0; i < len(irq.Cpus); i++ {
		cpu := fmt.Sprintf("cpu%d", i)
		fields[cpu] = irq.Cpus[i]
	}
	return tags, fields
}

func (s *Interrupts) Gather(acc telex.Accumulator) error {
	for measurement, file := range map[string]string{"interrupts": "/proc/interrupts", "soft_interrupts": "/proc/softirqs"} {
		f, err := os.Open(file)
		if err != nil {
			acc.AddError(fmt.Errorf("Could not open file: %s", file))
			continue
		}
		defer f.Close()
		irqs, err := parseInterrupts(f)
		if err != nil {
			acc.AddError(fmt.Errorf("Parsing %s: %s", file, err))
			continue
		}
		reportMetrics(measurement, irqs, acc, s.CpuAsTag)
	}
	return nil
}

func reportMetrics(measurement string, irqs []IRQ, acc telex.Accumulator, cpusAsTags bool) {
	for _, irq := range irqs {
		tags, fields := gatherTagsFields(irq)
		if cpusAsTags {
			for cpu, count := range irq.Cpus {
				cpuTags := map[string]string{"cpu": fmt.Sprintf("cpu%d", cpu)}
				for k, v := range tags {
					cpuTags[k] = v
				}
				acc.AddFields(measurement, map[string]interface{}{"count": count}, cpuTags)
			}
		} else {
			acc.AddFields(measurement, fields, tags)
		}
	}
}

func init() {
	inputs.Add("interrupts", func() telex.Input {
		return &Interrupts{}
	})
}
