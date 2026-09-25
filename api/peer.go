package api

import (
	"context"
	"log"
	"time"

	"github.com/mtgo-labs/mtgo/telegram"
	"github.com/mtgo-labs/mtgo/tg"
)

func CachePeer(ctx *telegram.Context, chatID int64) {
	if ctx.Update != nil {
		if chatID > 0 {
			if u, ok := ctx.Update.Users[chatID]; ok && u.AccessHash != 0 {
				ctx.Client.CachePeer(chatID, &tg.InputPeerUser{
					UserID:     chatID,
					AccessHash: u.AccessHash,
				})
				return
			}
		} else if ch, ok := ctx.Update.Chats[chatID]; ok {
			if channel, ok := ch.Raw.(*tg.Channel); ok {
				ctx.Client.CachePeer(chatID, &tg.InputPeerChannel{
					ChannelID:  channel.ID,
					AccessHash: channel.AccessHash,
				})
				return
			}
			if basicChat, ok := ch.Raw.(*tg.Chat); ok {
				ctx.Client.CachePeer(chatID, &tg.InputPeerChat{
					ChatID: basicChat.ID,
				})
				return
			}
		}
	}

	bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := ctx.Client.ResolvePeer(bgCtx, chatID); err != nil {
		log.Printf("resolve peer %d failed: %v", chatID, err)
	}
}
