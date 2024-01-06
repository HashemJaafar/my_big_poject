package http

import (
	"fmt"
	"net/http"
	"testing"
	"time"
	"tools"
)

func TestUrl(t *testing.T) {
	url := Url("localhost", 3333, "/work")
	tools.Test(url, "http://localhost:3333/work")
}

const (
	host = "localhost"
	port = 8000
)

func Test(t *testing.T) {
	var f1 = Create[int, string](host, port, "/f1", func(req int) (string, error) { return fmt.Sprint(req), nil })
	var f2 = Create[string, string](host, port, "/f2", func(req string) (string, error) { return fmt.Sprint(req), nil })
	var f3 = Create[string, error](host, port, "/f3", func(req string) (error, error) { return fmt.Errorf(req), nil })
	var f4 = Create[error, string](host, port, "/f4", func(req error) (string, error) { return req.Error(), nil })

	go func() {
		mux := http.NewServeMux()
		HandleFunc(mux, f1.Pattern, f1.Handle)
		HandleFunc(mux, f2.Pattern, f2.Handle)
		HandleFunc(mux, f3.Pattern, f3.Handle)
		HandleFunc(mux, f4.Pattern, f4.Handle)
		ListenAndServe(mux, host, port)
	}()

	time.Sleep(100 * time.Millisecond)

	{
		r, err := f1.Request(1)
		tools.Test(r, "1")
		tools.Test(err, nil)
	}
	{
		r, err := f2.Request("yes")
		tools.Test(r, "yes")
		tools.Test(err, nil)
	}
	{
		r, err := f3.Request("yes")
		tools.Test(r, nil)
		tools.Test(err.Error(), "json: cannot unmarshal object into Go value of type error")
	}
	{
		r, err := f4.Request(fmt.Errorf("yes"))
		tools.Test(r, "")
		tools.Test(err.Error(), `json: cannot unmarshal object into Go value of type error
500 Internal Server Error`)
	}
}
