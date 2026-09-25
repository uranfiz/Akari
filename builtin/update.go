package builtin

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"akari/api"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
)

const updateCheckFreq = 5 * time.Minute

var updateLastNotified string

func init() {
	core.Register(&core.Module{
		Name:        "update",
		Description: "Проверка и установка обновлений с GitHub",
		Commands: map[string]*core.Command{
			"update": {
				Handler:     cmdUpdate,
				Description: "Проверить или установить обновления (.update -f)",
			},
			"upd": {
				Handler:     cmdUpdate,
				Description: "Алиас для .update",
			},
		},
	})
}

func cmdUpdate(ctx *telegram.Context) error {
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
	force := false
	statusOnly := false
	for _, a := range args {
		switch a {
		case "-f", "--force":
			force = true
		case "status":
			statusOnly = true
		}
	}

	repoDir := repoPath()

	if !isGitRepo(repoDir) {
		_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n❌ %s — не git-репозиторий.\n\nПроверь config.yaml → repo_dir", repoDir))
		return err
	}

	current, err := gitCurrentCommit(repoDir)
	if err != nil {
		log.Printf("update: current commit: %v", err)
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось получить текущий коммит: %v", err))
		return e
	}

	if err := gitFetch(repoDir); err != nil {
		log.Printf("update: fetch: %v", err)
	}

	latest, err := gitRemoteCommit(repoDir)
	if err != nil {
		log.Printf("update: remote commit: %v", err)
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось получить коммит с GitHub: %v", err))
		return e
	}

	if statusOnly {
		text := fmt.Sprintf(
			"🌸 Akari · update · status\n━━━━━━━━━━━━━━━━━━━━\n\n"+
				"📁 repo: %s\n"+
				"🌿 branch: %s\n"+
				"📍 local: %s\n"+
				"☁️ remote: %s",
			repoDir, core.Cfg.GitHubBranch, shortHash(current), shortHash(latest),
		)
		_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, text)
		return err
	}

	_, _ = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, "🌸 Akari · проверка обновлений...")

	if current == latest {
		_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n✅ Обновлений нет.\n\nТекущий коммит: %s", shortHash(current)))
		return err
	}

	commits, _ := gitCommitsBetween(repoDir, current, latest)

	var sb strings.Builder
	sb.WriteString("🌸 Akari · update\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n\n")
	sb.WriteString("📦 Доступно обновление\n\n")
	sb.WriteString(fmt.Sprintf("  текущий: %s\n", shortHash(current)))
	sb.WriteString(fmt.Sprintf("  новый:   %s\n", shortHash(latest)))
	sb.WriteString(fmt.Sprintf("  коммитов: %d\n", len(commits)))

	if len(commits) > 0 {
		sb.WriteString("\n💬 Изменения:\n")
		max := len(commits)
		if max > 10 {
			max = 10
		}
		for i := 0; i < max; i++ {
			sb.WriteString(fmt.Sprintf("  • %s\n", commits[i]))
		}
		if len(commits) > 10 {
			sb.WriteString(fmt.Sprintf("  ... и ещё %d\n", len(commits)-10))
		}
	}

	if !force {
		sb.WriteString("\n💡 Напиши .update -f, чтобы установить.")
		_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, sb.String())
		return err
	}

	sb.WriteString("\n⬇️ git pull...")
	_, _ = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, sb.String())

	if err := gitPull(repoDir); err != nil {
		log.Printf("update: pull: %v", err)
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n❌ git pull:\n%s", trimOutput(err.Error(), 800)))
		return e
	}

	_, _ = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
		"🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n⚙️ go build...")

	if err := goBuild(repoDir); err != nil {
		log.Printf("update: build: %v", err)
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n❌ go build:\n%s", trimOutput(err.Error(), 800)))
		return e
	}

	_, _ = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
		"🌸 Akari · update\n━━━━━━━━━━━━━━━━━━━━\n\n✅ Готово, перезапуск...")

	time.Sleep(1 * time.Second)

	exe, err := os.Executable()
	if err != nil {
		log.Printf("update: os.Executable: %v", err)
		return nil
	}

	log.Printf("update: exec %s", exe)

	if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
		log.Printf("update: syscall.Exec: %v", err)
	}

	return nil
}

func repoPath() string {
	if core.Cfg.RepoDir == "" {
		return "."
	}
	abs, err := filepath.Abs(core.Cfg.RepoDir)
	if err != nil {
		return core.Cfg.RepoDir
	}
	return abs
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

func gitCurrentCommit(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitRemoteCommit(dir string) (string, error) {
	ref := fmt.Sprintf("origin/%s", core.Cfg.GitHubBranch)
	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func gitFetch(dir string) error {
	cmd := exec.Command("git", "fetch", "origin", "--quiet")
	cmd.Dir = dir
	return cmd.Run()
}

func gitPull(dir string) error {
	cmd := exec.Command("git", "pull", "origin", core.Cfg.GitHubBranch)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}

func gitCommitsBetween(dir, from, to string) ([]string, error) {
	cmd := exec.Command("git", "log", "--oneline", "--no-decorate",
		fmt.Sprintf("%s..%s", from, to))
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	return lines, nil
}

func goBuild(dir string) error {
	cmd := exec.Command("go", "build", "-o", "akari", "./cmd/akari")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %s", err, string(out))
	}
	return nil
}

func shortHash(h string) string {
	if len(h) > 7 {
		return h[:7]
	}
	return h
}

func trimOutput(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

func PollUpdates(client *telegram.Client) {
	if core.Cfg.OwnerID == 0 {
		return
	}

	repoDir := repoPath()
	if !isGitRepo(repoDir) {
		log.Printf("update poller: %s не git-репо, отключён", repoDir)
		return
	}

	log.Printf("update poller: started, repo=%s", repoDir)

	ticker := time.NewTicker(updateCheckFreq)
	defer ticker.Stop()

	for range ticker.C {
		_ = gitFetch(repoDir)

		current, err := gitCurrentCommit(repoDir)
		if err != nil {
			continue
		}

		latest, err := gitRemoteCommit(repoDir)
		if err != nil {
			continue
		}

		if current == latest {
			continue
		}

		if updateLastNotified == latest {
			continue
		}

		updateLastNotified = latest

		commits, _ := gitCommitsBetween(repoDir, current, latest)

		var sb strings.Builder
		sb.WriteString("🌸 Akari · обновление\n")
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n\n")
		sb.WriteString("📦 Доступно обновление\n\n")
		sb.WriteString(fmt.Sprintf("  %s → %s\n", shortHash(current), shortHash(latest)))
		sb.WriteString(fmt.Sprintf("  коммитов: %d\n", len(commits)))

		if len(commits) > 0 {
			sb.WriteString("\n💬 Изменения:\n")
			max := len(commits)
			if max > 10 {
				max = 10
			}
			for i := 0; i < max; i++ {
				sb.WriteString(fmt.Sprintf("  • %s\n", commits[i]))
			}
			if len(commits) > 10 {
				sb.WriteString(fmt.Sprintf("  ... и ещё %d\n", len(commits)-10))
			}
		}

		sb.WriteString("\n💡 Напиши .update -f, чтобы установить.")

		_, err = client.SendMessage(context.Background(), core.Cfg.OwnerID, sb.String())
		if err != nil {
			log.Printf("update poller: send: %v", err)
		}
	}
}
