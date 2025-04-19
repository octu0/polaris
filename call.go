package polaris

import (
	"io"
	"log"

	"github.com/pkg/errors"
)

type remoteCall interface {
	setLogger(Logger)
	setDefaultArgsFunc(func() map[string]any)
	callFunction(string, map[string]any) (map[string]any, error)
}

var (
	_ remoteCall = (*panicRemoteCall)(nil)
	_ remoteCall = (*defaultRemoteCall)(nil)
)

type panicRemoteCall struct{}

func (*panicRemoteCall) setLogger(Logger) {}

func (*panicRemoteCall) setDefaultArgsFunc(func() map[string]any) {}

func (*panicRemoteCall) callFunction(name string, args map[string]any) (map[string]any, error) {
	panic(errors.Errorf("not support callFunction: called func=%s args=%v", name, args))
}

type defaultRemoteCall struct {
	conn            *Conn
	logger          Logger
	defaultArgsFunc func() map[string]any
}

func (d *defaultRemoteCall) setLogger(lg Logger) {
	d.logger = lg
}

func (d *defaultRemoteCall) setDefaultArgsFunc(fn func() map[string]any) {
	d.defaultArgsFunc = fn
}

func (d *defaultRemoteCall) callFunction(name string, args map[string]any) (map[string]any, error) {
	if d.logger == nil {
		d.logger = &stdLogger{log.New(io.Discard, "", 0), false}
	}
	if d.defaultArgsFunc != nil {
		defaultArgs := d.defaultArgsFunc()
		for k, v := range defaultArgs {
			if _, ok := args[k]; ok != true {
				args[k] = v
			}
		}
	}

	d.logger.DebugF("callFunction: %s args=%v", name, args)
	resp, err := requestWithData(
		d.conn,
		tooltopic(name),
		JSONEncoder[map[string]any](),
		JSONEncoder[map[string]any](),
		args,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp, ok := resp["_error"]; ok {
		return nil, errors.Errorf("error: %s", resp)
	}
	delete(resp, "_error")
	return resp, nil
}
