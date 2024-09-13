package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
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
	Pokedex  *map[string]Pokemon
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
		"catch": {
			name:        "catch",
			description: "Try to catch a specified Pokemon",
			callback:    commandCatch,
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

type Pokemon struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Cries          struct {
		Latest string `json:"latest"`
		Legacy string `json:"legacy"`
	} `json:"cries"`
	Forms []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndices []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	Height    int `json:"height"`
	HeldItems []struct {
		Item struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"item"`
		VersionDetails []struct {
			Rarity  int `json:"rarity"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"held_items"`
	ID                     int    `json:"id"`
	IsDefault              bool   `json:"is_default"`
	LocationAreaEncounters string `json:"location_area_encounters"`
	Moves                  []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
			VersionGroup struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Name          string `json:"name"`
	Order         int    `json:"order"`
	PastAbilities []any  `json:"past_abilities"`
	PastTypes     []struct {
		Generation struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"generation"`
		Types []struct {
			Slot int `json:"slot"`
			Type struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"type"`
		} `json:"types"`
	} `json:"past_types"`
	Species struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string `json:"back_default"`
		BackFemale       any    `json:"back_female"`
		BackShiny        string `json:"back_shiny"`
		BackShinyFemale  any    `json:"back_shiny_female"`
		FrontDefault     string `json:"front_default"`
		FrontFemale      any    `json:"front_female"`
		FrontShiny       string `json:"front_shiny"`
		FrontShinyFemale any    `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault string `json:"front_default"`
				FrontFemale  any    `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
			Showdown struct {
				BackDefault      string `json:"back_default"`
				BackFemale       any    `json:"back_female"`
				BackShiny        string `json:"back_shiny"`
				BackShinyFemale  any    `json:"back_shiny_female"`
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"showdown"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           string `json:"back_default"`
					BackShiny             string `json:"back_shiny"`
					BackShinyTransparent  string `json:"back_shiny_transparent"`
					BackTransparent       string `json:"back_transparent"`
					FrontDefault          string `json:"front_default"`
					FrontShiny            string `json:"front_shiny"`
					FrontShinyTransparent string `json:"front_shiny_transparent"`
					FrontTransparent      string `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      string `json:"back_default"`
						BackFemale       any    `json:"back_female"`
						BackShiny        string `json:"back_shiny"`
						BackShinyFemale  any    `json:"back_shiny_female"`
						FrontDefault     string `json:"front_default"`
						FrontFemale      any    `json:"front_female"`
						FrontShiny       string `json:"front_shiny"`
						FrontShinyFemale any    `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
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

func commandCatch(config *Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no pokemon given")
	}
	var pokemonArg string = args[0]
	urlToCall := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", pokemonArg)
	//check cache
	val, ok := config.Cache.Get(urlToCall)
	pokemon := Pokemon{}
	if ok {
		fmt.Println("found in cache")
		err := json.Unmarshal(val, &pokemon)
		if err != nil {
			return err
		}
		//try to catch
		fmt.Printf("Throwing a pokeball at %s... \n", pokemon.Name)
		randomChance := rand.Intn(1000)
		if randomChance > pokemon.BaseExperience {
			pokedex := *config.Pokedex
			pokedex[pokemon.Name] = pokemon
			fmt.Printf("%s was caught! \n", pokemon.Name)
			return nil
		} else {
			fmt.Printf("%s escaped! \n", pokemon.Name)
		}
		return nil
	}
	//make api call
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
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return err
	}
	//add to cache
	config.Cache.Add(urlToCall, body)
	//try to catch
	fmt.Printf("Throwing a pokeball at %s... \n", pokemon.Name)
	randomChance := rand.Intn(800)
	if randomChance > pokemon.BaseExperience {
		pokedex := *config.Pokedex
		pokedex[pokemon.Name] = pokemon
		fmt.Printf("%s was caught! \n", pokemon.Name)
		return nil
	} else {
		fmt.Printf("%s escaped! \n", pokemon.Name)
	}
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
	fmt.Println("initializing Pokedex..")
	pokedex := make(map[string]Pokemon)
	fmt.Println("intializing config..")
	config := &Config{
		Next:     nil,
		Previous: nil,
		Cache:    cache,
		Pokedex:  &pokedex,
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
