package polaris

import (
	"context"
	"iter"
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/genai"
)

type Session interface {
	SendText(...string) (iter.Seq2[string, error], error)
	JSONOutput() bool
}

func createSession(ctx context.Context, tc toolConn, rc remoteCall, options ...UseOptionFunc) (Session, error) {
	opt := &UseOption{
		Model:           "gemini-2.5-pro-preview-03-25",
		UseLocalTool:    false,
		Temperature:     0.2,
		TopP:            0.95,
		MaxOutputTokens: 8192,
	}
	for _, f := range options {
		f(opt)
	}
	logger := opt.Logger
	if logger == nil {
		logger = &stdLogger{
			log.New(os.Stdout, "polaris ", log.LstdFlags),
			opt.DebugMode,
		}
	}
	rc.setLogger(logger)

	if opt.DefaultArgsFunc != nil {
		rc.setDefaultArgsFunc(opt.DefaultArgsFunc)
	}

	remoteTools, err := tc.listTools(opt.UseLocalTool)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	functionDeclarations := make([]*genai.FunctionDeclaration, len(remoteTools))
	functionNames := make([]string, len(remoteTools))
	for i, rt := range remoteTools {
		if _, ok := rt.Parameters.Properties["_error"]; ok != true {
			rt.Parameters.Properties["_error"] = &genai.Schema{
				Type:        genai.TypeString,
				Description: "Error details for failed function call",
				Nullable:    genai.Ptr[bool](true),
			}
		}
		functionDeclarations[i] = &genai.FunctionDeclaration{
			Name:        rt.Name,
			Description: rt.Description,
			Parameters:  rt.Parameters,
			Response:    rt.Response,
		}
		functionNames[i] = rt.Name
	}

	client, err := geminiClient(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	config := &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(opt.Temperature),
		TopP:            genai.Ptr(opt.TopP),
		MaxOutputTokens: opt.MaxOutputTokens,
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: false,
		},
	}

	if opt.JSONOutput {
		config.ResponseMIMEType = "application/json"
		if opt.OutputSchema != nil {
			config.ResponseSchema = opt.OutputSchema.Schema().ToGenAI()
		}
	}
	if 0 < len(opt.SystemInstructions) {
		config.SystemInstruction = genai.NewContentFromParts(opt.SystemInstructions, genai.RoleUser)
	}

	// JSONOutput && Tools = does not support
	if 0 < len(functionDeclarations) && opt.JSONOutput != true {
		config.Tools = []*genai.Tool{{
			FunctionDeclarations: functionDeclarations,
		}}
		config.ToolConfig = &genai.ToolConfig{
			FunctionCallingConfig: &genai.FunctionCallingConfig{
				Mode: genai.FunctionCallingConfigModeAuto,
				//AllowedFunctionNames: functionNames,
			},
		}
	}

	if 0 < opt.ThinkingBudget {
		config.ThinkingConfig.IncludeThoughts = true
		config.ThinkingConfig.ThinkingBudget = genai.Ptr(opt.ThinkingBudget)
	}

	chat, err := client.Chats.Create(ctx, opt.Model, config, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &LiveSession{ctx, opt, logger, rc, client, chat}, nil
}

type toolConn interface {
	listTools(bool) ([]genai.FunctionDeclaration, error)
}

var (
	_ toolConn = (*noToolConn)(nil)
)

type noToolConn struct{}

func (*noToolConn) listTools(bool) ([]genai.FunctionDeclaration, error) {
	return nil, nil
}

type funcallCtx struct {
	index int
	name  string
	resp  map[string]any
	err   error
}

type LiveSession struct {
	ctx     context.Context
	opt     *UseOption
	logger  Logger
	rc      remoteCall
	client  *genai.Client
	session *genai.Chat
}

func (s *LiveSession) JSONOutput() bool {
	return s.opt.JSONOutput
}

func (s *LiveSession) SendText(values ...string) (iter.Seq2[string, error], error) {
	texts := make([]*genai.Part, len(values))
	for i, v := range values {
		texts[i] = genai.NewPartFromText(v)
	}
	resp, err := s.session.Send(s.ctx, texts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return s.handleMsg(resp), nil
}

func (s *LiveSession) handleMsg(genContentResp *genai.GenerateContentResponse) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		generate := func(resp *genai.GenerateContentResponse) (*genai.GenerateContentResponse, error) {
			funcalls := resp.FunctionCalls()
			wg := new(sync.WaitGroup)
			ret := make(chan funcallCtx, len(funcalls))

			for i, fc := range funcalls {
				wg.Add(1)
				go func(i int, funcall *genai.FunctionCall) {
					defer wg.Done()

					r, err := s.rc.callFunction(funcall.Name, funcall.Args)
					if err != nil {
						err = errors.Wrapf(err, "name=%s, args=%v", funcall.Name, funcall.Args)
					}
					ret <- funcallCtx{
						i,
						funcall.Name,
						r,
						err,
					}
				}(i, fc)
			}
			wg.Wait()
			close(ret)

			for _, p := range resp.Candidates[0].Content.Parts {
				if p.Text != "" {
					if yield(p.Text, nil) != true {
						return endContent(), nil
					}
				}
			}

			funcResults := make([]*genai.Part, len(funcalls))
			for r := range ret {
				if r.err != nil {
					yield("", errors.WithStack(r.err))
					return nil, errors.WithStack(r.err)
				}
				funcResults[r.index] = genai.NewPartFromFunctionResponse(r.name, r.resp)
			}
			if len(funcResults) < 1 {
				return endContent(), nil
			}

			resp2, err := s.session.Send(s.ctx, funcResults...)
			if err != nil {
				err = errors.WithStack(err)
				yield("", err)
				return nil, err
			}
			return resp2, nil
		}

		resp, err := generate(genContentResp)
		if err != nil {
			s.logger.Warnf("%+v", err)
			return
		}
		for {
			s.logger.Debugf("finish reasion: %s", resp.Candidates[0].FinishReason)
			if resp.Candidates[0].FinishReason == genai.FinishReasonMalformedFunctionCall {
				err := errors.Errorf("malformed function call: %s", resp.Candidates[0].FinishMessage)
				s.logger.Warnf("%+v", err)
				yield("", err)
				return
			}

			if resp.UsageMetadata == nil {
				return
			}
			r, err := generate(resp)
			if err != nil {
				s.logger.Warnf("%+v", err)
				return
			}
			resp = r
		}
	}
}

func endContent() *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{
			FinishReason: genai.FinishReasonStop,
		}},
	}
}
