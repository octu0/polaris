package polaris

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"google.golang.org/genai"
)

type ConnectOptionFunc func(*ConnectOption)

type ConnectOption struct {
	NatsURL        []string
	Name           string
	Host           string
	Port           string
	UseTLS         bool
	AuthUser       string
	AuthPassword   string
	NoRandomize    bool
	NoEcho         bool
	Timeout        time.Duration
	AllowReconnect bool
	MaxReconnects  int
	ReconnectWait  time.Duration
	ReqTimeout     time.Duration
}

func NatsURL(url ...string) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.NatsURL = url
	}
}

func Name(name string) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.Name = name
	}
}

func ConnectAddress(host, port string) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.Host = host
		o.Port = port
	}
}

func ConnectTLS(useTLS bool) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.UseTLS = useTLS
	}
}

func ConnectAuth(user, password string) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.AuthUser = user
		o.AuthPassword = password
	}
}

func ConnectNoRandomize(noRandomize bool) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.NoRandomize = noRandomize
	}
}

func ConnectNoEcho(noEcho bool) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.NoEcho = noEcho
	}
}

func ConnectTimeout(timeout time.Duration) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.Timeout = timeout
	}
}

func AllowReconnect(allowReconnect bool) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.AllowReconnect = allowReconnect
	}
}

func MaxReconnects(maxReconnects int) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.MaxReconnects = maxReconnects
	}
}

func ReconnectWait(reconnectWait time.Duration) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.ReconnectWait = reconnectWait
	}
}

func RequestTimeout(timeout time.Duration) ConnectOptionFunc {
	return func(o *ConnectOption) {
		o.ReqTimeout = timeout
	}
}

type UseOptionFunc func(*UseOption)

type UseOption struct {
	Model              string
	UseLocalTool       bool
	SystemInstructions []*genai.Part
	History            []*genai.Content
	Temperature        float32
	TopP               float32
	MaxOutputTokens    int32
	JSONOutput         bool
	OutputSchema       TypeDef
	ThinkingConfig     *genai.ThinkingConfig
	Logger             Logger
	DebugMode          bool
	DefaultArgsFunc    func() map[string]any
}

func UseModel(name string) UseOptionFunc {
	return func(o *UseOption) {
		o.Model = name
	}
}

func UseLocalTool(enable bool) UseOptionFunc {
	return func(o *UseOption) {
		o.UseLocalTool = enable
	}
}

type SystemInstructionOptfion func() []*genai.Part

func AddTextSystemInstruction(values ...string) func() []*genai.Part {
	return func() []*genai.Part {
		parts := make([]*genai.Part, len(values))
		for i, v := range values {
			parts[i] = genai.NewPartFromText(v)
		}
		return parts
	}
}

func AddBinarySystemInstruction(data []byte, mimeType string) func() []*genai.Part {
	return func() []*genai.Part {
		return []*genai.Part{
			genai.NewPartFromBytes(data, mimeType),
		}
	}
}

func UseSystemInstruction(sysInstructionOptions ...SystemInstructionOptfion) UseOptionFunc {
	return func(o *UseOption) {
		parts := make([]*genai.Part, 0, len(sysInstructionOptions))
		for _, f := range sysInstructionOptions {
			parts = append(parts, f()...)
		}
		o.SystemInstructions = parts
	}
}

type HistoryOption func() []*genai.Content

func AddUserHistory(values ...string) HistoryOption {
	return func() []*genai.Content {
		contents := make([]*genai.Content, len(values))
		for i, v := range values {
			contents[i] = genai.Text(v)[0]
		}
		return contents
	}
}

func AddModelHistory(values ...string) HistoryOption {
	return func() []*genai.Content {
		contents := make([]*genai.Content, len(values))
		for i, v := range values {
			contents[i] = genai.NewContentFromText(v, genai.RoleModel)
		}
		return contents
	}
}

func UseHistory(historyOptions ...HistoryOption) UseOptionFunc {
	return func(o *UseOption) {
		contents := make([]*genai.Content, 0, len(historyOptions))
		for _, f := range historyOptions {
			contents = append(contents, f()...)
		}
		o.History = contents
	}
}

