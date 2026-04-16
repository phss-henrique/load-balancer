package main


import(
	"fmt"
	"net/http"
	"log"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"sync"
	"golang.org/x/time/rate"
	"net"

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

type visitor struct {
	limiter *rate.Limiter
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(5, 10)
		visitors[ip] = &visitor{limiter}
		return limiter
	}

	return v.limiter
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		limiter := getVisitor(ip)

		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests - Calma aí, amigão!", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
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

	finalHandler := rateLimitMiddleware(loadBalancer(pool))


	server := http.Server{
		Addr: ":8080",
		Handler: finalHandler,
	}

	fmt.Println("Load Balancer running on port 8080...")
	if err := server.ListenAndServe(); err != nil{
		log.Fatal(err)
	}

}