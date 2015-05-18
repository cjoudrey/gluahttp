package gluahttp

import "github.com/yuin/gopher-lua"
import "testing"
import "os"
import "bytes"
import "io"
import "net/http"
import "net"
import "fmt"
import "net/http/httputil"
import "strings"

func TestRequestNoMethod(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.request()

			print(body)
			print(status)
			print(headers)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `nil
unsupported protocol scheme ""
nil
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestNoUrl(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.request("get")

			print(body)
			print(status)
			print(headers)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `nil
Get : unsupported protocol scheme ""
nil
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestGet(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.request("get", "http://` + listener.Addr().String() + `")

			print(body)
			print(status)
			print(headers["Content-Length"])
			print(headers["Content-Type"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `GET / HTTP/1.1
Host: ` + listener.Addr().String() + `
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


200
97
text/plain; charset=utf-8
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestPostForm(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.request("post", "http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(body)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `POST / HTTP/1.1
Host: ` + listener.Addr().String() + `
Transfer-Encoding: chunked
Accept-Encoding: gzip
Content-Type: application/x-www-form-urlencoded
User-Agent: Go 1.1 package http

1c
username=bob&password=secret
0


`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestGetHeaders(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.request("get", "http://` + listener.Addr().String() + `", {
				headers={
					Something="Test"
				}
			})
	
			print(body)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `GET / HTTP/1.1
Host: ` + listener.Addr().String() + `
Accept-Encoding: gzip
Something: Test
User-Agent: Go 1.1 package http


`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestGetQuery(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.request("get", "http://` + listener.Addr().String() + `", {
				query="page=1"
			})
	
			print(body)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `GET /?page=1 HTTP/1.1
Host: ` + listener.Addr().String() + `
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestGet(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.get("http://` + listener.Addr().String() + `", {
				query="page=1"
			})

			print(body)
			print(status)
			print(headers["Content-Length"])
			print(headers["Content-Type"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `GET /?page=1 HTTP/1.1
Host: ` + listener.Addr().String() + `
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


200
104
text/plain; charset=utf-8
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestHead(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.head("http://` + listener.Addr().String() + `", {
				query="page=1"
			})

			print(headers["X-Request-Method"])
			print(headers["X-Request-Uri"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `HEAD
/?page=1
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestPost(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.post("http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(body)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `POST / HTTP/1.1
Host: ` + listener.Addr().String() + `
Transfer-Encoding: chunked
Accept-Encoding: gzip
Content-Type: application/x-www-form-urlencoded
User-Agent: Go 1.1 package http

1c
username=bob&password=secret
0


`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestPatch(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.patch("http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(body)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `PATCH / HTTP/1.1
Host: ` + listener.Addr().String() + `
Transfer-Encoding: chunked
Accept-Encoding: gzip
Content-Type: application/x-www-form-urlencoded
User-Agent: Go 1.1 package http

1c
username=bob&password=secret
0


`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestPut(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			body, status, headers = http.put("http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(body)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `PUT / HTTP/1.1
Host: ` + listener.Addr().String() + `
Transfer-Encoding: chunked
Accept-Encoding: gzip
Content-Type: application/x-www-form-urlencoded
User-Agent: Go 1.1 package http

1c
username=bob&password=secret
0


`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func captureStdout(inner func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	inner()

	w.Close()
	os.Stdout = oldStdout
	out := strings.Replace(<-outC, "\r", "", -1)

	return out
}

func setupEchoServer(listener net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("X-Request-Method", req.Method)
		w.Header().Set("X-Request-Uri", req.URL.String())
		if debug, err := httputil.DumpRequest(req, true); err == nil {
			fmt.Fprint(w, string(debug))
		} else {
			fmt.Fprintf(w, "Error: %s", err)
		}
	})
	s := &http.Server{
		Handler: mux,
	}
	go s.Serve(listener)
}
