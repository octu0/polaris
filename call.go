package polaris

import (
	"context"
	"io"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pkg/errors"
)

func handleToolCall(t Tool) func(map[string]any) map[string]any {
	return func(req map[string]any) map[string]any {
		j := make(jsonMap, len(req))
		for k, v := range req {
			j.Set(k, v)
		}

		resp, err := t.Handler(&ReqCtx{j, t.Parameters})
		if err != nil {
			if t.ErrorHandler != nil {
				t.ErrorHandler(err)
			}
			return map[string]any{
				"_error": err.Error(),
			}
		}
		return resp.ToMap()
	}
}

func handleMCPToolCall(ctx context.Context, client *client.Client, t Tool) func(map[string]any) map[string]any {
	return func(req map[string]any) map[string]any {
		r := mcp.CallToolRequest{}
		r.Params.Name = t.Name
		r.Params.Arguments = req
		res, err := client.CallTool(ctx, r)
		if err != nil {
			if t.ErrorHandler != nil {
				t.ErrorHandler(err)
			}
			return map[string]any{
				"_error": err.Error(),
			}
		}

		texts := make([]string, len(res.Content))
		for i, c := range res.Content {
			if tc, ok := c.(mcp.TextContent); ok {
				texts[i] = tc.Text
			}
		}
		resp := Resp{"results": texts}
		return resp.ToMap()
	}
}

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

	d.logger.Debugf("callFunction: %s args=%v", name, args)
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
	if err, ok := resp["_error"]; ok {
		d.logger.Warnf("error in %s err:%s", name, err)
	}
	return resp, nil
}
