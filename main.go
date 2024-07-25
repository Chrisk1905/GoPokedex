package main

import (
	"bufio"
	"fmt"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
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
	}
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!\nUsage:")
	fmt.Println("help: Displays a help message")
	fmt.Println("exit: Exit the Pokedex")
	return nil
}

func commandExit() error {
	os.Exit(0) // Exits the program
	return nil // This line will never be reached
}
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	commands := getCommands()

	for {
		fmt.Print("pokedex > ")
		if scanner.Scan() {
			text := scanner.Text()
			if command, exists := commands[text]; exists {
				if err := command.callback(); err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println("Unknown command:", text)
			}
		}
	}
}
