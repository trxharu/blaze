package blaze

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
)

// test handler
type HandleFunc[T any] struct{}
type Context struct {
	Val string
}

var (
	TestContext = ServerCtx[Context]{
		Data: Context{
			Val: "YOLO",
		},
	}
)

func (*HandleFunc[T]) Handle(req Request[T], res *Response) {
	id := req.GetParam("id", "0")
	fmt.Println("HandleFunc called: ", req.URL, id)
}

func TestRootRoute(t *testing.T) {
	router := NewRouter[Context]()
	router.Get("/", &HandleFunc[Context]{})
	router.Exec(Request[Context]{URL: &url.URL{Path: "/"}, Method: "GET"}, &Response{})
}

func TestApiRoute(t *testing.T) {
	router := NewRouter[Context]()
	router.Get("/api", &HandleFunc[Context]{})
	router.Exec(Request[Context]{URL: &url.URL{Path: "/api"}, Method: "GET"}, &Response{})
}

func TestMoreRoutes(t *testing.T) {
	router := NewRouter[Context]()
	router.Get("/", &HandleFunc[Context]{})
	router.Get("/api", &HandleFunc[Context]{})
	router.Get("/health", &HandleFunc[Context]{})
	router.Exec(Request[Context]{URL: &url.URL{Path: "/"}, Method: "GET"}, &Response{})
	router.Exec(Request[Context]{URL: &url.URL{Path: "/api"}, Method: "GET"}, &Response{})
	router.Exec(Request[Context]{URL: &url.URL{Path: "/health"}, Method: "GET"}, &Response{})
}

func TestChainedRoute(t *testing.T) {
	router := NewRouter[Context]()
	router.Get("/api/v1", &HandleFunc[Context]{})
	router.Get("/", &HandleFunc[Context]{})
	router.Exec(Request[Context]{URL: &url.URL{Path: "/api/v1"}, Method: "GET"}, &Response{})
	router.Exec(Request[Context]{URL: &url.URL{Path: "/"}, Method: "GET"}, &Response{})
}

func TestNestedRoute(t *testing.T) {
	root := NewRouter[Context]()
	subroute := NewRouter[Context]()
	subroute.Get("/api", &HandleFunc[Context]{})

	root.SubRoute("/v1", subroute)
	root.Exec(Request[Context]{
		URL:    &url.URL{Path: "/v1/api"},
		Method: "GET",
	},
		&Response{Header: make(http.Header)},
	)
}

func TestParams(t *testing.T) {
	root := NewRouter[Context]()
	root.Get("/api/:id", &HandleFunc[Context]{})
	root.Exec(Request[Context]{
		URL:    &url.URL{Path: "/api/noobmaster69"},
		Params: make(map[string]string),
		Method: "GET",
	},
		&Response{
			Header: make(http.Header),
		},
	)
}

func TestCheckWrongMethod(t *testing.T) {
	root := NewRouter[Context]()
	response := &Response{
		Header: make(http.Header),
	}
	root.Get("/health", &HandleFunc[Context]{})
	root.Exec(Request[Context]{
		URL:    &url.URL{Path: "/health"},
		Method: "POST",
	},
		response,
	)

	if response.StatusCode == http.StatusMethodNotAllowed {
		t.Logf("status code is %d, which is expected", response.StatusCode)
	} else {
		t.Fatalf("status code is %d, expected %d", response.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestCtx(t *testing.T) {
	root := NewRouter[Context]()
	root.Get("/", ResolveFunc[Context](func(req Request[Context], res *Response) {
		res.Logger.Println(req.Context.Data.Val)
	}))
	root.Exec(Request[Context]{
		URL:     &url.URL{Path: "/"},
		Method:  "GET",
		Context: TestContext,
	}, &Response{Logger: log.Default()})
}
