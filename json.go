package tdclient

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"

	"github.com/pkg/errors"
)

var JSON = jsoniter.ConfigCompatibleWithStandardLibrary

func debugJSON(v interface{}) string {
	b := new(bytes.Buffer)
	debugEncoder := JSON.NewEncoder(b)
	debugEncoder.SetIndent("", "    ")
	debugEncoder.SetEscapeHTML(false)
	if err := debugEncoder.Encode(v); err != nil {
		tdclientGlgLogger.Fatalln(errors.Wrap(err, "json marshall fail"))
	}
	return b.String()
}
