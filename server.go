package blaze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
)

type Config struct {
	Host   string
	Port   int
	Logger *log.Logger
}

type ServerCtx[T any] struct {
	Data T
}

type BlazeServer[T any] struct {
	srv        http.Server
	hostport   string
	router     *Router[T]
	middleware Middleware[T]
	context    ServerCtx[T]
	logger     *log.Logger
}

type DefaultContext struct{}

func DefaultRouter[T any]() *Router[T] {
	router := NewRouter[T]()
	router.Get("/", ResolveFunc[T](func(req Request[T], res *Response) {
		res.WriteText(http.StatusNotFound, "")
	}))
	return router
}

func DefaultMiddleware[T any](h Handler[T]) Handler[T] {
	return ResolveFunc[T](func(req Request[T], res *Response) {
		h.Handle(req, res)
	})
}

func (b *BlazeServer[T]) Serve() {
	listener, err := net.Listen("tcp", b.hostport)
	if err != nil {
		b.logger.Fatalln(err.Error())
	}
	b.logger.Printf("Server started (Powered by Blaze)")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	b.logger.Printf("Middlewares are setting up.")
	b.srv.Handler = b.setMiddleware(b.middleware, b.router, b.logger)
	b.logger.Printf("Server listening on: http://%s/", b.hostport)

	go b.srv.Serve(listener)
	<-sig

	b.logger.Printf("Interrupt signal received, closing the server.")
	err = b.srv.Close()
	if err != nil {
		b.logger.Fatalln(err.Error())
	}
}

func (b *BlazeServer[T]) UseMiddleware(middleware Middleware[T]) {
	b.middleware = middleware
}

func (b *BlazeServer[T]) UseRouter(router *Router[T]) {
	b.router = router
}

func (b *BlazeServer[T]) UseContext(ctx ServerCtx[T]) {
	b.context = ctx
}

func (b *BlazeServer[T]) setMiddleware(middleware Middleware[T], router *Router[T], args ...any) http.Handler {
	if middleware == nil {
		middleware = DefaultMiddleware
	}
	if router == nil {
		router = DefaultRouter[T]()
	}
	return NativeHandler(b.context, middleware, router, args[0].(*log.Logger))
}

func DefaultBlazeServer[T any]() *BlazeServer[T] {
	defaultConfig := Config{
		Host:   "localhost",
		Port:   8080,
		Logger: log.New(os.Stdout, "[Blaze-DBG] ", log.LstdFlags|log.Lmsgprefix),
	}

	// DefaultRouter.Get("/", ResolveFunc[T](func(req Request[T], res *Response) {
	// 	res.WriteText(http.StatusNotFound, "")
	// }))

	return NewBlazeServer[T](defaultConfig)
}

func NewBlazeServer[T any](config Config) *BlazeServer[T] {
	return &BlazeServer[T]{
		srv: http.Server{
			DisableGeneralOptionsHandler: true,
			ErrorLog:                     config.Logger,
		},
		hostport: fmt.Sprintf("%s:%d", config.Host, config.Port),
		logger:   config.Logger,
	}
}

type Response struct {
	Header     http.Header
	StatusCode int
	Body       bytes.Buffer
	Logger     *log.Logger
}

func (r *Response) Write(bytes []byte) {
	_, err := r.Body.Write(bytes)
	if err != nil {
		r.Logger.Println(err.Error())
	}
}

func (r *Response) SetStatus(statusCode int) {
	r.StatusCode = statusCode
}

func (r *Response) WriteText(statusCode int, text string) {
	r.Header.Set("Content-Type", "text/plain")
	r.SetStatus(statusCode)
	r.Write([]byte(text))
}

func (r *Response) WriteJson(statusCode int, object any) {
	r.Header.Set("Content-Type", "application/json")
	jsonres, err := json.Marshal(object)
	if err != nil {
		r.SetStatus(http.StatusInternalServerError)
		r.Logger.Println(err.Error())
	}
	r.SetStatus(statusCode)
	r.Write(jsonres)
}

type Request[T any] struct {
	URL         *url.URL
	Header      http.Header
	Method      string
	Body        io.ReadCloser
	Params      map[string]string
	Context     ServerCtx[T]
	rootPattern string
}

func (r *Request[T]) GetParam(key string, defaultVal string) string {
	if value, exist := r.Params[key]; exist {
		return value
	} else {
		return defaultVal
	}
}
