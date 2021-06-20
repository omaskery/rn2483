package fake

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/go-logr/logr"
	"go.uber.org/multierr"

	"github.com/omaskery/rn2483"
)

type Device struct {
	logger  logr.Logger
	writer  *io.PipeWriter
	reader  *io.PipeReader
	stopped chan error

	Sys   SysState
	Mac   MacState
	Radio RadioState
}

type commandContext struct {
	logger         logr.Logger
	command        string
	responseWriter *io.PipeWriter
}

func (c *commandContext) WriteResponse(format string, a ...interface{}) error {
	response := fmt.Sprintf(format, a...)
	c.logger.Info("writing response", "response", response)
	if _, err := fmt.Fprintf(c.responseWriter, "%s\r\n", response); err != nil {
		return fmt.Errorf("error writing command response: %w", err)
	}
	return nil
}

func (d *Device) run(commandReader *io.PipeReader, responseWriter *io.PipeWriter) error {
	scanner := bufio.NewScanner(commandReader)

	ctx := commandContext{
		logger:         d.logger.V(1),
		responseWriter: responseWriter,
	}

	for scanner.Scan() {
		ctx.command = scanner.Text()

		if err := d.processCommand(&ctx); err != nil {
			return fmt.Errorf("error processing command '%s': %w", ctx.command, err)
		}
	}

	return nil
}

func (d *Device) processCommand(ctx *commandContext) error {
	ctx.logger.Info("command received", "command", ctx.command)

	tokens := strings.Split(ctx.command, " ")
	if len(tokens) < 1 {
		return invalidParam(ctx)
	}

	switch tokens[0] {
	case "sys":
		return d.processSysCommand(ctx, tokens[1:])
	case "mac":
		return d.processMacCommand(ctx, tokens[1:])
	case "radio":
		return d.processRadioCommand(ctx, tokens[1:])
	default:
		return invalidParam(ctx)
	}
}

type Config struct {
	Logger logr.Logger
}

func New(cfg Config) *Device {
	logger := logr.Discard()
	if cfg.Logger != nil {
		logger = cfg.Logger
	}

	commandReader, commandWriter := io.Pipe()
	responseReader, responseWriter := io.Pipe()

	d := &Device{
		logger:  logger,
		writer:  commandWriter,
		reader:  responseReader,
		stopped: make(chan error),
	}

	d.Sys.ensureDefaults()
	d.Mac.ensureDefaults()
	d.Radio.ensureDefaults()

	go func() {
		err := d.run(commandReader, responseWriter)
		if closeErr := responseWriter.Close(); closeErr != nil {
			err = multierr.Combine(err, closeErr)
		}
		d.stopped <- err
		close(d.stopped)
	}()

	return d
}

func NewFakeDevice(cfg Config) (*Device, *rn2483.Device) {
	fake := New(cfg)
	device := rn2483.New(rn2483.Config{
		Serial: fake,
	})

	return fake, device
}

var _ io.ReadWriteCloser = (*Device)(nil)

func (d *Device) Read(p []byte) (n int, err error) {
	return d.reader.Read(p)
}

func (d *Device) Write(p []byte) (n int, err error) {
	return d.writer.Write(p)
}

func (d *Device) Close() error {
	if err := d.writer.Close(); err != nil {
		return fmt.Errorf("error closing command writer: %w", err)
	}

	return <-d.stopped
}
