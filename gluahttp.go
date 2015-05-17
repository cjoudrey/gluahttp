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
	"request": request,
}

func request(L *lua.LState) int {
	req, err := http.NewRequest(strings.ToUpper(L.ToString(1)), L.ToString(2), nil)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("%s", err)))
		return 2
	}

	if options := L.ToTable(3); options != nil {
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
