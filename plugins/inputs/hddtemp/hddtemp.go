package hddtemp

import (
	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/plugins/inputs"
	gohddtemp "github.com/lavaorg/telex/plugins/inputs/hddtemp/go-hddtemp"
)

const defaultAddress = "127.0.0.1:7634"

type HDDTemp struct {
	Address string
	Devices []string
	fetcher Fetcher
}

type Fetcher interface {
	Fetch(address string) ([]gohddtemp.Disk, error)
}

func (h *HDDTemp) Gather(acc telex.Accumulator) error {
	if h.fetcher == nil {
		h.fetcher = gohddtemp.New()
	}
	disks, err := h.fetcher.Fetch(h.Address)

	if err != nil {
		return err
	}

	for _, disk := range disks {
		for _, chosenDevice := range h.Devices {
			if chosenDevice == "*" || chosenDevice == disk.DeviceName {
				tags := map[string]string{
					"device": disk.DeviceName,
					"model":  disk.Model,
					"unit":   disk.Unit,
					"status": disk.Status,
				}

				fields := map[string]interface{}{
					"temperature": disk.Temperature,
				}

				acc.AddFields("hddtemp", fields, tags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("hddtemp", func() telex.Input {
		return &HDDTemp{
			Address: defaultAddress,
			Devices: []string{"*"},
		}
	})
}
