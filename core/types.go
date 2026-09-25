package core

import "github.com/mtgo-labs/mtgo/telegram"

type CommandHandler func(ctx *telegram.Context) error

type Command struct {
	Handler     CommandHandler
	Description string
}

type Module struct {
	Name        string
	Description string
	Commands    map[string]*Command
}
