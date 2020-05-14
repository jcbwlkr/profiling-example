An example web server for demonstrating `net/http/pprof`

There is a goroutine leak in this program. Generate a few hundred requests then find the leak using pprof:

```
go tool pprof -http=:9999 http://localhost:6060/debug/pprof/goroutine
```

More about profiling at:

- https://golang.org/pkg/net/http/pprof/
- https://golang.org/pkg/runtime/pprof/
- https://blog.golang.org/pprof
- https://github.com/google/pprof/blob/master/doc/README.md
