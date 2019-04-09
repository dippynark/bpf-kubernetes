package bpf

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/iovisor/gobpf/elf"
)

func Load(assetName string) (*elf.Module, error) {

	buf, err := Asset(assetName)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(buf)

	m := elf.NewModuleFromReader(reader)
	if m == nil {
		return nil, errors.New("failed to create new module from reader")
	}

	err = m.Load(map[string]elf.SectionParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to load BPF programs and maps: %s", err)
	}

	return m, nil
}
