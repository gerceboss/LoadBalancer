## Simple Load Balancer in Go

This project implements a basic load balancer in Go. It distributes incoming HTTP requests across multiple backend servers, offering two load balancing strategies: Round-Robin and Least Connections. It leverages Go's built-in `http` package for server functionality and the `httputil` package's reverse proxy capability to forward requests efficiently.

## Architecture

![Project Architecture](public/ArchitectureMiniProject.jpg)


## Key Components:

1. **Server Interface :**

   - Defines fundamental methods for backend servers:
     - `IsAlive()`: Determines server availability.
     - `Address()`: Returns the server's address.
     - `Serve(res http.ResponseWriter, req *http.Request)`: Forwards requests to the server.
     - `GetConnections() int`: Retrieves the current number of active connections.
     - `IncrConnections()`, `DecrConnections()`: Increase/decrease the active connection count (thread-safe using `sync.Mutex`).

2. **simpleServer :**

   - Implements the `Server` interface, functioning as a reverse proxy.
   - Forwards HTTP requests to a designated backend server.
   - Tracks active connections using synchronization (`sync.Mutex`) for thread safety during connection count updates.

3. **LoadBalancer :**

   - Manages a pool of backend servers.
   - Distributes incoming requests based on chosen strategies:
     - **Round-Robin:**
       - Cycles through available servers sequentially.
       - Skips unavailable servers (checked by `IsAlive()`).
     - **Least Connections:**
       - Routes requests to the server with the lowest active connection count for balanced distribution, especially when servers have varying processing times.

4. **Reverse Proxy:**

   - Utilizes `httputil.ReverseProxy` to forward HTTP requests to the designated backend server.
   - Enables the load balancer to act as an intermediary between clients and backend servers.

## Load Balancing Strategies:

1. **Round-Robin :**

   - Maintains an internal counter (`currentIndex`) that traverses the server pool in a circular fashion.
   - For each request:
     - Increments the counter.
     - Queries server availability using `IsAlive()`.
     - If unavailable, skips to the next server.
     - If available, forwards the request to the server at the `currentIndex` position in the pool.

2. **Least Connections :**

   - Selects the server with the fewest active connections.
   - Iterates through the server pool:
     - For each server, checks `GetConnections()`.
     - Tracks the server with the minimum connection count and its index.
   - Forwards the request to the server with the least number of connections.

## How to Run:

1. Ensure Go is installed ([https://golang.org/doc/install](https://www.google.com/url?sa=E&source=gmail&q=https://golang.org/doc/install)).

2. Clone or download the project repository.

3. Open a terminal in the project directory.

4. Run the command:

   ```bash
   go run main.go
   ```
   Open a browser in incognito mode and hit the url ([http://localhost:8000](http://localhost:8000))

