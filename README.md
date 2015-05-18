# gluahttp

[![](https://travis-ci.org/cjoudrey/gluahttp.svg)](https://travis-ci.org/cjoudrey/gluahttp)

gluahttp provides an easy way to make HTTP requests from within [GopherLua](https://github.com/yuin/gopher-lua).

## Installation

```
go get github.com/cjoudrey/gluahttp
```

## Usage

```lua
local http = require("http")

body, status, headers = http.request("GET", "http://example.com", {
  query="page=1"
  headers={
    Accept="*/*"
  }
})
```

## API

### http.get(url [, options])

- `url`: A `string` URL of the page to load.
- `options`: A `table` with one or many of the following parameters:
 - `query`: Query string in the form of a `string`.
 - `headers`: `table` of additional headers to send with the request.

Return:

- `body`: A `string` containing the response body.
- `status`: A `number` containing the HTTP status code.
- `headers`: A `table` containing the response headers.

In the event of an error, the return is as follows:

- `nil`
- `error`: A `string` containing the error message.

### http.head(url [, options])

- `url`: A `string` URL of the page to load.
- `options`: A `table` with one or many of the following parameters:
 - `query`: Query string in the form of a `string`.
 - `headers`: `table` of additional headers to send with the request.

Return:

- `body`: A `string` containing the response body.
- `status`: A `number` containing the HTTP status code.
- `headers`: A `table` containing the response headers.

In the event of an error, the return is as follows:

- `nil`
- `error`: A `string` containing the error message.

### http.patch(url [, options])

- `url`: A `string` URL of the page to load.
- `options`: A `table` with one or many of the following parameters:
 - `query`: Query string in the form of a `string`.
 - `form`: `string` value of the request's body data and set's `Content-Type` to `application/x-www-form-urlencoded`.
 - `headers`: `table` of additional headers to send with the request.

Return:

- `body`: A `string` containing the response body.
- `status`: A `number` containing the HTTP status code.
- `headers`: A `table` containing the response headers.

In the event of an error, the return is as follows:

- `nil`
- `error`: A `string` containing the error message.

### http.post(url [, options])

- `url`: A `string` URL of the page to load.
- `options`: A `table` with one or many of the following parameters:
 - `query`: Query string in the form of a `string`.
 - `form`: `string` value of the request's body data and set's `Content-Type` to `application/x-www-form-urlencoded`.
 - `headers`: `table` of additional headers to send with the request.

Return:

- `body`: A `string` containing the response body.
- `status`: A `number` containing the HTTP status code.
- `headers`: A `table` containing the response headers.

In the event of an error, the return is as follows:

- `nil`
- `error`: A `string` containing the error message.

### http.put(url [, options])

- `url`: A `string` URL of the page to load.
- `options`: A `table` with one or many of the following parameters:
 - `query`: Query string in the form of a `string`.
 - `form`: `string` value of the request's body data and set's `Content-Type` to `application/x-www-form-urlencoded`.
 - `headers`: `table` of additional headers to send with the request.

Return:

- `body`: A `string` containing the response body.
- `status`: A `number` containing the HTTP status code.
- `headers`: A `table` containing the response headers.

In the event of an error, the return is as follows:

- `nil`
- `error`: A `string` containing the error message.

### http.request(method, url [, options])

- `method`: The HTTP request method.
- `url`: A `string` URL of the page to load.
- `options`: A `table` with one or many of the following parameters:
 - `query`: Query string in the form of a `string`.
 - `form`: `string` value of the request's body data and set's `Content-Type` to `application/x-www-form-urlencoded`.
 - `headers`: `table` of additional headers to send with the request.

Return:

- `body`: A `string` containing the response body.
- `status`: A `number` containing the HTTP status code.
- `headers`: A `table` containing the response headers.

In the event of an error, the return is as follows:

- `nil`
- `error`: A `string` containing the error message.
