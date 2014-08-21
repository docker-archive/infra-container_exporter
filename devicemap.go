package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	procDiskStats    = "/proc/diskstats"
	unknownFormatStr = "unknown(%d,%d)"
)

type deviceMap map[int]map[int]string

func newDeviceMap(filename string) (dm deviceMap, err error) {
	dm = make(deviceMap)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := string(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) <= 3 {
			return nil, fmt.Errorf("Invalid line in %s: %s", filename, line)
		}
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if _, ok := dm[major]; !ok {
			dm[major] = map[int]string{}
		}
		dm[major][minor] = parts[3]
	}
	return dm, nil
}

func (dm deviceMap) name(major, minor uint64) string {
	name, ok := dm[int(major)][int(minor)]
	if ok {
		return name
	}
	return fmt.Sprintf(unknownFormatStr, major, minor)
}
