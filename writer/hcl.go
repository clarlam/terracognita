package writer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	kitlog "github.com/go-kit/kit/log"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/log"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/fmtcmd"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/pkg/errors"
)

// HCLWriter is a Writer implementation that writes to
// a static map to then transform it to HCL
type HCLWriter struct {
	// TODO: Change it to "map[string]map[string]schema.ResourceData"
	Config map[string]interface{}
	w      io.Writer
}

// NewHCLWriter rerturns an HCLWriter initialization
func NewHCLWriter(w io.Writer) *HCLWriter {
	cfg := make(map[string]interface{})
	cfg["resource"] = make(map[string]map[string]interface{})
	return &HCLWriter{
		Config: cfg,
		w:      w,
	}
}

// Write expects a key similar to "aws_instance.your_name"
// repeated keys will report an error
func (hclw *HCLWriter) Write(key string, value interface{}) error {
	if key == "" {
		return errcode.ErrWriterRequiredKey
	}

	if value == nil {
		return errcode.ErrWriterRequiredValue
	}

	keys := strings.Split(key, ".")
	if len(keys) != 2 || (keys[0] == "" || keys[1] == "") {
		return errors.Wrapf(errcode.ErrWriterInvalidKey, "with key %q", key)
	}

	name := strings.Join(keys[1:], "")

	if _, ok := hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][name]; ok {
		return errors.Wrapf(errcode.ErrWriterAlreadyExistsKey, "with key %q", key)
	}

	if _, ok := hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]]; !ok {
		hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]] = make(map[string]interface{})
	}

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	log.Get().Log("func", "writer.Write(HCL)", "msg", "writing to internal config", "key", keys[0], "content", string(b))

	hclw.Config["resource"].(map[string]map[string]interface{})[keys[0]][name] = value

	return nil
}

// Sync writes the content of the Config to the
// internal w with the correct format
func (hclw *HCLWriter) Sync() error {
	logger := log.Get()
	logger = kitlog.With(logger, "func", "writer.Write(HCL)")
	b, err := json.Marshal(hclw.Config)
	if err != nil {
		return err
	}

	logger.Log("msg", "parsing internal config to HCL", "json", string(b))
	f, err := hcl.ParseBytes(b)
	if err != nil {
		return fmt.Errorf("error while 'hcl.ParseBytes': %s", err)
	}

	buff := &bytes.Buffer{}
	err = printer.Fprint(buff, f.Node)
	if err != nil {
		return fmt.Errorf("error while pretty printing HCL: %s", err)
	}

	logger.Log("msg", "formatting HCL", "hcl", buff.String())

	formattedHCL := FormatHCL(buff.Bytes())
	logger.Log("msg", "formatted HCL", "hcl", formattedHCL)

	buff = bytes.NewBuffer(formattedHCL)

	err = fmtcmd.Run(nil, nil, buff, hclw.w, fmtcmd.Options{})
	if err != nil {
		return fmt.Errorf("error while fmt HCL: %s", err)
	}
	return nil
}
