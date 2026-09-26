# 🌸 Akari

Юзербот для Telegram на Go. Быстрый, лёгкий, модульный.

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![mtgo](https://img.shields.io/badge/mtgo-v0.21.0-green)](https://github.com/mtgo-labs/mtgo)

---

## 📖 Содержание

- [Что это](#-что-это)
- [Возможности](#-возможности)
- [Требования](#-требования)
- [Установка](#-установка)
- [Первый запуск](#-первый-запуск)
- [Команды](#-команды)
- [Модули](#-модули)
  - [Формат модуля](#формат-модуля)
  - [Примеры](#примеры)
  - [Установка модулей](#установка-модулей)
- [Конфигурация](#-конфигурация)
- [Обновление](#-обновление)
- [Структура проекта](#-структура-проекта)
- [Безопасность](#-безопасность)
- [Лицензия](#-лицензия)

---

## 🌸 Что это

**Akari** — это юзербот (userbot) для Telegram, написанный на Go с использованием библиотеки [mtgo](https://github.com/mtgo-labs/mtgo).

Юзербот работает **от твоего аккаунта**, а не как отдельный бот. Это значит, что команды выполняются от твоего имени, и доступ к ним есть только у тебя (по `owner_id`).

**Основные принципы:**

- **Скорость** — Go + MTProto напрямую, минимум прослоек.
- **Легкость** — модульная архитектура, плагины на лету.
- **Безопасность** — команды только от владельца, изоляция пользовательских модулей через `os.Root`.

---

## ✨ Возможности

- 📦 **Модульная система** — системные модули встроены, пользовательские загружаются на лету
- 🔥 **Yaegi** — модули пишутся на Go и подхватываются без компиляции
- 🛡️ **Изоляция** — пользовательские модули работают в песочнице `userland/`
- ⚡ **Быстрый пинг** — измерение задержки MTProto
- 💻 **Терминал** — выполнение shell-команд прямо из Telegram
- 🔄 **Обновления** — `.update` проверяет GitHub, `.update -f` устанавливает
- 🔁 **Перезапуск** — `.restart` через `syscall.Exec` (работает в любом окружении)
- 🔐 **Только владелец** — команды доступны только пользователю с `owner_id`

---

## 📋 Требования

- **Go 1.22+** (для сборки)
- **Linux / macOS / Windows** (Linux рекомендован)
- **git** — для модуля `.update`
- **bash** — для модуля `.terminal`
- Аккаунт Telegram
- `api_id` и `api_hash` с [my.telegram.org](https://my.telegram.org/apps)

---

## 🚀 Установка

### VPS/VDS

<details>
  <summary><b>Ubuntu / Debian</b></summary>

  ```bash
  sudo apt update && sudo apt install -y git golang-go && \
  git clone https://github.com/uranfiz/Akari && \
  cd Akari && \
  go mod download && \
  go build -o akari ./cmd/akari && \
  ./akari
  ```
</details>

При первом запуске Akari спросит:

```
Akari is running!
Get api_id and api_hash at https://my.telegram.org/apps

api_id: 
api_hash: 
phone (+...): 
```

Введи свои данные:

- **api_id** и **api_hash** — получи на [my.telegram.org/apps](https://my.telegram.org/apps)
- **phone** — номер телефона в формате `+79871234567` (без пробелов и/или иных текстовых символов)

После этого:

```
connecting...
login: sending code to +79871234567...
Enter the code sent to +79871234567: 
login: 2FA required
Enter 2FA password (hint: ...): 
login: 2FA authenticated
Akari started as @your_username (ID 123456789)
```

- **Код** — придёт в Telegram
- **2FA-пароль** — если включена двухфакторка

**Создадутся файлы:**

- `config.yaml` — конфиг (НЕ коммить в git!)
- `akari.db` — сессия Telegram (НЕ коммить в git!)

Со второго запуска логин не потребуется — сессия сохранена в `akari.db`.

---

## 💬 Команды

Все команды работают во **всех чатах**, но выполняются **только от владельца** (`owner_id`).

### Системные

| Команда | Описание |
|---------|----------|
| `.ping` | Пинг юзербота + uptime |
| `.modules` | Список всех модулей |
| `.help <модуль>` | Справка по модулю |
| `.restart` | Перезапуск Akari |
| `.terminal <команда>` | Выполнить shell-команду |
| `.update` | Проверить обновления |
| `.update -f` | Установить обновления |
| `.update status` | Статус репозитория |

### Загрузчик модулей

| Команда | Описание |
|---------|----------|
| `.lm <url>` | Скачать модуль по raw-ссылке GitHub |
| `.lm` | Установить модуль из реплая (текст или файл) |
| `.lmlist` | Список установленных модулей |
| `.unlm <имя>` | Удалить модуль |
| `.lmreload` | Перезагрузить все модули |
| `.lmhelp` | Справка по установке модулей |

---

## 📦 Модули

### Формат модуля

Модуль — это обычный `.go`-файл, который **не компилируется** с основным бинарником. Akari подхватывает его через интерпретатор [yaegi](https://github.com/traefik/yaegi).

**Правила:**

1. Файл лежит в `userland/`.
2. Первая строка — `//go:build ignore` (чтобы `go build` его игнорировал).
3. Пакет — `package main`.
4. Каждая функция с **большой буквы** — это команда.
5. Функция принимает **строку** (аргументы) и возвращает **строку** (ответ).
6. Имя команды = имя функции в нижнем регистре.

### Примеры

#### Простейший модуль

Создай `userland/hello.go`:

```go
//go:build ignore

package main

func Hello(args string) string {
	if args == "" {
		return "Привет, мир!"
	}
	return "Привет, " + args + "!"
}
```

Что это даёт:

- Команда `.hello` → `Привет, мир!`
- Команда `.hello Вася` → `Привет, Вася!`

#### Несколько команд в одном модуле

```go
//go:build ignore

package main

func Hello(args string) string {
	if args == "" {
		return "Привет, мир!"
	}
	return "Привет, " + args + "!"
}

func Echo(args string) string {
	if args == "" {
		return "напиши что-нибудь"
	}
	return "echo: " + args
}

func Time(args string) string {
	return "Текущее время: " + now()
}

func now() string {
	return "12:34:56"
}
```

Команды: `.hello`, `.echo`, `.time`.

#### Модуль с проверкой и вычислениями

```go
//go:build ignore

package main

import (
	"fmt"
	"strings"
)

func Upper(args string) string {
	if args == "" {
		return "укажи текст: .upper <текст>"
	}
	return strings.ToUpper(args)
}

func Calc(args string) string {
	parts := strings.Fields(args)
	if len(parts) != 3 {
		return "использование: .calc <число> <+ - * /> <число>"
	}

	var a, b float64
	fmt.Sscanf(parts[0], "%f", &a)
	fmt.Sscanf(parts[2], "%f", &b)

	switch parts[1] {
	case "+":
		return fmt.Sprintf("%v", a+b)
	case "-":
		return fmt.Sprintf("%v", a-b)
	case "*":
		return fmt.Sprintf("%v", a*b)
	case "/":
		if b == 0 {
			return "деление на ноль"
		}
		return fmt.Sprintf("%v", a/b)
	}
	return "неизвестная операция"
}
```

Команды: `.upper привет` → `ПРИВЕТ`, `.calc 2 + 2` → `4`.

### Установка модулей

#### Из GitHub raw

Найди `.go`-файл на GitHub, открой его в режиме **Raw** (кнопка Raw), скопируй URL. Потом:

```
.lm https://raw.githubusercontent.com/user/repo/main/mymodule.go
```

Akari скачает, сохранит в `userland/mymodule.go` и зарегистрирует команды.

#### Из реплая

**Вариант 1 — текст:**

1. Отправь код модуля сообщением.
2. Ответь на это сообщение командой `.lm`.
3. Akari прочитает текст и установит.

**Вариант 2 — файл:**

1. Отправь `.go`-файл как документ.
2. Ответь на него командой `.lm`.
3. Akari скачает файл и установит.

### Управление модулями

```
.lmlist          # список установленных
.lmreload        # перезагрузить все
.unlm hello      # удалить модуль hello
```

**Важно:** после `.lm` команды модуля доступны сразу. Никакой перезагрузки не нужно.

---

## ⚙️ Конфигурация

Файл `config.yaml` создаётся при первом запуске. Пример:

```yaml
api_id: 12345678
api_hash: "abcdef1234567890abcdef1234567890"
session_name: akari
owner_id: 123456789
owner_phone: "+79871234567"
lang: ru
prefix: "."
modules_dir: userland
system_modules:
  - ping
  - restart
  - loader
  - modules
  - terminal
  - update
disabled_modules: []
github_repo: uranfiz/Akari
github_branch: main
repo_dir: "."
auto_update: true
log_level: info
```

| Поле | Описание |
|------|----------|
| `api_id` | ID приложения Telegram |
| `api_hash` | Hash приложения Telegram |
| `session_name` | Имя файла сессии (без `.db`) |
| `owner_id` | ID владельца (команды работают только для него) |
| `owner_phone` | Телефон владельца |
| `prefix` | Префикс команд (по умолчанию `.`) |
| `modules_dir` | Папка для пользовательских модулей |
| `system_modules` | Встроенные модули (нельзя удалить) |
| `github_repo` | Репозиторий для `.update` |
| `github_branch` | Ветка для `.update` |
| `repo_dir` | Путь к git-репозиторию |
| `auto_update` | Зарезервировано (сейчас не используется) |
| `log_level` | Уровень логов |

**`config.yaml` и `akari.db` — конфиденциальны. Никогда не коммить их в публичный репозиторий.**

---

## 🔄 Обновление

Akari умеет проверять и устанавливать обновления с GitHub.

### Проверка

```
.update
```

Если есть новые коммиты — покажет:

```
🚨 ОБНОВЛЕНИЕ AKARI
━━━━━━━━━━━━━━━━━━━━

⬆️ Доступна новая версия!

📍 Было:  cebbe58
📍 Стало: 432ec14
📦 Коммитов: 3

💬 Что нового:
  • 432ec14 fix ping module
  • ab1974f update README
  • 43581ce add new feature

━━━━━━━━━━━━━━━━━━━━
👉 .update -f — установить
```

Если обновлений нет:

```
✅ Обновлений нет.

Текущий коммит: cebbe58
```

### Установка

```
.update -f
```

Akari сделает:

1. `git pull origin main`
2. `go build -o akari ./cmd/akari`
3. `syscall.Exec` — перезапуск

Сообщение обновится на:

```
🌸 Akari успешно обновлена и перезапущена
```

### Статус

```
.update status
```

Покажет:

```
🌸 Akari · update · status
━━━━━━━━━━━━━━━━━━━━

📁 repo: /root/Akari
🌿 branch: main
📍 local: cebbe58
☁️ remote: cebbe58
```

---

## 🗂️ Структура проекта

```
akari/
├── cmd/
│   └── akari/
│       └── main.go         ← точка входа
│
├── core/                   ← ядро
│   ├── config.go           ← конфиг
│   ├── registry.go         ← реестр модулей
│   └── types.go            ← общие типы
│
├── api/                    ← обёртка над Telegram API
│   ├── dispatcher.go       ← роутинг команд
│   ├── peer.go             ← кэширование peer
│   └── utils.go            ← утилиты
│
├── builtin/                ← системные модули
│   ├── ping.go
│   ├── restart.go
│   ├── terminal.go
│   ├── update.go
│   ├── modules.go
│   ├── userloader.go       ← загрузчик модулей
│   └── sandbox.go          ← песочница yaegi
│
├── userland/               ← пользовательские модули
│   └── .gitkeep
│
├── assets/                 ← ассеты
├── scripts/                ← shell-скрипты
│
├── go.mod
├── go.sum
├── config.yaml             ← НЕ коммить
├── akari.db                ← НЕ коммить
└── LICENSE
```

---

## 🔒 Безопасность

### Что защищено

1. **Команды только от владельца.** `owner_id` в конфиге.
2. **Изоляция модулей.** Пользовательские модули работают через `os.Root`, который не позволяет выйти за пределы `userland/`.
3. **Ограниченный набор функций.** Модули могут вызывать только `strings.Upper` / `strings.Lower` из стандартной библиотеки (пока).
4. **Секреты не в git.**

### Что НЕ защищено

- **Модули могут возвращать любой текст** в ответ на команду. Не устанавливай модули из ненадёжных источников.
- **Модули могут делать HTTP-запросы** — если ты сам добавишь им такую возможность. Сейчас — нет.
- **Юзербот нарушает ToS Telegram.** Аккаунт могут забанить в любой момент. Это не баг Akari, это природа юзерботов.

### Рекомендации

- Не давай никому доступ к серверу.
- Не устанавливай модули, в которых не понимаешь код.
- Не используй юзербота для спама и массовых рассылок — забанят.
- Регулярно делай бэкап `config.yaml` и `akari.db`.

---

## 🛠️ Разработка

### Собрать

```bash
go build -o akari ./cmd/akari
```

### Запустить в фоне

```bash
nohup ./akari > akari.log 2>&1 &
```

### Остановить

```bash
pkill -f './akari'
```

### Запустить в tmux

```bash
tmux new -s akari
./akari
# Ctrl+B, D — отсоединиться
```

Вернуться:

```bash
tmux attach -t akari
```

### Добавить системный модуль

1. Создай файл в `builtin/`, например `builtin/mymodule.go`.
2. Пакет — `package builtin`.
3. В `init()` вызови `core.Register(&core.Module{...})`.

Пример:

```go
package builtin

import (
	"akari/api"
	"akari/core"

	"github.com/mtgo-labs/mtgo/telegram"
)

func init() {
	core.Register(&core.Module{
		Name:        "mymodule",
		Description: "Мой модуль",
		Commands: map[string]*core.Command{
			"mycmd": {
				Handler:     cmdMyCmd,
				Description: "Моя команда",
			},
		},
	})
}

func cmdMyCmd(ctx *telegram.Context) error {
	if ctx.Message == nil {
		return nil
	}
	chatID := ctx.Message.ChatID
	msgID := ctx.Message.ID
	api.CachePeer(ctx, chatID)

	_, err := ctx.Client.EditMessageText(ctx.Ctx, chatID, msgID, "Привет из mymodule")
	return err
}
```

---

## 📄 Лицензия

GNU Affero General Public License v3.0 (AGPL-3.0). Смотри [LICENSE](LICENSE).

**Кратко:**

- ✅ Используй, модифицируй, распространяй
- ✅ Используй в коммерческих целях
- ⚠️ Производные работы должны быть под той же лицензией
- ⚠️ **Указать авторство** — в исходниках, README или в описании проекта должна быть ссылка на [оригинальный репозиторий](https://github.com/uranfiz/Akari) и указание автора
- ⚠️ Если запускаешь как сетевой сервис — обязан опубликовать исходники

---

## 👥 Авторы

- **#dream** — [@devuranium](https://t.me/devuranium) — разработчик
- **アルチョム #ユエホスト** — [@familiarrrrrr](https://t.me/familiarrrrrr) — разработчик

Если ты внёс вклад в Akari — я добавлю тебя сюда.

---

## 🔗 Полезные ссылки

- [mtgo](https://github.com/mtgo-labs/mtgo) — библиотека MTProto
- [yaegi](https://github.com/traefik/yaegi) — интерпретатор Go
- [my.telegram.org](https://my.telegram.org/apps) — получить api_id/api_hash
- [AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html) — текст лицензии

---

<div align="center">

🌸 **Akari** — юзербот, который просто работает.

</div>
