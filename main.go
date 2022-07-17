package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func (s *simpleServer) Address() string {
	return s.addr
}

func (s *simpleServer) IsAlive() bool {
	return true
}

func (s *simpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.Proxy.ServeHTTP(w, r)
}

type simpleServer struct {
	addr  string
	Proxy *httputil.ReverseProxy
}
type Server interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type LoadBalancer struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		servers:         servers,
		roundRobinCount: 0,
	}
}
func newsimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)
	return &simpleServer{
		addr:  addr,
		Proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount += 1
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding request to addr, %q\n", targetServer.Address())
	targetServer.Serve(w, r)
}

func main() {
	servers := []Server{
		newsimpleServer("https://www.facebook.com"),
		newsimpleServer("https://www.bing.com"),
		newsimpleServer("http://www.duckduckgo.com"),
	}

	lb := NewLoadBalancer("8000", servers)
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Println("Serving request at localhost:8001")
	http.ListenAndServe(":8001", nil)
}