func UseTemperature(v float32) UseOptionFunc {
	return func(o *UseOption) {
		o.Temperature = v
	}
}

func UseTopP(v float32) UseOptionFunc {
	return func(o *UseOption) {
		o.TopP = v
	}
}

func UseJSONOutput(schema TypeDef) UseOptionFunc {
	return func(o *UseOption) {
		o.JSONOutput = true
		o.OutputSchema = schema
	}
}

func UseToolJSONOutput(conn *Conn, toolName string) UseOptionFunc {
	return func(o *UseOption) {
		if t, ok := conn.Tool(toolName); ok {
			o.JSONOutput = true
			o.OutputSchema = t.Response
		} else {
			o.JSONOutput = true
			o.OutputSchema = nil
		}
	}
}

func UseMaxOutputTokens(size int32) UseOptionFunc {
	return func(o *UseOption) {
		o.MaxOutputTokens = size
	}
}

func UseThinkingConfig(includeThoughts bool, budgetSize int32) UseOptionFunc {
	return func(o *UseOption) {
		o.ThinkingConfig = &genai.ThinkingConfig{
			IncludeThoughts: includeThoughts,
			ThinkingBudget:  genai.Ptr(budgetSize),
		}
	}
}

func UseLogger(lg Logger) UseOptionFunc {
	return func(o *UseOption) {
		o.Logger = lg
	}
}

func UseDebugMode(enable bool) UseOptionFunc {
	return func(o *UseOption) {
		o.DebugMode = enable
	}
}

func UseDefaultArgs(fn func() map[string]any) UseOptionFunc {
	return func(o *UseOption) {
		o.DefaultArgsFunc = fn
	}
}

func Connect(options ...ConnectOptionFunc) (*Conn, error) {
	opt := &ConnectOption{
		AllowReconnect: true,
		MaxReconnects:  -1,
		ReconnectWait:  1 * time.Second,
		NoRandomize:    true,
		NoEcho:         true,
		ReqTimeout:     5 * time.Second,
	}
	for _, f := range options {
		f(opt)
	}

	name := opt.Name
	if name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		name = hostname
	}

	url := func() []string {
		if 0 < len(opt.NatsURL) {
			return opt.NatsURL
		}
		return []string{fmt.Sprintf("nats://%s:%s", opt.Host, opt.Port)}
	}()
	natsOpt := nats.GetDefaultOptions()
	natsOpt.Name = name
	natsOpt.AllowReconnect = opt.AllowReconnect
	natsOpt.MaxReconnect = opt.MaxReconnects
	natsOpt.NoRandomize = opt.NoRandomize
	natsOpt.NoEcho = opt.NoEcho
	natsOpt.Timeout = opt.Timeout
	natsOpt.ReconnectWait = opt.ReconnectWait
	natsOpt.Servers = url

	nc, err := natsOpt.Connect()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return newConn(natsOpt, opt, nc), nil
}

func Generate(ctx context.Context, options ...UseOptionFunc) (Session, error) {
	tc := &noToolConn{}
	rc := &panicRemoteCall{}
	return createSession(ctx, tc, rc, options...)
}

type GenerateJSONFunc func(...string) (Resp, error)

