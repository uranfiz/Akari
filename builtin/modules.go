package builtin

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"akari/api"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
)

func init() {
	core.Register(&core.Module{
		Name:        "modules",
		Description: "Список модулей и справка по ним",
		Commands: map[string]*core.Command{
			"modules": {
				Handler:     cmdModules,
				Description: "Список модулей или справка по конкретному модулю",
			},
			"help": {
				Handler:     cmdModules,
				Description: "Справка по модулю или по всем модулям",
			},
			"commands": {
				Handler:     cmdModules,
				Description: "Алиас для .help",
			},
		},
	})
}

func cmdModules(ctx *telegram.Context) error {
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

	var text string
	if len(args) > 0 {
		text = moduleHelp(args[0])
	} else {
		text = moduleList()
	}

	_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, text)
	if err != nil {
		log.Printf("modules edit error: %v", err)
	}
	return err
}

func moduleList() string {
	all := core.AllModules()

	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)

	var sysNames, userNames []string
	for _, name := range names {
		if core.Cfg.IsSystemModule(name) {
			sysNames = append(sysNames, name)
		} else {
			userNames = append(userNames, name)
		}
	}

	var sb strings.Builder
	sb.WriteString("🌸 Akari · modules\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	if len(sysNames) > 0 {
		sb.WriteString("\n⚙️ System\n")
		for _, name := range sysNames {
			sb.WriteString(formatModuleLine(name, all[name]))
		}
	}

	if len(userNames) > 0 {
		sb.WriteString("\n📦 User\n")
		for _, name := range userNames {
			sb.WriteString(formatModuleLine(name, all[name]))
		}
	}

	if len(sysNames) == 0 && len(userNames) == 0 {
		sb.WriteString("\n(нет загруженных модулей)\n")
	}

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString(fmt.Sprintf(
		"всего: %d · системных: %d · пользовательских: %d",
		len(names), len(sysNames), len(userNames),
	))

	return sb.String()
}

func moduleHelp(name string) string {
	m, ok := core.FindModule(name)
	if !ok {
		m, ok = core.FindModuleByCommand(name)
	}
	if !ok {
		return fmt.Sprintf("🌸 Akari · help\n━━━━━━━━━━━━━━━━━━━━\n\nМодуль «%s» не найден.", name)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🌸 Akari · %s\n", m.Name))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")

	if m.Description != "" {
		sb.WriteString("\n")
		sb.WriteString(m.Description)
		sb.WriteString("\n")
	}

	if core.Cfg.IsSystemModule(m.Name) {
		sb.WriteString("\n⚙️ Тип: системный\n")
	} else {
		sb.WriteString("\n📦 Тип: пользовательский\n")
	}

	if len(m.Commands) > 0 {
		cmds := make([]string, 0, len(m.Commands))
		for c := range m.Commands {
			cmds = append(cmds, c)
		}
		sort.Strings(cmds)

		sb.WriteString("\n💬 Команды\n")
		for _, c := range cmds {
			cmd := m.Commands[c]
			sb.WriteString(fmt.Sprintf("  • %s%s", core.Cfg.Prefix, c))
			if cmd.Description != "" {
				sb.WriteString(" — ")
				sb.WriteString(cmd.Description)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func formatModuleLine(name string, m *core.Module) string {
	cmds := make([]string, 0, len(m.Commands))
	for c := range m.Commands {
		cmds = append(cmds, core.Cfg.Prefix+c)
	}
	sort.Strings(cmds)

	if len(cmds) == 0 {
		return fmt.Sprintf("  • %s\n", name)
	}
	return fmt.Sprintf("  • %s — %s\n", name, strings.Join(cmds, ", "))
}
