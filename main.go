package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Chrisk1905/pokedexcli/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
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
		"explore": {
			name:        "explore",
			description: "Displays the Pokemon in the given area",
			callback:    commandExplore,
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

type locationAreasExplore struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

func commandHelp(config *Config, args []string) error {
	commands := getCommands()
	for _, c := range commands {
		fmt.Printf("%s: %s \n", c.name, c.description)
	}
	return nil
}

func commandExit(config *Config, args []string) error {
	os.Exit(0) // Exits the program
	return nil // This line will never be reached
}

func commandMap(config *Config, args []string) error {
	//get URL
	var urlToCall string
	if config.Next != nil {
		urlToCall = *config.Next // Use the next URL if available
	} else {
		urlToCall = "https://pokeapi.co/api/v2/location-area/" // Default URL
	}
	//check cache
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

func commandMapb(config *Config, args []string) error {
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

func commandExplore(config *Config, args []string) error {
	// edge case
	if len(args) == 0 {
		return fmt.Errorf("no area specified")
	}
	var areaName string = args[0]
	urlToCall := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)
	// check cache
	val, ok := config.Cache.Get(urlToCall)
	locationAreasExplore := locationAreasExplore{}
	if ok {
		fmt.Println("Found in Cache")
		err := json.Unmarshal(val, &locationAreasExplore)
		if err != nil {
			return err
		}
		fmt.Printf("exploring %s", locationAreasExplore.Name)
		for _, pokemonEncounter := range locationAreasExplore.PokemonEncounters {
			fmt.Println(pokemonEncounter.Pokemon.Name)
		}
		return nil
	}
	// make api call
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
	err = json.Unmarshal(body, &locationAreasExplore)
	if err != nil {
		return err
	}
	for _, pokemonEncounter := range locationAreasExplore.PokemonEncounters {
		fmt.Println(pokemonEncounter.Pokemon.Name)
	}
	// add to cache
	config.Cache.Add(urlToCall, body)
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
			split_text := strings.Split(text, " ")
			args := split_text[1:]
			if command, exists := commands[split_text[0]]; exists {
				if err := command.callback(config, args); err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println("Unknown command:", text)
			}
		}
	}
}