func GenerateJSON(ctx context.Context, options ...UseOptionFunc) (GenerateJSONFunc, error) {
	tc := &noToolConn{}
	rc := &panicRemoteCall{}
	s, err := createSession(ctx, tc, rc, options...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if s.JSONOutput() != true {
		return nil, errors.Errorf("require JSONOutput=true")
	}
	return func(text ...string) (Resp, error) {
		it, err := s.SendText(text...)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for ret, err := range it {
			if err != nil {
				return nil, errors.WithStack(err)
			}

			resp := Resp{}
			if err := json.Unmarshal([]byte(ret), &resp); err != nil {
				return nil, errors.WithStack(err)
			}
			return resp, nil
		}
		return nil, nil
	}, nil
}

type Conn struct {
	ctx     context.Context
	cancel  context.CancelFunc
	natsOpt nats.Options
	opt     *ConnectOption
	nc      *nats.Conn
	subs    []*nats.Subscription
	tools   []Tool
	logger  Logger
}

func (c *Conn) NewConnection() (*Conn, error) {
	nc, err := c.natsOpt.Connect()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return newConn(c.natsOpt, c.opt, nc), nil
}

func (c *Conn) Close() {
	c.cancel()
	for _, sub := range c.subs {
		sub.Unsubscribe()
	}
	c.UnregisterTools()
	c.subs = nil
	c.nc.Close()
}

func (c *Conn) UnregisterTools() error {
	if len(c.tools) < 1 {
		return nil
	}
	list := make([]genai.FunctionDeclaration, len(c.tools))
	for i, t := range c.tools {
		list[i] = t.FunctionDeclaration()
	}
	for _, dec := range list {
		resp, err := requestWithData(
			c,
			TopicUnregisterTool,
			JSONEncoder[genai.FunctionDeclaration](),
			JSONEncoder[RespError](),
			dec,
		)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := resp.Err(); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (c *Conn) RegisterTool(t Tool) error {
	if err := subscribeReqResp(
		c,
		tooltopic(t.Name),
		JSONEncoder[map[string]any](),
		JSONEncoder[map[string]any](),
		handleToolCall(t),
	); err != nil {
		return errors.WithStack(err)
	}

	resp, err := requestWithData(
		c,
		TopicRegisterTool,
		JSONEncoder[genai.FunctionDeclaration](),
		JSONEncoder[RespError](),
		t.FunctionDeclaration(),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := resp.Err(); err != nil {
		return errors.WithStack(err)
	}
	c.tools = append(c.tools, t)
	return nil
}

func (c *Conn) Tool(name string) (Tool, bool) {
	for _, t := range c.tools {
		if t.Name == name {
			return t, true
		}
	}
	return Tool{}, false
}

func (c *Conn) listTools(useLocalTool bool) ([]genai.FunctionDeclaration, error) {
	remoteList, err := request(
		c,
		TopicListTool,
		JSONEncoder[[]genai.FunctionDeclaration](),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	declares := make([]genai.FunctionDeclaration, 0, len(remoteList))
	for _, d := range remoteList {
		if _, ok := c.Tool(d.Name); ok {
			if useLocalTool != true {
				continue
			}
		}
		declares = append(declares, d)
	}
	return declares, nil
}

func (c *Conn) Use(ctx context.Context, options ...UseOptionFunc) (Session, error) {
	rc := &defaultRemoteCall{c, nil, nil}
	return createSession(ctx, c, rc, options...)
}

func (c *Conn) Call(ctx context.Context, name string, req Req) (Resp, error) {
	rc := &defaultRemoteCall{c, nil, nil}
	ret, err := rc.callFunction(name, req.ToMap())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return Resp(ret), nil
}

func (c *Conn) toolKeepAlive() {
	if len(c.tools) < 1 {
		return
	}

	list := make([]genai.FunctionDeclaration, len(c.tools))
	for i, t := range c.tools {
		list[i] = t.FunctionDeclaration()
	}
	resp, err := requestWithData(
		c,
		TopicToolKeepalive,
		JSONEncoder[[]genai.FunctionDeclaration](),
		JSONEncoder[RespError](),
		list,
	)
	if err != nil {
		log.Printf("WARN: keepalive: %+v", err)
		return
	}
	if err := resp.Err(); err != nil {
		log.Printf("WARN: keepalive response: %+v", err)
		return
	}
}

func (c *Conn) toolKeepAliveLoop(ctx context.Context) {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-tick.C:
			c.toolKeepAlive()
		}
	}
}

func newConn(natsOpt nats.Options, opt *ConnectOption, nc *nats.Conn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		ctx:     ctx,
		cancel:  cancel,
		natsOpt: natsOpt,
		opt:     opt,
		nc:      nc,
		subs:    make([]*nats.Subscription, 0),
		tools:   make([]Tool, 0),
		logger: &stdLogger{
			log.New(os.Stdout, "polaris ", log.LstdFlags),
			false,
		},
	}
	go c.toolKeepAliveLoop(ctx)
	return c
}

func handleToolCall(t Tool) func(map[string]any) map[string]any {
	return func(req map[string]any) map[string]any {
		j := make(JSONMap, len(req))
		for k, v := range req {
			j.Set(k, v)
		}
		resp := Resp{}
		ctx := Ctx{j, t.Parameters, &resp}
		if err := t.Handler(&ctx); err != nil {
			return map[string]any{
				"_error": err.Error(),
			}
		}
		return resp.ToMap()
	}
}

func tooltopic(name string) string {
	return fmt.Sprintf("polaris:user-func:%s", name)
}

func request[Resp any](c *Conn, topic string, encResp Encoder[Resp]) (Resp, error) {
	var resp Resp
	msg, err := c.nc.Request(topic, []byte{}, c.opt.ReqTimeout)
	if err != nil {
		return resp, errors.WithStack(err)
	}
	c.nc.Flush()
	rr, err := encResp.Decode(msg.Data)
	if err != nil {
		return resp, errors.WithStack(err)
	}
	return rr, nil
}

func requestWithData[Req any, Resp any](c *Conn, topic string, encReq Encoder[Req], encResp Encoder[Resp], req Req) (Resp, error) {
	var resp Resp
	data, err := encReq.Encode(req)
	if err != nil {
		return resp, errors.WithStack(err)
	}

	msg, err := c.nc.Request(topic, data, c.opt.ReqTimeout)
	if err != nil {
		return resp, errors.WithStack(err)
	}
	c.nc.Flush()
	rr, err := encResp.Decode(msg.Data)
	if err != nil {
		return resp, errors.WithStack(err)
	}
	return rr, nil
}

type respHandler[Resp any] func() Resp
type reqrespHandler[Req any, Resp any] func(Req) Resp

func subscribeResp[Resp any](c *Conn, topic string, encResp Encoder[Resp], handler respHandler[Resp]) error {
	sub, err := c.nc.Subscribe(topic, func(msg *nats.Msg) {
		resp := handler()
		data, err := encResp.Encode(resp)
		if err != nil {
			log.Printf("WARN: resp: %+v", errors.WithStack(err))
		}
		if err := msg.Respond(data); err != nil {
			log.Printf("WARN: respond: %+v", errors.WithStack(err))
		}
	})
	if err != nil {
		return errors.WithStack(err)
	}
	c.nc.Flush()
	c.subs = append(c.subs, sub)
	return nil
}

func subscribeReqResp[Req any, Resp any](c *Conn, topic string, encReq Encoder[Req], encResp Encoder[Resp], handler reqrespHandler[Req, Resp]) error {
	sub, err := c.nc.Subscribe(topic, func(msg *nats.Msg) {
		req, err := encReq.Decode(msg.Data)
		if err != nil {
			log.Printf("WARN: req %+v", errors.WithStack(err))
		}
		resp := handler(req)
		data, err := encResp.Encode(resp)
		if err != nil {
			log.Printf("WARN: resp: %+v", errors.WithStack(err))
		}
		if err := msg.Respond(data); err != nil {
			log.Printf("WARN: respond: %+v", errors.WithStack(err))
		}
	})
	if err != nil {
		return errors.WithStack(err)
	}
	c.nc.Flush()
	c.subs = append(c.subs, sub)
	return nil
}

func geminiClient(ctx context.Context) (*genai.Client, error) {
	// require ENV for
	//  VertexAI mode::
	//   GOOGLE_GENAI_USE_VERTEXAI=1 or GOOGLE_GENAI_USE_VERTEXAI=yes
	//   GOOGLE_CLOUD_PROJECT
	//   GOOGLE_CLOUD_LOCATION
	//   GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_API_KEY
	//  GeminiAPI mode::
	//   GOOGLE_API_KEY
	//

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return client, nil
}
