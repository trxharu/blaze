package blaze

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

type Middleware[T any] func(h Handler[T]) Handler[T]

func LoggingMiddleware[T any](h Handler[T]) Handler[T] {
	return ResolveFunc[T](func(req Request[T], res *Response) {
		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(req.Body)
		}

		res.Logger.Printf("Request log:\n\tPath: %s\n\tMethod:%s\n\tHeaders: %+v\n\tBody: %s",
			req.URL.String(), req.Method, req.Header.Clone(), string(body))

		req.Body = io.NopCloser(bytes.NewBuffer(body))

		h.Handle(req, res)
		res.Logger.Printf("Response log:\n\tResponseCode: %d\n\tHeaders: %+v\n\tBody: %s",
			res.StatusCode, res.Header.Clone(), res.Body.String())
	})
}

func CorsMiddleware[T any](h Handler[T]) Handler[T] {
	return ResolveFunc[T](func(req Request[T], res *Response) {
		res.Header.Set("Access-Control-Allow-Origin", "*")
		h.Handle(req, res)
	})
}

func ChainMiddleware[T any](middlewares ...Middleware[T]) Middleware[T] {
	return func(finalHandler Handler[T]) Handler[T] {
		for i := len(middlewares) - 1; i >= 0; i-- {
			finalHandler = middlewares[i](finalHandler)
		}
		return finalHandler
	}
}

func NativeHandler[T any](ctx ServerCtx[T], middleware Middleware[T], router *Router[T], logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := nativeRequestAdaptor[T](r, ctx)
		response := &Response{
			Logger: logger,
			Header: make(http.Header),
		}

		handle := middleware(router)
		handle.Handle(request, response)

		for key, values := range response.Header {
			for _, val := range values {
				w.Header().Add(key, val)
			}
		}
		w.WriteHeader(response.StatusCode)
		w.Write(response.Body.Bytes())
	})
}
