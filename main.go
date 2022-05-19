package main

import (
	"fmt"
	"gomysql/db"
	"gomysql/restapi"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	go db.Database()
	go restapi.Api()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)

	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigs
		_ = sig
		done <- true
	}()

	<-done
	fmt.Println("Program closed")
}