package api

import (
	"strings"

	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
)

func Dispatch(ctx *telegram.Context) error {
	if ctx.Message == nil {
		return nil
	}

	senderID := ctx.Message.FromID
	if senderID == 0 {
		if s := ctx.Sender(); s != nil {
			senderID = s.ID
		}
	}

	if senderID == 0 {
		return nil
	}

	if senderID != core.Cfg.OwnerID {
		return nil
	}

	text := ctx.Message.Text
	if !strings.HasPrefix(text, core.Cfg.Prefix) {
		return nil
	}

	parts := strings.Fields(text[len(core.Cfg.Prefix):])
	if len(parts) == 0 {
		return nil
	}

	cmdName := parts[0]
	args := parts[1:]

	handler, ok := core.FindCommand(cmdName)
	if !ok {
		return nil
	}

	if ctx.PluginData == nil {
		ctx.PluginData = map[string]any{}
	}
	ctx.PluginData["args"] = args

	return handler(ctx)
}

func CommandArgs(ctx *telegram.Context) []string {
	if ctx.PluginData == nil {
		return nil
	}
	if v, ok := ctx.PluginData["args"]; ok {
		if a, ok := v.([]string); ok {
			return a
		}
	}
	return nil
}
