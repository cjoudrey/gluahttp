# gluahttp

[![](https://travis-ci.org/cjoudrey/gluahttp.svg)](https://travis-ci.org/cjoudrey/gluahttp)

gluahttp provides an easy way to make HTTP requests from within [GopherLua](https://github.com/yuin/gopher-lua).

## Installation

```
go get github.com/cjoudrey/gluahttp
```

## Usage

```go
L := lua.NewState()
defer L.Close()

L.PreloadModule("http", NewHttpModule().Loader)

if err := L.DoString(`

    local http = require("http")

    response, error_message = http.request("GET", "http://example.com", {
        query="page=1"
        headers={
            Accept="*/*"
        }
    })

`); err != nil {
    panic(err)
}
```

## API

### http.get(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| query   | String | URL encoded query params |
| headers | Table  | Additional headers to send with the request |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.head(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| query   | String | URL encoded query params |
| headers | Table  | Additional headers to send with the request |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.patch(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| query   | String | URL encoded query params |
| form    | String | URL encoded request body. This will also set the `Content-Type` header to `application/x-www-form-urlencoded` |
| headers | Table  | Additional headers to send with the request |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.post(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| query   | String | URL encoded query params |
| form    | String | URL encoded request body. This will also set the `Content-Type` header to `application/x-www-form-urlencoded` |
| headers | Table  | Additional headers to send with the request |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.put(url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| query   | String | URL encoded query params |
| form    | String | URL encoded request body. This will also set the `Content-Type` header to `application/x-www-form-urlencoded` |
| headers | Table  | Additional headers to send with the request |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.request(method, url [, options])

**Attributes**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| method  | String | The HTTP request method |
| url     | String | URL of the resource to load |
| options | Table  | Additional options |

**Options**

| Name    | Type   | Description |
| ------- | ------ | ----------- |
| query   | String | URL encoded query params |
| form    | String | URL encoded request body. This will also set the `Content-Type` header to `application/x-www-form-urlencoded` |
| headers | Table  | Additional headers to send with the request |

**Returns**

[http.response](#httpresponse) or (nil, error message)

### http.response

The `http.response` table contains information about a completed HTTP request.

**Attributes**

| Name        | Type   | Description |
| ----------- | ------ | ----------- |
| body        | String | The HTTP response body |
| headers     | Table  | The HTTP response headers |
| status_code | Number | The HTTP response status code |
