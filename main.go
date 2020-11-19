package main

import (
	"chatServer/server"
	"os"
)

func main() {
	db := "postgres://postgres:mysecretpassword@localhost:5432/users"
	s := server.NewServer(os.Stdout, "127.0.0.1:8080", db)
	s.Start()

}
