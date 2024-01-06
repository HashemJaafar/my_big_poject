package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	tools "tools"
)

func ListenAndServe(mux *http.ServeMux, host string, port uint) {
	server := http.Server{
		Addr:    fmt.Sprintf("%v:%d", host, port),
		Handler: mux,
	}
	fmt.Println("############ start ############")
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("error running http server: %s\n", err)
		}
	}
}

func Url(host string, port uint, pattern string) string {
	return fmt.Sprintf("http://%v:%v%v", host, port, pattern)
}

type networkClass[ReqT any, ResT any] struct {
	Pattern string
	Handle  func(req []byte) ([]byte, error)
	Request func(req ReqT) (ResT, error)
	Process func(req ReqT) (ResT, error)
}

func Create[ReqT any, ResT any](host string, port uint, pattern string, process func(req ReqT) (ResT, error)) networkClass[ReqT, ResT] {
	var zero ResT
	var s networkClass[ReqT, ResT]

	s.Pattern = pattern
	s.Handle = func(req []byte) ([]byte, error) {
		d, err := tools.Decode[ReqT](req)
		if err != nil {
			return nil, err
		}
		res, err := process(d)
		if err != nil {
			return nil, err
		}
		e := tools.Encode(res)
		return e, nil
	}
	s.Request = func(req ReqT) (ResT, error) {
		e := tools.Encode(req)
		res, err := NewRequest(host, port, pattern, e)
		if err != nil {
			return zero, err
		}
		d, err := tools.Decode[ResT](res)
		if err != nil {
			return zero, err
		}
		return d, nil
	}
	s.Process = process

	return s
}

func NewRequest(host string, port uint, pattern string, reqBodyByte []byte) ([]byte, error) {

	req, err := http.NewRequest(http.MethodPost, Url(host, port, pattern), bytes.NewReader(reqBodyByte))
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: 30 * time.Second}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	resBodyByte, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s%v", resBodyByte, res.Status)
	}

	return resBodyByte, nil
}

func HandleFunc(mux *http.ServeMux, pattern string, handle func(req []byte) ([]byte, error)) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		reqBodyByte, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resBodyByte, err := handle(reqBodyByte)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(resBodyByte)
	})
}
