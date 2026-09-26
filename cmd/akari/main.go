package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"akari/api"
	"akari/builtin"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
	"github.com/mtgo-labs/storage"
	"github.com/mtgo-labs/storage/sqlite"
)

func main() {
	var err error

	cfg, err := core.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg == nil {
		runBanner()
		fmt.Println()
		fmt.Println("Get api_id and api_hash at https://my.telegram.org/apps")
		fmt.Println()
		cfg = core.DefaultConfig()
		cfg.APIID = int32(promptInt("api_id: "))
		cfg.APIHash = prompt("api_hash: ")
		cfg.OwnerPhone = prompt("phone (+...): ")
		if err := cfg.Save(); err != nil {
			log.Fatalf("save config: %v", err)
		}
		core.Cfg = cfg
	}

	if cfg.OwnerPhone == "" {
		cfg.OwnerPhone = prompt("phone (+...): ")
		if err := cfg.Save(); err != nil {
			log.Fatalf("save config: %v", err)
		}
	}

	ext, err := sqlite.Open(cfg.SessionName + ".db")
	if err != nil {
		log.Fatalf("open storage: %v", err)
	}
	defer ext.Close()

	deviceCfg := core.GetOrCreateDeviceProfile(cfg)

	client, err := telegram.NewClient(cfg.APIID, cfg.APIHash, &telegram.Config{
		PhoneNumber: cfg.OwnerPhone,
		SessionName: cfg.SessionName,
		SavePeers:   true,
		Storage:     storage.NewAdapter(ext),
		Device:      deviceCfg,
	})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}

	client.OnMessage(func(ctx *telegram.Context) {
		if err := api.Dispatch(ctx); err != nil {
			log.Printf("dispatch: %v", err)
		}
	})

	fmt.Println("connecting...")
	if err := client.Connect(0); err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Stop()

	me, err := client.GetMe(context.Background())
	if err != nil {
		log.Fatalf("get me: %v", err)
	}

	if cfg.OwnerID == 0 {
		cfg.OwnerID = me.ID
		log.Printf("owner_id set to %d (@%s)", me.ID, me.Username)
		if err := cfg.Save(); err != nil {
			log.Printf("save config: %v", err)
		}
	}

	log.Printf("Akari started as @%s (ID %d)", me.Username, me.ID)

	for name := range core.AllModules() {
		log.Printf("registered module: %s", name)
	}

	builtin.LoadAllModules()

	go builtin.NotifyRestartDone(client)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Akari stopped")
}

func prompt(label string) string {
	fmt.Print(label)
	s, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.TrimSpace(s)
}

func promptInt(label string) int {
	for {
		n, err := strconv.Atoi(prompt(label))
		if err == nil {
			return n
		}
		fmt.Println("invalid number")
	}
}

func runBanner() {
	const scriptPath = "scripts/banner.sh"

	info, err := os.Stat(scriptPath)
	if err != nil {
		return
	}

	if err := os.Chmod(scriptPath, info.Mode()|0111); err != nil {
		log.Printf("chmod banner script: %v", err)
		return
	}

	info, err = os.Stat(scriptPath)
	if err != nil {
		return
	}

	if info.Mode().Perm()&0111 == 0 {
		log.Printf("banner script is not executable: %s", scriptPath)
		return
	}

	cmd := exec.Command("./" + scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
