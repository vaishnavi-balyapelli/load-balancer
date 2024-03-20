package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}
type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}
type LoadBalancer struct {
	port            string
	roundRobinCount int
	server          []Server
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)
	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}
func handleErr(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
func newLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:            port,
		roundRobinCount: 0,
		server:          servers,
	}
}
func (s *simpleServer) Address() string {
	return s.addr
}
func (s *simpleServer) IsAlive() bool {
	return true
}
func (s *simpleServer) Serve(rw http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(rw, r)
}
func (lb *LoadBalancer) GetNextAvailableServer() Server {
	server := lb.server[lb.roundRobinCount%len(lb.server)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.server[lb.roundRobinCount%len(lb.server)]
	}
	lb.roundRobinCount++
	return server
}
func (lb *LoadBalancer) serverProxy(rw http.ResponseWriter, r *http.Request) {
	targetServer := lb.GetNextAvailableServer()
	fmt.Printf("Forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(rw, r)
}
func main() {
	servers := []Server{
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.bing.com"),
		newSimpleServer("https://www.duckduckgo.com"),
	}
	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, r *http.Request) {
		lb.serverProxy(rw, r)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("Serving request at 'localhost:%v' \n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
