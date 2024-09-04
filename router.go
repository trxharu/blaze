package blaze

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Handler[T any] interface {
	Handle(req Request[T], res *Response)
}

type ResolveFunc[T any] func(req Request[T], res *Response)

func (rf ResolveFunc[T]) Handle(req Request[T], res *Response) {
	rf(req, res)
}

type HandlerNode[T any] struct {
	pattern *regexp.Regexp
	method  string
	handler Handler[T]
	next    *HandlerNode[T]
}

type Router[T any] struct {
	root *HandlerNode[T]
}

// For Development and Testing
func (r *Router[T]) Exec(req Request[T], res *Response) {
	r.Handle(req, res)
}

func (r *Router[T]) NativeHandler(ctx ServerCtx[T], args ...any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		request := nativeRequestAdaptor[T](req, ctx)
		response := &Response{
			Logger: args[0].(*log.Logger),
			Header: make(http.Header),
		}
		r.Handle(request, response)
		for key, values := range response.Header {
			for _, val := range values {
				w.Header().Add(key, val)
			}
		}
		w.WriteHeader(response.StatusCode)
		w.Write(response.Body.Bytes())
	})
}

func nativeRequestAdaptor[T any](r *http.Request, ctx ServerCtx[T]) Request[T] {
	return Request[T]{
		Body:    r.Body,
		URL:     r.URL,
		Header:  r.Header,
		Method:  r.Method,
		Context: ctx,
		Params:  make(map[string]string),
	}
}

func (r *Router[T]) Handle(req Request[T], res *Response) {
	regexParam := regexp.MustCompile(`:[\w_]+`)

	if r.root.pattern.MatchString(req.URL.Path) {
		if r.root.handler == nil {
			res.WriteText(http.StatusNotImplemented, "")
			return
		}
		r.root.handler.Handle(req, res)
		return
	}

	head := r.root
	for head != nil {
		urlParmsPattern := regexParam.ReplaceAllString(head.pattern.String(), `([\w-_]+)`)
		regex := regexp.MustCompile(urlParmsPattern)

		if regex.MatchString(req.rootPattern + req.URL.Path) {
			if head.method != req.Method {
				res.WriteText(http.StatusMethodNotAllowed, "")
				return
			}
			req.rootPattern = getStringFromRegex(*head.pattern)
			paramKeys := extractParamKeys(regexParam, head.pattern.String())
			populateParams(req.Params, paramKeys, req.URL.Path, regex)
			head.handler.Handle(req, res)
			return
		}

		head = head.next
	}
	res.WriteText(http.StatusNotFound, "")
	res.Logger.Println("No routes found. Request reached end of the chain.")
}

func (r *Router[T]) Get(pattern string, handler Handler[T]) {
	if r.root.pattern.MatchString(pattern) {
		if r.root.handler == nil {
			r.root.handler = handler
			return
		}
	}

	urlPattern := fmt.Sprintf(`%s/?$`, pattern)
	regex := regexp.MustCompile(urlPattern)

	head := r.root.next
	node := &HandlerNode[T]{
		pattern: regex,
		method:  "GET",
		handler: handler,
		next:    nil,
	}

	if head == nil {
		r.root.next = node
	} else {
		for head.next != nil {
			head = head.next
		}
		head.next = node
	}
}

func (r *Router[T]) Post(pattern string, handler Handler[T]) {
	if r.root.pattern.MatchString(pattern) {
		if r.root.handler == nil {
			r.root.handler = handler
			return
		}
	}

	urlPattern := fmt.Sprintf(`%s/?$`, pattern)
	regex := regexp.MustCompile(urlPattern)

	head := r.root.next
	node := &HandlerNode[T]{
		pattern: regex,
		method:  "POST",
		handler: handler,
		next:    nil,
	}

	if head == nil {
		r.root.next = node
	} else {
		for head.next != nil {
			head = head.next
		}
		head.next = node
	}
}

func (r *Router[T]) Patch(pattern string, handler Handler[T]) {
	if r.root.pattern.MatchString(pattern) {
		if r.root.handler == nil {
			r.root.handler = handler
			return
		}
	}

	urlPattern := fmt.Sprintf(`%s/?$`, pattern)
	regex := regexp.MustCompile(urlPattern)

	head := r.root.next
	node := &HandlerNode[T]{
		pattern: regex,
		method:  "PATCH",
		handler: handler,
		next:    nil,
	}

	if head == nil {
		r.root.next = node
	} else {
		for head.next != nil {
			head = head.next
		}
		head.next = node
	}
}

func (r *Router[T]) Delete(pattern string, handler Handler[T]) {
	if r.root.pattern.MatchString(pattern) {
		if r.root.handler == nil {
			r.root.handler = handler
			return
		}
	}

	urlPattern := fmt.Sprintf(`%s/?$`, pattern)
	regex := regexp.MustCompile(urlPattern)

	head := r.root.next
	node := &HandlerNode[T]{
		pattern: regex,
		method:  "DELETE",
		handler: handler,
		next:    nil,
	}

	if head == nil {
		r.root.next = node
	} else {
		for head.next != nil {
			head = head.next
		}
		head.next = node
	}
}

func (r *Router[T]) Put(pattern string, handler Handler[T]) {
	if r.root.pattern.MatchString(pattern) {
		if r.root.handler == nil {
			r.root.handler = handler
			return
		}
	}

	urlPattern := fmt.Sprintf(`%s/?$`, pattern)
	regex := regexp.MustCompile(urlPattern)

	head := r.root.next
	node := &HandlerNode[T]{
		pattern: regex,
		method:  "PUT",
		handler: handler,
		next:    nil,
	}

	if head == nil {
		r.root.next = node
	} else {
		for head.next != nil {
			head = head.next
		}
		head.next = node
	}
}

func (r *Router[T]) SubRoute(pattern string, handler Handler[T]) {
	regex := regexp.MustCompile(fmt.Sprintf(`^%s`, pattern))

	node := &HandlerNode[T]{
		pattern: regex,
		method:  "ALL",
		handler: handler,
		next:    nil,
	}

	head := r.root.next
	if head == nil {
		r.root.next = node
	} else {
		for head.next != nil {
			head = head.next
		}
		head.next = node
	}
}

func NewRouter[T any]() *Router[T] {
	router := &Router[T]{}
	router.root = &HandlerNode[T]{
		pattern: regexp.MustCompile(`^/$`),
		handler: nil,
		next:    nil,
	}
	return router
}

func getStringFromRegex(regex regexp.Regexp) string {
	return strings.TrimFunc(regex.String(), func(r rune) bool {
		return r == '^' || r == '$'
	})
}

func extractParamKeys(regex *regexp.Regexp, pattern string) []string {
	keys := make([]string, 0)
	if matches := regex.FindAllString(pattern, -1); matches != nil {
		keys = append(keys, matches...)
	}
	for index, key := range keys {
		keys[index] = strings.TrimPrefix(key, ":")
	}
	return keys
}

func populateParams(params map[string]string, keys []string, url string, regex *regexp.Regexp) {
	if matches := regex.FindAllStringSubmatch(url, -1); matches != nil {
		for _, match := range matches {
			for i := 1; i < len(match); i++ {
				params[keys[i-1]] = match[i]
			}
		}
	}
}
