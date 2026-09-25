package builtin

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"akari/api"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
	"github.com/mtgo-labs/mtgo/telegram/params"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const (
	modulesDirName  = "userland"
	maxModuleSize   = 256 * 1024
	downloadTimeout = 30 * time.Second
)

type LoadedModule struct {
	Name string
	Path string
}

var loadedModules = map[string]*LoadedModule{}

func init() {
	core.Register(&core.Module{
		Name:        "loader",
		Description: "Установка и управление пользовательскими модулями",
		Commands: map[string]*core.Command{
			"lm": {
				Handler:     cmdLoad,
				Description: "Установить модуль: .lm <url> или .lm в реплай на .go",
			},
			"unlm": {
				Handler:     cmdUnload,
				Description: "Удалить модуль: .unlm <имя>",
			},
			"lmlist": {
				Handler:     cmdLmList,
				Description: "Список установленных модулей",
			},
			"lmreload": {
				Handler:     cmdLmReload,
				Description: "Перезагрузить все модули из userland/",
			},
			"lmhelp": {
				Handler:     cmdLmHelp,
				Description: "Справка по установке модулей",
			},
		},
	})
}

func cmdLmHelp(ctx *telegram.Context) error {
	if ctx.Message == nil {
		return nil
	}
	chatID := ctx.Message.ChatID
	msgID := ctx.Message.ID
	if chatID == 0 || msgID == 0 {
		return nil
	}
	api.CachePeer(ctx, chatID)

	text := `🌸 Akari · loader
━━━━━━━━━━━━━━━━━━━━

  .lm <url>       — скачать .go по raw-ссылке GitHub
  .lm             — установить .go из реплая (текст или файл)
  .lmlist         — список установленных
  .unlm <имя>     — удалить модуль
  .lmreload       — перезагрузить все модули

Формат модуля:

package main

func Hello(args string) string {
    return "Привет, " + args + "!"
}

Каждая функция с большой буквы — это команда.
Имя команды = имя функции в нижнем регистре.

Модули хранятся в userland/ и изолированы.`

	_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, text,
		&params.EditMessage{ParseMode: params.HTML})
	return err
}

func cmdLoad(ctx *telegram.Context) error {
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

	log.Printf("loader: cmdLoad args=%v replyToID=%d", args, ctx.Message.ReplyToID)

	if len(args) == 0 {
		if ctx.Message.ReplyToID != 0 {
			return installFromReply(ctx, chatID, msgID)
		}
		return cmdLmHelp(ctx)
	}

	if strings.HasPrefix(args[0], "http") {
		return installFromURL(ctx, chatID, msgID, args[0])
	}

	return cmdLmHelp(ctx)
}

func cmdUnload(ctx *telegram.Context) error {
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
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Укажи имя: .unlm <имя>")
		return e
	}

	name := sanitizeModuleName(args[0])
	modulesDir := ensureModulesDir()

	root, err := os.OpenRoot(modulesDir)
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось открыть userland/")
		return e
	}
	defer root.Close()

	fileName := name + ".go"
	if err := root.Remove(fileName); err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось удалить: %v", err))
		return e
	}

	delete(loadedModules, name)

	_, err = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
		fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n✅ Модуль %s удалён", name))
	return err
}

func cmdLmList(ctx *telegram.Context) error {
	if ctx.Message == nil {
		return nil
	}

	chatID := ctx.Message.ChatID
	msgID := ctx.Message.ID
	if chatID == 0 || msgID == 0 {
		return nil
	}

	api.CachePeer(ctx, chatID)

	modulesDir := ensureModulesDir()

	root, err := os.OpenRoot(modulesDir)
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось открыть userland/")
		return e
	}
	defer root.Close()

	f, err := root.Open(".")
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось открыть userland/")
		return e
	}
	defer f.Close()

	entries, err := f.ReadDir(-1)
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось прочитать userland/")
		return e
	}

	var sb strings.Builder
	sb.WriteString("🌸 Akari · loader\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n\n")

	count := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		count++
		sb.WriteString(fmt.Sprintf("  • %s\n", strings.TrimSuffix(e.Name(), ".go")))
	}

	if count == 0 {
		sb.WriteString("(нет установленных модулей)\n")
	}
	sb.WriteString(fmt.Sprintf("\nвсего: %d", count))

	_, err = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, sb.String())
	return err
}

