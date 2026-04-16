package main


import(
	"fmt"
	"net/http"
	"log"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

)

type ServerPool struct{
	backends []*url.URL
	current uint64
}

func  (s *ServerPool) nextBackend() *url.URL{
	next := atomic.AddUint64(&s.current, uint64(1))

	index := next % uint64(len(s.backends))
	return s.backends[index]	
}

func loadBalancer(pool *ServerPool) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		backend := pool.nextBackend()
		fmt.Print("Forwarding request to backend: ", backend.Host, "\n")
		proxy := httputil.NewSingleHostReverseProxy(backend)
		proxy.ServeHTTP(w, r)

	}
}

func main(){
	backends := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}
	pool := &ServerPool{}
	for _, b := range backends{
		url, _ := url.Parse(b)
		pool.backends = append(pool.backends, url)
	}
	server := http.Server{
		Addr: ":8080",
		Handler: loadBalancer(pool),
	}

	fmt.Println("Load Balancer running on port 8080...")
	if err := server.ListenAndServe(); err != nil{
		log.Fatal(err)
	}

}