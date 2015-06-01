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

	L.PreloadModule("http", NewHttpModule().Loader)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.request()

			print(response)
			print(error)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `nil
unsupported protocol scheme ""
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestNoUrl(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.request("get")

			print(response)
			print(error)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `nil
Get : unsupported protocol scheme ""
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestBatch(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			responses, errors = http.request_batch({
				{"get", "http://` + listener.Addr().String() + `", {query="page=1"}},
				{"post", "http://` + listener.Addr().String() + `/set_cookie"},
				{"post", ""},
				1
			})

			print(responses[1]["body"])
			print(responses[2]["body"])
			print(responses[2]["cookies"]["session_id"])
			print(responses[3])
			print(responses[4])

			print(errors[1])
			print(errors[2])
			print(errors[3])
			print(errors[4])

			responses, errors = http.request_batch({
				{"get", "http://` + listener.Addr().String() + `/get_cookie"}
			})

			print(responses[1]["body"])
			print(errors)
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `GET /?page=1 HTTP/1.1
Host: ` + listener.Addr().String() + `
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


Cookie set!
12345
nil
nil
nil
nil
Post : unsupported protocol scheme ""
Request must be a table
session_id=12345
nil
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestGet(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.request("get", "http://` + listener.Addr().String() + `")

			print(response["body"])
			print(response["status_code"])
			print(response["headers"]["Content-Length"])
			print(response["headers"]["Content-Type"])
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

func TestRequestGetWithRedirect(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.request("get", "http://` + listener.Addr().String() + `/redirect")

			print(response["body"])
			print(response["status_code"])
			print(response["url"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `GET / HTTP/1.1
Host: ` + listener.Addr().String() + `
Accept-Encoding: gzip
Referer: http://` + listener.Addr().String() + `/redirect
User-Agent: Go 1.1 package http


200
http://` + listener.Addr().String() + `/
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestPostForm(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.request("post", "http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(response["body"])
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

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.request("get", "http://` + listener.Addr().String() + `", {
				headers={
					Something="Test"
				}
			})
	
			print(response["body"])
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

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.request("get", "http://` + listener.Addr().String() + `", {
				query="page=1"
			})
	
			print(response["body"])
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

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.get("http://` + listener.Addr().String() + `", {
				query="page=1"
			})

			print(response["body"])
			print(response["status_code"])
			print(response["headers"]["Content-Length"])
			print(response["headers"]["Content-Type"])
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

func TestDelete(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.delete("http://` + listener.Addr().String() + `", {
				query="page=1"
			})

			print(response["body"])
			print(response["status_code"])
			print(response["headers"]["Content-Length"])
			print(response["headers"]["Content-Type"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `DELETE /?page=1 HTTP/1.1
Host: ` + listener.Addr().String() + `
Accept-Encoding: gzip
User-Agent: Go 1.1 package http


200
107
text/plain; charset=utf-8
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestHead(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.head("http://` + listener.Addr().String() + `", {
				query="page=1"
			})

			print(response["headers"]["X-Request-Method"])
			print(response["headers"]["X-Request-Uri"])
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

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.post("http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(response["body"])
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

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.patch("http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(response["body"])
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

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.put("http://` + listener.Addr().String() + `", {
				form="username=bob&password=secret"
			})

			print(response["body"])
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

func TestResponseCookies(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.get("http://` + listener.Addr().String() + `/set_cookie")
			print(response["status_code"])
			print(response["body"])
			print(response["cookies"]["session_id"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `200
Cookie set!
12345
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestRequestCookies(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.get("http://` + listener.Addr().String() + `/get_cookie", {
				cookies={
					["session_id"]="test"
				}
			})
			print(response["status_code"])
			print(response["body"])

			response, error = http.get("http://` + listener.Addr().String() + `/get_cookie")
			print(response["status_code"])
			print(response["body"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `200
session_id=test
200
<nil>
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}
}

func TestCookiesPerLState(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out := captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")
			response, error = http.get("http://` + listener.Addr().String() + `/set_cookie")
			print(response["status_code"])
			print(response["body"])

			response, error = http.get("http://` + listener.Addr().String() + `/get_cookie")
			print(response["status_code"])
			print(response["body"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `200
Cookie set!
200
session_id=12345
`; expected != out {
		t.Errorf("Expected output does not match actual output\nExpected: %s\nActual: %s", expected, out)
	}

	L = lua.NewState()
	defer L.Close()

	L.PreloadModule("http", NewHttpModule().Loader)

	listener, _ = net.Listen("tcp", "127.0.0.1:0")
	setupEchoServer(listener)

	out = captureStdout(func() {
		if err := L.DoString(`
			local http = require("http")

			response, error = http.get("http://` + listener.Addr().String() + `/get_cookie")
			print(response["status_code"])
			print(response["body"])
		`); err != nil {
			t.Errorf("Failed to evaluate script: %s", err)
		}
	})

	if expected := `200
<nil>
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
	mux.HandleFunc("/set_cookie", func(w http.ResponseWriter, req *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "12345"})
		fmt.Fprint(w, "Cookie set!")
	})
	mux.HandleFunc("/get_cookie", func(w http.ResponseWriter, req *http.Request) {
		session_id, _ := req.Cookie("session_id")
		fmt.Fprint(w, session_id)
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/", http.StatusFound)
	})
	s := &http.Server{
		Handler: mux,
	}
	go s.Serve(listener)
}