func cmdLmReload(ctx *telegram.Context) error {
	if ctx.Message == nil {
		return nil
	}

	chatID := ctx.Message.ChatID
	msgID := ctx.Message.ID
	if chatID == 0 || msgID == 0 {
		return nil
	}

	api.CachePeer(ctx, chatID)

	modulesDir := ensureModulesDir()

	root, err := os.OpenRoot(modulesDir)
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось открыть userland/")
		return e
	}
	defer root.Close()

	f, err := root.Open(".")
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось открыть userland/")
		return e
	}
	defer f.Close()

	entries, err := f.ReadDir(-1)
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось прочитать userland/")
		return e
	}

	loaded := 0
	failed := 0

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".go")
		if _, err := loadUserModule(filepath.Join(modulesDir, e.Name()), name); err != nil {
			log.Printf("loader: reload %s: %v", name, err)
			failed++
			continue
		}
		loadedModules[name] = &LoadedModule{Name: name, Path: e.Name()}
		loaded++
	}

	_, err = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
		fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n✅ Загружено: %d\n❌ Ошибок: %d", loaded, failed))
	return err
}

func installFromURL(ctx *telegram.Context, chatID int64, msgID int32, url string) error {
	name := moduleNameFromURL(url)

	_, _ = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
		fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n⬇️ Скачиваю %s...", name))

	code, err := downloadModule(url)
	if err != nil {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Ошибка: %v", err))
		return e
	}

	return saveAndLoad(ctx, chatID, msgID, name, code)
}

func installFromReply(ctx *telegram.Context, chatID int64, msgID int32) error {
	log.Printf("loader: installFromReply chatID=%d msgID=%d replyToID=%d",
		chatID, msgID, ctx.Message.ReplyToID)

	if ctx.Message.ReplyToID == 0 {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Ответь командой на сообщение с .go-кодом или файлом")
		return e
	}

	replyID := ctx.Message.ReplyToID

	msgs, err := ctx.Client.GetMessages(ctx.Ctx, chatID, []int32{replyID})
	log.Printf("loader: GetMessages returned %d msgs, err=%v", len(msgs), err)

	if err != nil || len(msgs) == 0 {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось получить сообщение из реплая")
		return e
	}

	reply := msgs[0]

	log.Printf("loader: reply.Text=%q DocumentNil=%v",
		reply.Text, reply.Document == nil)

	name := fmt.Sprintf("module_%d", time.Now().Unix())
	code := ""

	if reply.Document != nil {
		data, err := reply.Download()
		if err != nil {
			_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
				fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось скачать файл: %v", err))
			return e
		}

		if len(data) > maxModuleSize {
			_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
				"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Файл слишком большой")
			return e
		}

		code = string(data)

		if reply.Document.FileName != "" {
			baseName := strings.TrimSuffix(reply.Document.FileName, ".go")
			cleanName := sanitizeModuleName(baseName)
			if cleanName != "" {
				name = cleanName
			}
		}
	} else if reply.Text != "" {
		code = reply.Text
	} else {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ В реплае нет текста и нет файла")
		return e
	}

	return saveAndLoad(ctx, chatID, msgID, name, code)
}

func saveAndLoad(ctx *telegram.Context, chatID int64, msgID int32, name, code string) error {
	cleanName := sanitizeModuleName(name)
	if cleanName == "" {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Недопустимое имя модуля")
		return e
	}

	if len(code) > maxModuleSize {
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Модуль слишком большой")
		return e
	}

	modulesDir := ensureModulesDir()

	root, err := os.OpenRoot(modulesDir)
	if err != nil {
		log.Printf("loader: open root: %v", err)
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			"🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Не удалось открыть userland/")
		return e
	}
	defer root.Close()

	fileName := cleanName + ".go"
	if err := root.WriteFile(fileName, []byte(code), 0o600); err != nil {
		log.Printf("loader: write: %v", err)
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Ошибка записи: %v", err))
		return e
	}

	count, err := loadUserModule(filepath.Join(modulesDir, fileName), cleanName)
	if err != nil {
		_ = root.Remove(fileName)
		log.Printf("loader: load %s: %v", cleanName, err)
		_, e := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
			fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n❌ Ошибка загрузки: %v", err))
		return e
	}

	loadedModules[cleanName] = &LoadedModule{Name: cleanName, Path: fileName}

	_, err = ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID,
		fmt.Sprintf("🌸 Akari · loader\n━━━━━━━━━━━━━━━━━━━━\n\n✅ Модуль %s установлен (%d команд)", cleanName, count))
	return err
}

