package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
)

func handleErr(err error){
	if err!=nil{
		fmt.Printf("error: %v\n",err)
		os.Exit(1)
	}
}

type Server interface{
	IsAlive() bool
	Address() string
	Serve( res http.ResponseWriter, req *http.Request)
	GetConnections() int
	IncrConnections()
	DecrConnections()
}


type simpleServer struct{
	address string
	proxy *httputil.ReverseProxy
	activeConnections int
	mutex sync.Mutex
}

func newsimpleServer(address string)  *simpleServer{
	serverURL,err:=url.Parse(address)
	handleErr(err)

	return &simpleServer{
		address: address,
		proxy: httputil.NewSingleHostReverseProxy(serverURL),
	}
}

func (s *simpleServer)Address() string{
	return s.address
}

func (s *simpleServer)IsAlive() bool{
	return true
}

func (s *simpleServer)Serve(res http.ResponseWriter,req *http.Request){
	s.proxy.ServeHTTP(res,req)
}

func (s *simpleServer)GetConnections() int{
	return s.activeConnections
}

func (s *simpleServer)IncrConnections(){
	s.mutex.Lock()
    defer s.mutex.Unlock()
    s.activeConnections++
	fmt.Printf("Server %s incremented connections: %d\n", s.address, s.activeConnections)
}

func (s *simpleServer)DecrConnections(){
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.activeConnections--
	fmt.Printf("Server %s decremented connections: %d\n", s.address, s.activeConnections)
}


type LoadBalancer struct{
	port string
	roundRobinCount int
	servers []Server
}

func newLoadBalancer(port string,servers []Server) *LoadBalancer{
	return &LoadBalancer{
		port:port,
		roundRobinCount: 0,
		servers: servers,
	}
}

func (lb *LoadBalancer)getNextRoundRobin() Server{
	//  go to the next till you get a alive one
	server:= lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive(){
		lb.roundRobinCount++
		server=lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *LoadBalancer)getNextLeastConnection() Server{
	minConnections := lb.servers[0].GetConnections()
	minServer := lb.servers[0]

	for _, server := range lb.servers {
		if server.GetConnections() < minConnections {
			minConnections = server.GetConnections()
			minServer = server
		}
	}

	minServer.IncrConnections()
	return minServer
}

func (lb *LoadBalancer) serveProxyRoundRobin(req *http.Request,res http.ResponseWriter){
	targetServer:=lb.getNextRoundRobin()

	fmt.Printf("forwarding request to the server with address : %v\n",targetServer.Address())
	targetServer.Serve(res,req)
}

func (lb *LoadBalancer) serveProxyLeastConnection(req *http.Request,res http.ResponseWriter){
	targetServer := lb.getNextLeastConnection()

	defer func() {
		targetServer.DecrConnections()
	}()

	fmt.Printf("forwarding request to the server with address : %v\n",targetServer.Address())
	targetServer.Serve(res,req)
}



func main(){

	// make some simple servers using some existing urls and run it
	servers := []Server{
		newsimpleServer("https://postman-echo.com/delay/5"),
		newsimpleServer("https://postman-echo.com/delay/10"),
		newsimpleServer("https://www.facebook.com"),
		newsimpleServer("https://www.bing.com"),
		newsimpleServer("https://www.duckduckgo.com"),
	}

	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(res http.ResponseWriter, req *http.Request) {
		lb.serveProxyLeastConnection(req, res)
	}

	// register a proxy handler to handle all requests
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
