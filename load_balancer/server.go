package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"
)

// Config is a configuration.
type Config struct {
	Proxy    Proxy     `json:"proxy"`
	Backends []Backend `json:"backends"`
}

// Proxy is a reverse proxy, and means load balancer.
type Proxy struct {
	Port string `json:"port"`
}

// Backend is servers which load balancer is transferred.
type Backend struct {
	URL    string `json:"url"`
	IsDead bool
	mu     sync.RWMutex
}

// SetDead updates the value of IsDead in Backend.
func (backend *Backend) SetDead(b bool) {
	backend.mu.Lock()
	backend.IsDead = b
	backend.mu.Unlock()
}

// GetIsDead returns the value of IsDead in Backend.
func (backend *Backend) GetIsDead() bool {
	backend.mu.RLock()
	isAlive := backend.IsDead
	backend.mu.RUnlock()
	return isAlive
}

var mu sync.Mutex
var idx int = 0

// lbHandler is a handler for loadbalancing
func lbHandler(w http.ResponseWriter, r *http.Request) {
	maxLen := len(cfg.Backends)
	// Round Robin
	mu.Lock()
	currentBackend := cfg.Backends[idx%maxLen]
	if currentBackend.GetIsDead() {
		idx++
	}
	targetURL, err := url.Parse(cfg.Backends[idx%maxLen].URL)
	if err != nil {
		log.Panicln(err.Error())
	}
	idx++
	mu.Unlock()
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		// NOTE: It is better to implement retry.
		log.Printf("%v is dead.", targetURL)
		currentBackend.SetDead(true)
		lbHandler(w, r)
	}
	reverseProxy.ServeHTTP(w, r)
}

// pingBackend checks if the backend is alive.
func isAlive(url *url.URL) bool {
	conn, err := net.DialTimeout("tcp", url.Host, time.Minute*1)
	if err != nil {
		log.Printf("Unreachable to %v, error:%v", url.Host, err.Error())
		return false
	}
	defer conn.Close()
	return true
}

// healthCheck is a function for healthcheck
func healthCheck() {
	t := time.NewTicker(time.Minute * 1)
	for {
		select {
		case <-t.C:
			for _, backend := range cfg.Backends {
				pingURL, err := url.Parse(backend.URL)
				if err != nil {
					log.Panicln(err.Error())
				}
				isAlive := isAlive(pingURL)
				backend.SetDead(!isAlive)
				msg := "ok"
				if !isAlive {
					msg = "dead"
				}
				log.Printf("%v checked %v by healthcheck", backend.URL, msg)
			}
		}
	}

}

var cfg Config

// Serve serves a loadbalancer.
func Serve() {
	data, err := os.ReadFile("C:/Users/hashem/Desktop/programing/my_big_project/load_balancer/config.json")
	if err != nil {
		log.Panicln(err.Error())
	}
	json.Unmarshal(data, &cfg)

	go healthCheck()

	s := http.Server{
		Addr:    ":" + cfg.Proxy.Port,
		Handler: http.HandlerFunc(lbHandler),
	}
	if err = s.ListenAndServe(); err != nil {
		log.Panicln(err.Error())
	}
}

func serveBackend(name string, port string) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Backend server name:%v and status %v\n", name, http.StatusOK)
	}))
	http.ListenAndServe(port, mux)
}

func main() {
	wg := new(sync.WaitGroup)
	wg.Add(5)

	go func() {
		Serve()
		wg.Done()
	}()

	go func() {
		serveBackend("web1", ":8081")
		wg.Done()
	}()

	go func() {
		serveBackend("web2", ":8082")
		wg.Done()
	}()

	go func() {
		serveBackend("web3", ":8083")
		wg.Done()
	}()

	go func() {
		serveBackend("web4", ":8084")
		wg.Done()
	}()

	wg.Wait()
}
