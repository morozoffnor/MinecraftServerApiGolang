package main

import (
	"fmt"
	"github.com/gorcon/rcon"
	"log"
	"net/http"
)

// TODO: Make it universal function
func getDifficulty(w http.ResponseWriter, r *http.Request) {

	conn, err := rcon.Dial("testserver-mc-1:25575", "1234123")
	if err != nil {
		log.Print(err)
	}

	defer conn.Close()

	response, err := conn.Execute("time set day")
	if err != nil {
		log.Print(err)
	}

	fmt.Fprint(w, response)

}
