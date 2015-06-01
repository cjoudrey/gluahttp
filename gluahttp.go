package gluahttp

import "github.com/yuin/gopher-lua"
import "net/http"
import "net/http/cookiejar"
import "fmt"
import "errors"
import "io/ioutil"
import "strings"

type httpModule struct {
	client *http.Client
}

type empty struct{}

func NewHttpModule() *httpModule {
	cookieJar, _ := cookiejar.New(nil)

	return &httpModule{
		client: &http.Client{
			Jar: cookieJar,
		},
	}
}

func (h *httpModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"get":           h.get,
		"delete":        h.delete,
		"head":          h.head,
		"patch":         h.patch,
		"post":          h.post,
		"put":           h.put,
		"request":       h.request,
		"request_batch": h.requestBatch,
	})
	L.Push(mod)
	return 1
}

func (h *httpModule) get(L *lua.LState) int {
	return h.doRequestAndPush(L, "get", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) delete(L *lua.LState) int {
	return h.doRequestAndPush(L, "delete", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) head(L *lua.LState) int {
	return h.doRequestAndPush(L, "head", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) patch(L *lua.LState) int {
	return h.doRequestAndPush(L, "patch", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) post(L *lua.LState) int {
	return h.doRequestAndPush(L, "post", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) put(L *lua.LState) int {
	return h.doRequestAndPush(L, "put", L.ToString(1), L.ToTable(2))
}

func (h *httpModule) request(L *lua.LState) int {
	return h.doRequestAndPush(L, L.ToString(1), L.ToString(2), L.ToTable(3))
}

func (h *httpModule) requestBatch(L *lua.LState) int {
	requests := L.ToTable(1)
	amountRequests := requests.Len()

	errs := make([]error, amountRequests)
	responses := make([]*lua.LTable, amountRequests)
	sem := make(chan empty, amountRequests)

	i := 0

	requests.ForEach(func(_ lua.LValue, value lua.LValue) {
		requestTable := toTable(value)

		if requestTable != nil {
			method := requestTable.RawGet(lua.LNumber(1)).String()
			url := requestTable.RawGet(lua.LNumber(2)).String()
			options := toTable(requestTable.RawGet(lua.LNumber(3)))

			go func(i int, L *lua.LState, method string, url string, options *lua.LTable) {
				response, err := h.doRequest(L, method, url, options)

				if err == nil {
					errs[i] = nil
					responses[i] = response
				} else {
					errs[i] = err
					responses[i] = nil
				}

				sem <- empty{}
			}(i, L, method, url, options)
		} else {
			errs[i] = errors.New("Request must be a table")
			responses[i] = nil
			sem <- empty{}
		}

		i = i + 1
	})

	for i = 0; i < amountRequests; i++ {
		<-sem
	}

	hasErrors := false
	errorsTable := L.NewTable()
	responsesTable := L.NewTable()
	for i = 0; i < amountRequests; i++ {
		if errs[i] == nil {
			responsesTable.Append(responses[i])
			errorsTable.Append(lua.LNil)
		} else {
			responsesTable.Append(lua.LNil)
			errorsTable.Append(lua.LString(fmt.Sprintf("%s", errs[i])))
			hasErrors = true
		}
	}

	if hasErrors {
		L.Push(responsesTable)
		L.Push(errorsTable)
		return 2
	} else {
		L.Push(responsesTable)
		return 1
	}
}

func (h *httpModule) doRequest(L *lua.LState, method string, url string, options *lua.LTable) (*lua.LTable, error) {
	req, err := http.NewRequest(strings.ToUpper(method), url, nil)
	if err != nil {
		return nil, err
	}

	if options != nil {
		if reqHeaders, ok := options.RawGet(lua.LString("headers")).(*lua.LTable); ok {
			reqHeaders.ForEach(func(key lua.LValue, value lua.LValue) {
				req.Header.Set(key.String(), value.String())
			})
		}

		if reqCookies, ok := options.RawGet(lua.LString("cookies")).(*lua.LTable); ok {
			reqCookies.ForEach(func(key lua.LValue, value lua.LValue) {
				req.AddCookie(&http.Cookie{Name: key.String(), Value: value.String()})
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

	res, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	headers := L.NewTable()
	for key, _ := range res.Header {
		headers.RawSetString(key, lua.LString(res.Header.Get(key)))
	}

	cookies := L.NewTable()
	for _, cookie := range res.Cookies() {
		cookies.RawSetString(cookie.Name, lua.LString(cookie.Value))
	}

	response := L.NewTable()
	response.RawSetString("body", lua.LString(body))
	response.RawSetString("headers", headers)
	response.RawSetString("cookies", cookies)
	response.RawSetString("status_code", lua.LNumber(res.StatusCode))
	response.RawSetString("url", lua.LString(res.Request.URL.String()))

	return response, nil
}

func (h *httpModule) doRequestAndPush(L *lua.LState, method string, url string, options *lua.LTable) int {
	response, err := h.doRequest(L, method, url, options)

	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(fmt.Sprintf("%s", err)))
		return 2
	}

	L.Push(response)
	return 1
}

func toTable(v lua.LValue) *lua.LTable {
	if lv, ok := v.(*lua.LTable); ok {
		return lv
	}
	return nil
}
