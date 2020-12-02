## Simple Reverse Proxy with Caching capability in Go

### Testing

I've added db.json where you can `json-serve --watch db.json` it for a simple api server from localhost which listens on port `3000`. Then once your api service is running. You may run the the code using `go run .` N.B. [Caching is only applicable for GET requests only]

Simple query you may use: `http://locahost:8080/posts`.
