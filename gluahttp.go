package gluahttp

import "github.com/yuin/gopher-lua"
import "net/http"
import "fmt"
import "io/ioutil"
import "strings"

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

var exports = map[string]lua.LGFunction{
	"get":     get,
	"head":    head,
	"patch":   patch,
	"post":    post,
	"put":     put,
	"request": request,
}

func get(L *lua.LState) int {
	return doRequest(L, "get", L.ToString(1), L.ToTable(2))
}

func head(L *lua.LState) int {
	return doRequest(L, "head", L.ToString(1), L.ToTable(2))
}

func patch(L *lua.LState) int {
	return doRequest(L, "patch", L.ToString(1), L.ToTable(2))
}

func post(L *lua.LState) int {
	return doRequest(L, "post", L.ToString(1), L.ToTable(2))
}

func put(L *lua.LState) int {
	return doRequest(L, "put", L.ToString(1), L.ToTable(2))
}

func request(L *lua.LState) int {
	return doRequest(L, L.ToString(1), L.ToString(2), L.ToTable(3))
}

func doRequest(L *lua.LState, method string, url string, options *lua.LTable) int {
	req, err := http.NewRequest(strings.ToUpper(method), url, nil)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("%s", err)))
		return 2
	}

	if options != nil {
		if reqHeaders, ok := options.RawGet(lua.LString("headers")).(*lua.LTable); ok {
			reqHeaders.ForEach(func(key lua.LValue, value lua.LValue) {
				req.Header.Set(key.String(), value.String())
			})
		}

		switch reqQuery := options.RawGet(lua.LString("query")).(type) {
		case *lua.LNilType:
			break

		case lua.LString:
			req.URL.RawQuery = reqQuery.String()
			break
		}

		switch reqForm := options.RawGet(lua.LString("form")).(type) {
		case *lua.LNilType:
			break

		case lua.LString:
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Body = ioutil.NopCloser(strings.NewReader(reqForm.String()))
			break
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("%s", err)))
		return 2
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("%s", err)))
		return 2
	}

	L.Push(lua.LString(body))
	L.Push(lua.LNumber(resp.StatusCode))

	headers := L.NewTable()
	for key, _ := range resp.Header {
		headers.RawSet(lua.LString(key), lua.LString(resp.Header.Get(key)))
	}
	L.Push(headers)

	return 3
}
