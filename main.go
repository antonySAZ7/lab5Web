package main

import (
	"database/sql"
	"fmt"
	"net"

	_ "modernc.org/sqlite"
)

func main() {

	db, _ := sql.Open("sqlite", "file:series.db")
	defer db.Close()

	ln, _ := net.Listen("tcp", ":8080")
	fmt.Println("Servidor en http://localhost:8080")

	for {
		conn, _ := ln.Accept()
		go handleRequest(conn, db)
	}
}
