package builtin

import (
	"fmt"
	"log"
	"time"

	"akari/api"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
)

var startTime = time.Now()

func init() {
	core.Register(&core.Module{
		Name:        "ping",
		Description: "Показывает задержку юзербота и uptime",
		Commands: map[string]*core.Command{
			"ping": {
				Handler:     cmdPing,
				Description: "Пинг юзербота с аптаймом",
			},
		},
	})
}

func cmdPing(ctx *telegram.Context) error {
	start := time.Now()

	if ctx.Message == nil {
		return nil
	}

	chatID := ctx.Message.ChatID
	msgID := ctx.Message.ID

	if chatID == 0 || msgID == 0 {
		return nil
	}

	api.CachePeer(ctx, chatID)

	pingMs := float64(time.Since(start).Microseconds()) / 1000.0

	text := fmt.Sprintf(
		"🏓 Ping: %.3f ms\n⏱ Uptime: %s",
		pingMs,
		api.FormatUptime(time.Since(startTime)),
	)

	_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, text)
	if err != nil {
		log.Printf("ping edit error: %v", err)
	}
	return err
}
