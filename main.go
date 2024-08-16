package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type Config struct {
	Next     *string // Pointer to handle absence of a next URL
	Previous *string // Pointer to handle absence of a previous URL
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas in the Pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapb,
		},
	}
}

type locationAreas struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func commandHelp(config *Config) error {
	commands := getCommands()
	for _, c := range commands {
		fmt.Printf("%s: %s \n", c.name, c.description)
	}
	return nil
}

func commandExit(config *Config) error {
	os.Exit(0) // Exits the program
	return nil // This line will never be reached
}

func commandMap(config *Config) error {
	var urlToCall string

	if config.Next != nil {
		urlToCall = *config.Next // Use the next URL if available
	} else {
		urlToCall = "https://pokeapi.co/api/v2/location-area/" // Default URL
	}

	res, err := http.Get(urlToCall)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	locationArea := locationAreas{}
	err = json.Unmarshal(body, &locationArea)
	if err != nil {
		fmt.Println(err)
	}
	for _, area := range locationArea.Results {
		fmt.Println(area.Name)
	}
	config.Next = &locationArea.Next
	config.Previous = &locationArea.Previous
	return nil
}

func commandMapb(config *Config) error {
	if config.Previous == nil || *config.Previous == "" {
		return errors.New("no previous page")
	}
	res, err := http.Get(*config.Previous)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	locationArea := locationAreas{}
	err = json.Unmarshal(body, &locationArea)
	if err != nil {
		fmt.Println(err)
	}
	for _, area := range locationArea.Results {
		fmt.Println(area.Name)
	}
	config.Next = &locationArea.Next
	config.Previous = &locationArea.Previous
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	// Initialize Config
	config := &Config{ // using an address to make a pointer
		Next:     nil, // No next URL at the start
		Previous: nil, // No previous URL at the start
	}

	for {
		fmt.Print("pokedex > ")
		if scanner.Scan() {
			text := scanner.Text()
			if command, exists := commands[text]; exists {
				if err := command.callback(config); err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println("Unknown command:", text)
			}
		}
	}
}
