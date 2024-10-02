package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
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
}


type simpleServer struct{
	address string
	proxy *httputil.ReverseProxy
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

func (lb *LoadBalancer)nextAvailableServer() Server{

	//  go to the next till you get a alive one
	server:= lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive(){
		lb.roundRobinCount++
		server=lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *LoadBalancer) serveProxy(req *http.Request,res http.ResponseWriter){
	targetServer:=lb.nextAvailableServer()

	fmt.Printf("forwarding request to the server with address : %v\n",targetServer.Address())
	targetServer.Serve(res,req)
}


func main(){

	// make some simple servers using some existing urls and run it
	servers := []Server{
		newsimpleServer("https://www.facebook.com"),
		newsimpleServer("https://www.bing.com"),
		newsimpleServer("https://www.duckduckgo.com"),
	}

	lb := newLoadBalancer("8000", servers)
	handleRedirect := func(res http.ResponseWriter, req *http.Request) {
		lb.serveProxy(req, res)
	}

	// register a proxy handler to handle all requests
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("serving requests at 'localhost:%s'\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
