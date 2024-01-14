package main

import (
	"log"
	"sync"
)

func main() {
	api := Api{}
	err := api.InitApiServer()
	if err != nil {
		log.Fatal("Failed to initialize the app:", err)
	}

	grpc := Grpc{}
	grpc.InitGrpcServer()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		grpc.RunGrpcServer("localhost:8002")
	}()

	go func() {
		defer wg.Done()
		api.RunApiServer("localhost:8001")
	}()

	wg.Wait()
}
