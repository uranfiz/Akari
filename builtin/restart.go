package builtin

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"akari/api"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
	"github.com/mtgo-labs/mtgo/tg"
)

const restartStateFile = "restart.json"

type restartState struct {
	ChatID     int64  `json:"chat_id"`
	MsgID      int32  `json:"msg_id"`
	PeerType   string `json:"peer_type"`
	AccessHash int64  `json:"access_hash,omitempty"`
	ChannelID  int64  `json:"channel_id,omitempty"`
}

func init() {
	core.Register(&core.Module{
		Name:        "restart",
		Description: "Перезапуск юзербота",
		Commands: map[string]*core.Command{
			"restart": {
				Handler:     cmdRestart,
				Description: "Перезапустить Akari",
			},
			"reboot": {
				Handler:     cmdRestart,
				Description: "Алиас для .restart",
			},
		},
	})
}

func cmdRestart(ctx *telegram.Context) error {
	if ctx.Message == nil {
		return nil
	}

	chatID := ctx.Message.ChatID
	msgID := ctx.Message.ID
	if chatID == 0 || msgID == 0 {
		return nil
	}

	api.CachePeer(ctx, chatID)

	state := restartState{ChatID: chatID, MsgID: msgID}

	if ctx.Update != nil {
		if chatID > 0 {
			if u, ok := ctx.Update.Users[chatID]; ok && u.AccessHash != 0 {
				state.PeerType = "user"
				state.AccessHash = u.AccessHash
			}
		} else if ch, ok := ctx.Update.Chats[chatID]; ok {
			if channel, ok := ch.Raw.(*tg.Channel); ok {
				state.PeerType = "channel"
				state.AccessHash = channel.AccessHash
				state.ChannelID = channel.ID
			} else if _, ok := ch.Raw.(*tg.Chat); ok {
				state.PeerType = "chat"
			}
		}
	}

	if data, err := json.Marshal(state); err == nil {
		if err := os.WriteFile(restartStateFile, data, 0o600); err != nil {
			log.Printf("restart: save state: %v", err)
		}
	}

	_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, "🌸 Akari · перезапуск...")
	if err != nil {
		log.Printf("restart edit error: %v", err)
	}

	go func() {
		time.Sleep(1 * time.Second)

		if isSystemdService() {
			log.Printf("restart: systemd service detected, exiting for auto-restart")
			os.Exit(0)
		}

		if runtime.GOOS == "windows" {
			log.Printf("restart: windows detected, exiting")
			os.Exit(0)
		}

		exe, err := os.Executable()
		if err != nil {
			log.Printf("restart: os.Executable: %v", err)
			os.Exit(1)
		}

		log.Printf("restart: syscall.Exec %s", exe)

		if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
			log.Printf("restart: syscall.Exec failed: %v", err)
			os.Exit(1)
		}
	}()

	return nil
}

func isSystemdService() bool {
	if os.Getenv("INVOCATION_ID") != "" {
		return true
	}

	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), ".service")
}

func NotifyRestartDone(client *telegram.Client) {
	data, err := os.ReadFile(restartStateFile)
	if err != nil {
		return
	}

	log.Printf("restart: notify state found: %s", string(data))

	_ = os.Remove(restartStateFile)

	var state restartState
	if err := json.Unmarshal(data, &state); err != nil {
		log.Printf("restart: parse state: %v", err)
		return
	}

	time.Sleep(3 * time.Second)

	switch state.PeerType {
	case "user":
		if state.AccessHash != 0 {
			client.CachePeer(state.ChatID, &tg.InputPeerUser{
				UserID:     state.ChatID,
				AccessHash: state.AccessHash,
			})
		}
	case "channel":
		if state.AccessHash != 0 {
			client.CachePeer(state.ChatID, &tg.InputPeerChannel{
				ChannelID:  state.ChannelID,
				AccessHash: state.AccessHash,
			})
		}
	case "chat":
		client.CachePeer(state.ChatID, &tg.InputPeerChat{
			ChatID: state.ChatID,
		})
	}

	_, err = client.EditMessageText(
		context.Background(),
		state.ChatID,
		state.MsgID,
		"🌸 Akari успешно обновлена и перезапущена",
	)
	if err != nil {
		log.Printf("restart: notify edit error: %v", err)
		return
	}
	log.Printf("restart: notify done")
}
