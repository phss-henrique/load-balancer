package main

import(
	"fmt"
	"log"
	"net/http"
)

func startServer(port string){
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		fmt.Fprintf(w, "Backend answer coming from: %s\n", port)
	})
	fmt.Println("Backend rodando na porta", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil{
		log.Fatal(err)
	}
}
func main(){
	go startServer("8081")
	go startServer("8082")
	

	startServer("8083")
}