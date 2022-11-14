package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/oklog/oklog/pkg/group"
)

func main() {
	fmt.Println("Tic - Toe Game Started")

	// mux middleware
	var mwf []mux.MiddlewareFunc

	//router
	httpRouter := mux.NewRouter().StrictSlash(false)
	httpRouter.Use(mwf...)


	// Health Check
	httpRouter.PathPrefix("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("Successfully Health Pass")
		w.WriteHeader(http.StatusOK)
		
	})

	var server group.Group
	{
		httpServer := &http.Server{
			Addr:    ":8080",
			Handler: httpRouter,
		}

		server.Add(func() error {
			fmt.Printf("Server is start running on port: %d", 8080)
			return httpServer.ListenAndServe()

		}, func(err error) {

			// write code here for graceful shutDown

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			defer cancel()
			httpServer.Shutdown(ctx)

		})
	}
	// interuption handling
	{
		cancelInterrupt := make(chan struct{})

		server.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)

			select {
			case sig := <-c:
				return fmt.Errorf("received signal: %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(err error) {
			close(cancelInterrupt)
		})
	}

	fmt.Printf("Exiting....... Error: %v\n", server.Run())

}
