package main

import (
	"sync"
)

func main() {
	app := App{}
	app.Initialise()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		app.RunGrpcServer("localhost:8002")
	}()

	go func() {
		defer wg.Done()
		app.RunApiServer("localhost:8001")
	}()

	wg.Wait()
}
