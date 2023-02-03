package lib

import (
	"fmt"
	"strings"
)

type CustomHeaders []string

func (c *CustomHeaders) String() string {
	s := ""
	for _, v := range *c {
		s += v
	}
	return s
}

func (c *CustomHeaders) Set(value string) error {
	*c = append(*c, value)
	return nil
}

func (c *CustomHeaders) Has(header string) bool {
	for _, v := range *c {
		if strings.HasPrefix(strings.ToLower(v), strings.ToLower(header)) {
			return true
		}
	}
	return false
}

func (c *CustomHeaders) TransformMap() (map[string]string, error) {
	var header map[string]string
	for _, h := range *c {
		keyAndValue := strings.SplitN(h, ":", 2)
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("invalid header format for header %q", h)
		}
		key := strings.TrimSpace(keyAndValue[0])
		value := strings.TrimSpace(keyAndValue[1])
		if len(key) == 0 {
			return header, fmt.Errorf("invalid header format for header %q - name is empty", h)
		}
		header[key] = value
	}
	return header, nil
}
