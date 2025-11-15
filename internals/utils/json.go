package utils

import (
	"encoding/json"
	"io"
)

func PrettyJSON(w io.Writer, payload any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}
