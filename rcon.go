package main

import (
	"github.com/gorcon/rcon"
	"log"
	"os"
)

var RCON_HOST = os.Getenv("RCON_HOST")
var RCON_PORT = os.Getenv("RCON_PORT")
var RCON_PASS = os.Getenv("RCON_PASS")

func sendRCONCommand(cmd string) (string, error) {
	conn, err := rcon.Dial(RCON_HOST+":"+RCON_PORT, RCON_PASS)
	if err != nil {
		log.Print(err)
	}
	defer conn.Close()
	response, err := conn.Execute(cmd)
	if err != nil {
		log.Print(err)
	}
	return response, nil
}

func stopServerRCON() (string, error) {
	response, err := sendRCONCommand("stop")
	if err != nil {
		log.Print("Error stopping the server: ", err)
		return "", err
	}
	return response, nil
}

func getDifficultyRCON() (string, error) {
	response, err := sendRCONCommand("difficulty")
	if err != nil {
		log.Print("Error while getting difficulty: ", err)
		return "", err
	}
	return response, err
}

func changeDifficultyRCON(difficulty string) (string, error) {
	response, err := sendRCONCommand("difficulty " + difficulty)
	if err != nil {
		log.Print("Error changing the difficulty: ", err)
		return "", err
	}
	return response, nil
}

func changeGameruleRCON(rule, value string) (string, error) {
	response, err := sendRCONCommand("gamerule " + rule + " " + value)
	if err != nil {
		log.Print("Error changing gamerule: ", err)
		return "", err
	}
	return response, nil
}
