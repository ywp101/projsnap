package utils

import (
	"fmt"
	"strings"
)

func Flatten2d(strs2d [][]string) []string {
	// flatten 2d slice to 1d slice
	var results []string
	for _, strs := range strs2d {
		for _, str := range strs {
			str = strings.TrimSpace(str)
			if str == "" {
				continue
			}
			results = append(results, str)
		}
	}
	return results
}

func SliceSplit(strs []string, sep string, expectLen, expectIdx int) ([]string, error) {
	results := make([]string, 0)
	for _, str := range strs {
		str = strings.TrimSpace(str)
		if str == "" {
			continue
		}
		tmp := strings.Split(str, sep)
		if expectLen != len(tmp) {
			return nil, fmt.Errorf("overflow for %s, length: %d", str, len(tmp))
		}
		results = append(results, tmp[expectIdx])
	}
	return results, nil
}

func FindMatched(strs []string, match string) []string {
	results := make([]string, 0)
	match = strings.ToLower(match)
	for _, str := range strs {
		str = strings.ToLower(strings.TrimSpace(str))
		if str == "" {
			continue
		}
		if strings.Contains(str, match) {
			results = append(results, str)
		}
	}
	return results
}