func loadUserModule(path, name string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	sandbox := newSandbox()

	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)
	i.Use(sandbox.Symbols())

	log.Printf("loader: loadUserModule name=%s, len=%d", name, len(data))

	if _, err := i.Eval(string(data)); err != nil {
		log.Printf("loader: eval error: %v", err)
		return 0, fmt.Errorf("eval: %w", err)
	}

	log.Printf("loader: eval ok")

	count := 0

	moduleName := name

	moduleFuncs, err := scanModuleFunctions(i, name)
	if err != nil {
		return 0, err
	}

	for _, fn := range moduleFuncs {
		cmdName := strings.ToLower(fn.name)
		fnValue := fn.value

		wrapped := makeUserFuncHandler(cmdName, fnValue)

		existing, ok := core.AllModules()[moduleName]
		if ok {
			existing.Commands[cmdName] = &core.Command{
				Handler:     wrapped,
				Description: "",
			}
		} else {
			core.Register(&core.Module{
				Name:        moduleName,
				Description: "Пользовательский модуль",
				Commands: map[string]*core.Command{
					cmdName: {
						Handler:     wrapped,
						Description: "",
					},
				},
			})
		}
		count++
	}

	return count, nil
}

type moduleFunc struct {
	name  string
	value reflect.Value
}

func scanModuleFunctions(i *interp.Interpreter, moduleName string) ([]moduleFunc, error) {
	var result []moduleFunc

	names := []string{
		"Hello", "Echo", "Test", "Ping", "Info", "Start", "Help", "Run",
	}

	for _, name := range names {
		val, err := i.Eval(name)
		if err != nil {
			continue
		}

		v := val
		if !v.IsValid() {
			continue
		}

		if v.Kind() != reflect.Func {
			continue
		}

		if v.Type().NumIn() != 1 || v.Type().NumOut() != 1 {
			continue
		}

		if v.Type().In(0).Kind() != reflect.String {
			continue
		}

		if v.Type().Out(0).Kind() != reflect.String {
			continue
		}

		result = append(result, moduleFunc{name: name, value: v})
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid command functions found (need func(string) string)")
	}

	return result, nil
}

func makeUserFuncHandler(cmdName string, fn reflect.Value) core.CommandHandler {
	return func(ctx *telegram.Context) error {
		if ctx.Message == nil {
			return nil
		}
		chatID := ctx.Message.ChatID
		msgID := ctx.Message.ID
		if chatID == 0 || msgID == 0 {
			return nil
		}
		api.CachePeer(ctx, chatID)

		args := strings.Join(api.CommandArgs(ctx), " ")

		results := fn.Call([]reflect.Value{reflect.ValueOf(args)})
		if len(results) == 0 {
			return nil
		}

		reply := results[0].String()
		if reply == "" {
			return nil
		}

		_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, reply)
		return err
	}
}

func downloadModule(url string) (string, error) {
	client := &http.Client{Timeout: downloadTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxModuleSize+1))
	if err != nil {
		return "", err
	}

	if len(body) > maxModuleSize {
		return "", fmt.Errorf("файл больше %d байт", maxModuleSize)
	}

	return string(body), nil
}

func moduleNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]
	name = strings.TrimSuffix(name, ".go")
	if name == "" {
		name = fmt.Sprintf("module_%d", time.Now().Unix())
	}
	return name
}

func sanitizeModuleName(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSuffix(name, ".go")

	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func ensureModulesDir() string {
	dir := modulesDirName
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

func LoadAllModules() {
	modulesDir := ensureModulesDir()

	root, err := os.OpenRoot(modulesDir)
	if err != nil {
		log.Printf("loader: startup: open userland/: %v", err)
		return
	}
	defer root.Close()

	f, err := root.Open(".")
	if err != nil {
		log.Printf("loader: startup: open userland/: %v", err)
		return
	}
	defer f.Close()

	entries, err := f.ReadDir(-1)
	if err != nil {
		log.Printf("loader: startup: read userland/: %v", err)
		return
	}

	loaded := 0
	failed := 0

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".go")
		if _, err := loadUserModule(filepath.Join(modulesDir, e.Name()), name); err != nil {
			log.Printf("loader: startup: %s: %v", name, err)
			failed++
			continue
		}
		loadedModules[name] = &LoadedModule{Name: name, Path: e.Name()}
		loaded++
	}

	log.Printf("loader: startup: loaded %d modules, %d failed", loaded, failed)
}
