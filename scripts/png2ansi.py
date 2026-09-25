#!/usr/bin/env python3
"""
PNG -> ANSI-арт конвертер.
Выводит в stdout (или в файл) ANSI-коды.

python3 png2ansi.py lumox1.png --width 90 --out assets/banner.txt

Использование:
  python3 png2ansi.py image.png
  python3 png2ansi.py image.png > banner.txt
  python3 png2ansi.py image.png --width 80 --height 40
  python3 png2ansi.py image.png --out banner.txt
  python3 png2ansi.py image.png --bg 0,0,0
  python3 png2ansi.py image.png --halfblock off
  python3 png2ansi.py image.png --3d            # включить мягкое затемнение низа (3D-эффект)
"""

import sys
import argparse
from PIL import Image

TRANSPARENT = None
UPPER_HALF = "\u2580"  # ▀
LOWER_HALF = "\u2584"  # ▄
FULL_BLOCK = "\u2588"  # █


def _auto_darken(rgb, strength=0.55):
    """
    Автоматическое затемнение с сохранением тона.
    strength 0.0 = без изменений, 1.0 = почти чёрный.
    Даёт мягкий 3D-переход как в пиксель-арте.
    """
    r, g, b = rgb
    factor = 1.0 - strength * 0.65
    return (
        max(0, int(r * factor)),
        max(0, int(g * factor)),
        max(0, int(b * factor)),
    )


def _esc_bg(rgb):
    r, g, b = rgb
    return f"\033[48;2;{r};{g};{b}m"


def _esc_fg(rgb):
    r, g, b = rgb
    return f"\033[38;2;{r};{g};{b}m"


def load_rows(path, width=0, height=0, smooth=False, half_block=True):
    img = Image.open(path).convert("RGBA")
    w0, h0 = img.size

    # Для пиксель-арта критичен NEAREST (чёткие края).
    # LANCZOS размывает при апскейле — только через --smooth.
    resample = Image.LANCZOS if smooth else Image.NEAREST
    if width and height:
        img = img.resize((width, height * (2 if half_block else 1)), resample)
    elif width and width != w0:
        h = round(h0 * width / w0)
        img = img.resize((width, h), resample)
    elif height and height != h0:
        w = round(w0 * height / h0)
        img = img.resize((w, height), resample)

    w, h = img.size
    px = img.load()
    rows = []
    for y in range(h):
        row = []
        for x in range(w):
            r, g, b, a = px[x, y]
            row.append(None if a < 128 else (r, g, b))
        rows.append(row)
    return rows


def strip_margins(rows):
    """Убираем полностью прозрачные края, чтобы арт не плыл."""
    while rows and all(c is TRANSPARENT for c in rows[0]):
        rows.pop(0)
    while rows and all(c is TRANSPARENT for c in rows[-1]):
        rows.pop()
    if not rows:
        return rows
    while all(row[0] is TRANSPARENT for row in rows):
        rows = [row[1:] for row in rows]
    while all(row[-1] is TRANSPARENT for row in rows):
        rows = [row[:-1] for row in rows]
    return rows


def to_ansi(rows, bg=None, half_block=True, effect3d=False):
    """
    half_block=True: упаковываем ДВЕ строки пикселей в одну строку терминала.
      - верхний пиксель -> цвет фона (48;2)
      - нижний пиксель  -> цвет символа (38;2)
      - ▀ когда оба есть, ▄ когда только низ, пробел с фоном когда только верх.
    effect3d=True: вместо нижнего пикселя рисуем затемнённый верхний (эффект глубины).
    """
    out = []
    h = len(rows)
    w = max((len(r) for r in rows), default=0)
    rows = [r + [TRANSPARENT] * (w - len(r)) for r in rows]

    def resolve(c, is_top):
        if c is not TRANSPARENT:
            return c
        if bg is not None:
            return bg if is_top else _auto_darken(bg, strength=0.5)
        return None

    if half_block and h % 2 == 1:
        rows = rows + [[TRANSPARENT] * w]
        h += 1

    step = 2 if half_block else 1
    for y in range(0, h, step):
        parts = []
        cur_pair = None
        run = 0
        for x in range(w + 1):
            if x < w:
                top = resolve(rows[y][x], True)
                if half_block:
                    bot_src = rows[y + 1][x]
                    if effect3d:
                        bot = _auto_darken(top, strength=0.55) if top is not None else None
                    else:
                        bot = resolve(bot_src, False)
                else:
                    bot = None
                pair = (top, bot)
            else:
                pair = None  # маркер конца строки — сброс run
            if pair == cur_pair:
                run += 1
            else:
                if run:
                    t, b = cur_pair
                    if half_block:
                        if t is None and b is None:
                            parts.append("\033[49m\033[39m" + " " * run)
                        elif t is None:
                            parts.append("\033[49m" + _esc_fg(b) + LOWER_HALF * run)
                        elif b is None:
                            parts.append(_esc_bg(t) + " " * run)
                        elif t == b:
                            parts.append(_esc_fg(t) + FULL_BLOCK * run)
                        else:
                            parts.append(_esc_bg(t) + _esc_fg(b) + UPPER_HALF * run)
                    else:
                        if t is None:
                            parts.append("\033[49m" + " " * run)
                        else:
                            parts.append(_esc_bg(t) + FULL_BLOCK * run)
                cur_pair, run = pair, 1
        out.append("".join(parts) + "\033[0m")
    return "\n".join(out) + "\n"


def parse_color(s):
    """Принимает 'r,g,b' или '#rrggbb'. Возвращает (r,g,b) или None."""
    if not s:
        return None
    s = s.strip()
    if s.startswith("#"):
        s = s[1:]
        if len(s) == 3:
            s = "".join(c * 2 for c in s)
        if len(s) != 6:
            raise ValueError(f"неверный цвет: {s}")
        return tuple(int(s[i:i + 2], 16) for i in (0, 2, 4))
    parts = [p.strip() for p in s.split(",")]
    if len(parts) != 3:
        raise ValueError(f"неверный цвет: {s}")
    return tuple(int(p) for p in parts)


def main():
    ap = argparse.ArgumentParser(description="PNG -> ANSI арт")
    ap.add_argument("image", help="путь к PNG")
    ap.add_argument("--width", type=int, default=0, help="ширина в символах")
    ap.add_argument("--height", type=int, default=0, help="высота в строках картинки (строк терминала будет в 2 раза меньше)")
    ap.add_argument("--out", help="файл для сохранения (по умолчанию stdout)")
    ap.add_argument("--bg", default=None,
                    help="цвет фона для прозрачных участков, напр. 0,0,0 или #000000")
    ap.add_argument("--halfblock", default="on", choices=["on", "off"],
                    help="использовать half-block упаковку (по умолчанию on)")
    ap.add_argument("--3d", dest="effect3d", action="store_true",
                    help="эффект глубины: низ каждого пикселя затемняется")
    ap.add_argument("--smooth", action="store_true",
                    help="сглаженный ресайз (LANCZOS). По умолчанию NEAREST — чёткие пиксели")
    args = ap.parse_args()

    bg = parse_color(args.bg)
    rows = strip_margins(load_rows(args.image, args.width, args.height,
                                 smooth=args.smooth, half_block=(args.halfblock == "on")))
    text = to_ansi(
        rows,
        bg=bg,
        half_block=(args.halfblock == "on"),
        effect3d=args.effect3d,
    )

    if args.out:
        with open(args.out, "w", encoding="utf-8") as f:
            f.write(text)
        print(f"OK: сохранено в {args.out} — показать: cat {args.out}", file=sys.stderr)
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main()