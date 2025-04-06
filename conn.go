package polaris

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

type ConnectOptionFunc func(*ConnectOption)

type ConnectOption struct {
	NatsURL        string
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

func NatsURL(url string) ConnectOptionFunc {
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

type UseOptionFunc func(*UseOption)

type UseOption struct {
	Model              string
	UseLocalTool       bool
	SystemInstructions []genai.Part
	Temperature        float32
	TopP               float32
	MaxOutputTokens    int32
	JSONOutput         bool
	OutputSchema       TypeDef
	Logger             Logger
	DebugMode          bool
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

type systemInstructionOptfion func() []genai.Part

func AddTextSystemInstruction(values ...string) func() []genai.Part {
	return func() []genai.Part {
		parts := make([]genai.Part, len(values))
		for i, v := range values {
			parts[i] = genai.Text(v)
		}
		return parts
	}
}

func AddBinarySystemInstruction(data []byte, mimeType string) func() []genai.Part {
	return func() []genai.Part {
		return []genai.Part{
			&genai.Blob{MIMEType: mimeType, Data: data},
		}
	}
}

func UseSystemInstruction(sysInstructionOptions ...systemInstructionOptfion) UseOptionFunc {
	return func(o *UseOption) {
		parts := make([]genai.Part, 0, len(sysInstructionOptions))
		for _, f := range sysInstructionOptions {
			parts = append(parts, f()...)
		}
		o.SystemInstructions = parts
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

func UseMaxOutputTokens(size int32) UseOptionFunc {
	return func(o *UseOption) {
		o.MaxOutputTokens = size
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

func Connect(options ...ConnectOptionFunc) (*Conn, error) {
	opt := &ConnectOption{
		AllowReconnect: true,
		MaxReconnects:  -1,
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

	url := func() string {
		if opt.NatsURL != "" {
			return opt.NatsURL
		}
		return fmt.Sprintf("nats://%s:%s", opt.Host, opt.Port)
	}()
	natsOpt := nats.GetDefaultOptions()
	natsOpt.Name = name
	natsOpt.AllowReconnect = opt.AllowReconnect
	natsOpt.MaxReconnect = opt.MaxReconnects
	natsOpt.NoRandomize = opt.NoRandomize
	natsOpt.NoEcho = opt.NoEcho
	natsOpt.Timeout = opt.Timeout
	natsOpt.ReconnectWait = opt.ReconnectWait
	natsOpt.Url = url

	nc, err := natsOpt.Connect()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return newConn(opt, nc), nil
}

func Use(ctx context.Context, options ...UseOptionFunc) (Session, error) {
	tc := &noToolConn{}
	rc := &panicRemoteCall{}
	return createSession(ctx, tc, rc, options...)
}

type Conn struct {
	ctx    context.Context
	cancel context.CancelFunc
	opt    *ConnectOption
	nc     *nats.Conn
	subs   []*nats.Subscription
	tools  []Tool
	logger Logger
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
			GobEncoder[genai.FunctionDeclaration](),
			GobEncoder[RespError](),
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
	subscribeReqResp(
		c,
		tooltopic(t.Name),
		JSONEncoder[map[string]any](),
		JSONEncoder[map[string]any](),
		handleToolCall(t),
	)
	resp, err := requestWithData(
		c,
		TopicRegisterTool,
		GobEncoder[genai.FunctionDeclaration](),
		GobEncoder[RespError](),
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

func (c *Conn) listTools(useLocalTool bool) ([]genai.FunctionDeclaration, error) {
	list, err := request(
		c,
		TopicListTool,
		GobEncoder[[]genai.FunctionDeclaration](),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	declares := make([]genai.FunctionDeclaration, 0, len(list))
	for _, d := range list {
		if c.hasTool(d) {
			if useLocalTool != true {
				continue
			}
		}
		declares = append(declares, d)
	}
	return declares, nil
}

func (c *Conn) Use(ctx context.Context, options ...UseOptionFunc) (Session, error) {
	return createSession(ctx, c, c, options...)
}

func (c *Conn) hasTool(d genai.FunctionDeclaration) bool {
	for _, t := range c.tools {
		if t.Name == d.Name {
			return true
		}
	}
	return false
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
		GobEncoder[[]genai.FunctionDeclaration](),
		GobEncoder[RespError](),
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

func (c *Conn) callFunction(name string, args map[string]any) (map[string]any, error) {
	c.logger.DebugF("callFunction: %s args=%v", name, args)
	resp, err := requestWithData(
		c,
		tooltopic(name),
		JSONEncoder[map[string]any](),
		JSONEncoder[map[string]any](),
		args,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
}

func newConn(opt *ConnectOption, nc *nats.Conn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		ctx:    ctx,
		cancel: cancel,
		opt:    opt,
		nc:     nc,
		subs:   make([]*nats.Subscription, 0),
		tools:  make([]Tool, 0),
		logger: &stdLogger{
			log.New(os.Stdout, " polaris", log.LstdFlags),
			false,
		},
	}
	go c.toolKeepAliveLoop(ctx)
	return c
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
			log.New(os.Stdout, " polaris", log.LstdFlags),
			opt.DebugMode,
		}
	}

	remoteTools, err := tc.listTools(opt.UseLocalTool)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	functionDeclarations := make([]*genai.FunctionDeclaration, len(remoteTools))
	functionNames := make([]string, len(remoteTools))
	for i, rt := range remoteTools {
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
	model := client.GenerativeModel(opt.Model)
	model.Temperature = genai.Ptr(opt.Temperature)
	model.TopP = genai.Ptr(opt.TopP)
	model.MaxOutputTokens = genai.Ptr(opt.MaxOutputTokens)

	if opt.JSONOutput {
		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = opt.OutputSchema.Schema()
	}
	if 0 < len(opt.SystemInstructions) {
		model.SystemInstruction = &genai.Content{
			Parts: opt.SystemInstructions,
		}
	}
	// JSONOutput && Tools = does not support
	if 0 < len(functionDeclarations) && opt.JSONOutput != true {
		model.Tools = []*genai.Tool{{
			FunctionDeclarations: functionDeclarations,
		}}
		model.ToolConfig = &genai.ToolConfig{
			FunctionCallingConfig: &genai.FunctionCallingConfig{
				Mode: genai.FunctionCallingAuto,
				//AllowedFunctionNames: functionNames,
			},
		}
	}

	return &LiveSession{ctx, logger, rc, client, model.StartChat()}, nil
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
				"error": err.Error(),
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
	//   GOOGLE_CLOUD_PROJECT
	//   GOOGLE_CLOUD_LOCATION
	//   GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_API_KEY
	//

	project, ok := os.LookupEnv("GOOGLE_CLOUD_PROJECT")
	if ok != true {
		return nil, errors.Errorf("require ENV{GOOGLE_CLOUD_PROJECT}")
	}

	loc, ok := os.LookupEnv("GOOGLE_CLOUD_LOCATION")
	if ok != true {
		if region, ok := os.LookupEnv("GOOGLE_CLOUD_REGION"); ok {
			loc = region
		} else {
			loc = "us-central1"
		}
	}

	credentials, hasCredential := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	apiKey, hasAPIKey := os.LookupEnv("GOOGLE_API_KEY")
	clientOption := []option.ClientOption{}
	if hasCredential {
		clientOption = append(clientOption, option.WithCredentialsFile(credentials))
	}
	if hasAPIKey {
		clientOption = append(clientOption, option.WithAPIKey(apiKey))
	}
	client, err := genai.NewClient(ctx, project, loc, clientOption...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return client, nil
}
