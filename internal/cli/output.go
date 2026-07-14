package cli

import (
	"encoding/json"
	"io"
)

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}

type stringList []string

func (values *stringList) String() string { return "" }
func (values *stringList) Set(value string) error {
	*values = append(*values, value)
	return nil
}
