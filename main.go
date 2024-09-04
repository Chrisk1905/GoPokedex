package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Chrisk1905/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type Config struct {
	Next     *string // Pointer to handle absence of a next URL
	Previous *string // Pointer to handle absence of a previous URL
	Cache    *pokecache.Cache
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
	//get URL
	var urlToCall string
	if config.Next != nil {
		urlToCall = *config.Next // Use the next URL if available
	} else {
		urlToCall = "https://pokeapi.co/api/v2/location-area/" // Default URL
	}
	//check cache first
	val, ok := config.Cache.Get(urlToCall)
	locationArea := locationAreas{}
	if ok {
		fmt.Println("found in cache")
		err := json.Unmarshal(val, &locationArea)
		if err != nil {
			return err
		}
		for _, area := range locationArea.Results {
			fmt.Println(area.Name)
		}
		config.Next = &locationArea.Next
		config.Previous = &locationArea.Previous
		return nil
	}
	//http get call
	res, err := http.Get(urlToCall)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &locationArea)
	if err != nil {
		return err
	}
	for _, area := range locationArea.Results {
		fmt.Println(area.Name)
	}
	config.Next = &locationArea.Next
	config.Previous = &locationArea.Previous
	//add result to cache
	config.Cache.Add(urlToCall, body)
	return nil
}

func commandMapb(config *Config) error {
	if config.Previous == nil || *config.Previous == "" {
		return errors.New("no previous page")
	}
	//check cache first
	val, ok := config.Cache.Get(*config.Previous)
	locationArea := locationAreas{}
	if ok {
		fmt.Println("found in cache")
		err := json.Unmarshal(val, &locationArea)
		if err != nil {
			return err
		}
		for _, area := range locationArea.Results {
			fmt.Println(area.Name)
		}
		config.Next = &locationArea.Next
		config.Previous = &locationArea.Previous
		return nil
	}
	res, err := http.Get(*config.Previous)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		return fmt.Errorf("response failed with status code: %d and\nbody: %s", res.StatusCode, body)
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &locationArea)
	if err != nil {
		return err
	}
	for _, area := range locationArea.Results {
		fmt.Println(area.Name)
	}
	config.Next = &locationArea.Next
	config.Previous = &locationArea.Previous
	//add to cache
	config.Cache.Add(*config.Previous, body)
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()
	fmt.Println("creating cleanInterval..")
	cleanInterval, err := time.ParseDuration("1m")
	if err != nil {
		fmt.Printf("error creating time.ParseDuration %v", err)
	}
	fmt.Println("initializing cache..")
	cache := pokecache.NewCache(cleanInterval)
	fmt.Println("intializing config..")
	config := &Config{
		Next:     nil,
		Previous: nil,
		Cache:    cache,
	}
	fmt.Println("starting REPL..")
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
