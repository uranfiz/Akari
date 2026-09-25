package builtin

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"akari/api"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
)

const (
	termMaxOutput = 3200
	termTimeout   = 60 * time.Second
)

func init() {
	core.Register(&core.Module{
		Name:        "terminal",
		Description: "Выполнение shell-команд на сервере",
		Commands: map[string]*core.Command{
			"terminal": {
				Handler:     cmdTerminal,
				Description: "Выполнить shell-команду: .terminal <команда>",
			},
			"term": {
				Handler:     cmdTerminal,
				Description: "Алиас для .terminal",
			},
			"sh": {
				Handler:     cmdTerminal,
				Description: "Алиас для .terminal",
			},
		},
	})
}

func cmdTerminal(ctx *telegram.Context) error {
	if ctx.Message == nil {
		return nil
	}

	chatID := ctx.Message.ChatID
	msgID := ctx.Message.ID
	if chatID == 0 || msgID == 0 {
		return nil
	}

	api.CachePeer(ctx, chatID)

	args := api.CommandArgs(ctx)
	if len(args) == 0 {
		_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · terminal\n━━━━━━━━━━━━━━━━━━━━\n\nИспользование: .terminal <команда>")
		return err
	}

	command := strings.Join(args, " ")

	log.Printf("terminal: exec %q", command)

	runCtx, cancel := context.WithTimeout(context.Background(), termTimeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "bash", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if runCtx.Err() == context.DeadlineExceeded {
			exitCode = -1
		} else {
			exitCode = -2
		}
	}

	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	output = strings.TrimRight(output, "\n")

	if output == "" {
		output = "(пусто)"
	}

	truncated := false
	if len(output) > termMaxOutput {
		output = output[:termMaxOutput]
		truncated = true
	}

	statusEmoji := "✅"
	statusText := "ok"
	if exitCode == -1 {
		statusEmoji = "⏱"
		statusText = "timeout"
	} else if exitCode != 0 {
		statusEmoji = "❌"
		statusText = fmt.Sprintf("exit %d", exitCode)
	}

	var sb strings.Builder
	sb.WriteString("🌸 Akari · terminal\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString(fmt.Sprintf("$ %s\n\n", command))
	sb.WriteString(output)
	if truncated {
		sb.WriteString(fmt.Sprintf("\n\n... вывод обрезан (лимит %d символов)", termMaxOutput))
	}
	sb.WriteString("\n\n━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf("%s %s · %s", statusEmoji, statusText, elapsed.Round(time.Millisecond)))

	text := sb.String()
	if len(text) > 4096 {
		text = text[:4096]
	}

	_, editErr := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, text)
	if editErr != nil {
		log.Printf("terminal edit error: %v", editErr)
	}
	return editErr
}
