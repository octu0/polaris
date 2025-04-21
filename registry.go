package polaris

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"cloud.google.com/go/vertexai/genai"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/pkg/errors"
)

const (
	TopicRegisterTool   string = "polaris:tool:register"
	TopicUnregisterTool string = "polaris:tool:unregister"
	TopicToolKeepalive  string = "polaris:tool:keepalive"
	TopicListTool       string = "polaris:tool:list"
)

type toolDeclareWithDeadline struct {
	Declare  genai.FunctionDeclaration
	Deadline time.Time
}

type (
	RespError struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}
)

func (e RespError) String() string {
	return e.Msg
}

func (e RespError) Err() error {
	if e.Success {
		return nil
	}
	return errors.Errorf(e.Msg)
}

type Registry struct {
	ctx    context.Context
	cancel context.CancelFunc
	mutex  *sync.RWMutex
	ns     *server.Server
	conn   *Conn
	tools  map[string]*toolDeclareWithDeadline
}

func (r *Registry) Close() {
	r.conn.Close()
	r.ns.Shutdown()
}

func (r *Registry) subscribeTool() error {
	if err := subscribeReqResp(
		r.conn,
		TopicRegisterTool,
		GobEncoder[genai.FunctionDeclaration](),
		GobEncoder[RespError](),
		r.handleRegisterTool,
	); err != nil {
		return errors.WithStack(err)
	}

	if err := subscribeReqResp(
		r.conn,
		TopicUnregisterTool,
		GobEncoder[genai.FunctionDeclaration](),
		GobEncoder[RespError](),
		r.handleUnregisterTool,
	); err != nil {
		return errors.WithStack(err)
	}

	if err := subscribeResp(
		r.conn,
		TopicListTool,
		GobEncoder[[]genai.FunctionDeclaration](),
		r.handleListTool,
	); err != nil {
		return errors.WithStack(err)
	}

	if err := subscribeReqResp(
		r.conn,
		TopicToolKeepalive,
		GobEncoder[[]genai.FunctionDeclaration](),
		GobEncoder[RespError](),
		r.handleToolKeepAlive,
	); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Registry) toolGC() {
	now := time.Now()
	r.mutex.RLock()
	deadTools := make([]string, 0, len(r.tools))
	for name, decdeadline := range r.tools {
		if now.Before(decdeadline.Deadline) != true {
			deadTools = append(deadTools, name)
		}
	}
	r.mutex.RUnlock()

	for _, name := range deadTools {
		log.Printf("INFO: expire tool %s", name)
		r.mutex.Lock()
		delete(r.tools, name)
		r.mutex.Unlock()
	}
}

func (r *Registry) toolGCLoop() {
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return

		case <-tick.C:
			r.toolGC()
		}
	}
}

func (r *Registry) handleRegisterTool(declare genai.FunctionDeclaration) RespError {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.tools[declare.Name]; ok {
		return RespError{false, fmt.Sprintf("register: %s already registered", declare.Name)}
	}
	r.tools[declare.Name] = &toolDeclareWithDeadline{
		Declare:  declare,
		Deadline: time.Now().Add(time.Hour),
	}
	log.Printf("INFO: tool %s registered", declare.Name)
	return RespError{true, "OK"}
}

func (r *Registry) handleUnregisterTool(declare genai.FunctionDeclaration) RespError {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.tools[declare.Name]; ok != true {
		return RespError{false, fmt.Sprintf("unregister: %s not found", declare.Name)}
	}
	delete(r.tools, declare.Name)
	log.Printf("INFO: tool %s unregistered", declare.Name)
	return RespError{true, "OK"}
}

func (r *Registry) handleListTool() []genai.FunctionDeclaration {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	list := make([]genai.FunctionDeclaration, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t.Declare)
	}
	return list
}

func (r *Registry) handleToolKeepAlive(list []genai.FunctionDeclaration) RespError {
	for _, d := range list {
		r.mutex.Lock()
		if v, ok := r.tools[d.Name]; ok {
			v.Deadline = time.Now().Add(time.Hour)
		} else {
			r.tools[d.Name] = &toolDeclareWithDeadline{
				Declare:  d,
				Deadline: time.Now().Add(time.Hour),
			}
		}
		r.mutex.Unlock()
	}
	return RespError{true, "OK"}
}

func newRegistry(ns *server.Server, conn *Conn) *Registry {
	ctx, cancel := context.WithCancel(context.Background())
	return &Registry{
		ctx:    ctx,
		cancel: cancel,
		mutex:  new(sync.RWMutex),
		ns:     ns,
		conn:   conn,
		tools:  make(map[string]*toolDeclareWithDeadline, 0),
	}
}

type (
	RegistryOption        func(*server.Options)
	RegistryClusterOption func(*server.ClusterOpts)
)

func WithBind(host string, port int) RegistryOption {
	return func(o *server.Options) {
		o.Host = host
		o.Port = port
	}
}

func WithMaxPayload(size int32) RegistryOption {
	return func(o *server.Options) {
		o.MaxPayload = size
	}
}

func WithRoutes(routesStr string) RegistryOption {
	return func(o *server.Options) {
		o.Routes = server.RoutesFromStr(routesStr)
	}
}

func WithClusterOption(opts ...RegistryClusterOption) RegistryOption {
	return func(o *server.Options) {
		opt := server.ClusterOpts{}
		for _, fn := range opts {
			fn(&opt)
		}
		o.Cluster = opt
	}
}

func WithClusterName(name string) RegistryClusterOption {
	return func(o *server.ClusterOpts) {
		o.Name = name
	}
}

func WithClusterHost(host string) RegistryClusterOption {
	return func(o *server.ClusterOpts) {
		o.Host = host
	}
}

func WithClusterPort(port int) RegistryClusterOption {
	return func(o *server.ClusterOpts) {
		o.Port = port
	}
}

func WithClussterAdvertise(advertise string) RegistryClusterOption {
	return func(o *server.ClusterOpts) {
		o.Advertise = advertise
	}
}

func CreateRegistry(opts ...RegistryOption) (*Registry, error) {
	o := &server.Options{
		Debug:  false,
		NoSigs: true,
		NoLog:  true,
	}
	for _, fn := range opts {
		fn(o)
	}
	if o.Cluster.PoolSize < 1 {
		o.Cluster.PoolSize = -1
	}

	ns := server.New(o)
	ns.DisableJetStream()
	go ns.Start()

	waitRouting := make(chan struct{})
	if 0 < len(o.Routes) {
		go ns.StartRouting(waitRouting)
	}

	if ns.ReadyForConnections(10*time.Second) != true {
		return nil, errors.Errorf("failed to start server")
	}
	close(waitRouting)

	conn, err := Connect(
		NatsURL(ns.ClientURL()),
		Name("registry"),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := newRegistry(ns, conn)
	if err := r.subscribeTool(); err != nil {
		return nil, errors.WithStack(err)
	}
	go r.toolGCLoop()
	return r, nil
}
