# blaze
Go Minimal HTTP framework with Chain based routing and Context

## Introduction
Blaze is a minimal HTTP framework for Golang, built only using standard library.
It currently features:
1. Chain based routing.
2. Basic http server setup using logger and context.
3. Middleware support.

## Usage
```go
import (
	"net/http"
	"github.com/trxharu/blaze"
)

type Context struct {
	Db *sql.DB
}

func main() {
	server := blaze.DefaultBlazeServer[Context]()

	middlewareChain := blaze.ChainMiddleware[Context](
		blaze.LoggingMiddleware,
		blaze.CorsMiddleware,
	)

	ctx := blaze.ServerCtx[Context]{
		Data: Context{Db: // your db instance},
	}

	server.UseContext(ctx)

	router := blaze.NewRouter[Context]()
	router.Get("/health", blaze.ResolveFunc[Context](health))

	server.UseRouter(router)
	server.UseMiddleware(middlewareChain)
	server.Serve()
}

func health(req blaze.Request[Context], res *blaze.Response) {
	res.WriteText(http.StatusOK, "Healthy")
}
```
