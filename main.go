package main

import (
	"fmt"
	"goTogether/handler"
	"goTogether/middleware"
	"goTogether/sockets"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"golang.org/x/sys/unix"
)

func main() {
	fmt.Println("hello")
	godotenv.Load(".env")
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "8080"
	}

	// portString := "8008"

	var rLimit unix.Rlimit

	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatalf("Error getting rlimit: %v", err)
	}

	rLimit.Cur = rLimit.Max
	if err := unix.Setrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatalf("Error setting rlimit: %v", err)
	}
	var err error
	sockets.Epoller, err = sockets.MkEpoll()
	if err != nil {
		panic(err)
	}

	go sockets.Start()

	fmt.Println("File descriptor limit successfully increased.", rLimit.Cur)
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Only allow requests from React app
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	router.Get("/ws", sockets.SocketHandler)
	router.Get("/ws1", sockets.PollsocketHandler)

	V1router := chi.NewRouter()

	V1router.Get("/", handler.Testhandler)
	V1router.Get("/check", middleware.Verify(handler.Testhandler))

	router.Mount("/v1", V1router)
	// router.Mount("/debug/pprof", http.StripPrefix("/debug/pprof", http.DefaultServeMux))
	// Start pprof server on port 6060 in a separate goroutine
	// go func() {
	// 	log.Println("Starting pprof server on :6060")
	// 	log.Println(http.ListenAndServe(":6060", nil))
	// }()
	server := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server starting at port : %v", portString)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
