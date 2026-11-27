#!/usr/bin/env python3
"""Render the Tr product intro.

The video is drawn with Pillow frame by frame and piped to ffmpeg. It does not
use screen recording, so the result is deterministic and never exposes a real
terminal path or API key.

Usage:
    python video/render.py --still 5.5 --out Output/tr-preview.png
    python video/render.py --all
"""
from __future__ import annotations

import argparse
import math
import os
import subprocess
import sys
from dataclasses import dataclass
from functools import lru_cache

import numpy as np
from PIL import Image, ImageDraw, ImageFilter, ImageFont

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ASSETS = os.path.join(ROOT, "assets")
OUTPUT = os.path.join(ROOT, "Output")
VIDEO_DIR = os.path.dirname(os.path.abspath(__file__))

FPS = 30
DURATION = 60.0

# Palette ---------------------------------------------------------------
BG_TOP = (6, 10, 18)
BG_BOTTOM = (13, 22, 38)
TEXT = (236, 243, 255)
MUTED = (139, 157, 184)
DIM = (91, 108, 134)
ACCENT = (54, 211, 222)
ACCENT_2 = (91, 133, 255)
GREEN = (82, 220, 153)
AMBER = (255, 194, 102)
RED = (255, 108, 112)
PANEL = (13, 21, 36)
PANEL_2 = (18, 28, 47)
BORDER = (43, 61, 88)

FONT_ZH = os.path.join(os.environ.get("WINDIR", r"C:\Windows"), "Fonts", "msyh.ttc")
FONT_ZH_B = os.path.join(os.environ.get("WINDIR", r"C:\Windows"), "Fonts", "msyhbd.ttc")
FONT_MONO = os.path.join(os.environ.get("WINDIR", r"C:\Windows"), "Fonts", "consola.ttf")
FONT_MONO_B = os.path.join(os.environ.get("WINDIR", r"C:\Windows"), "Fonts", "consolab.ttf")
FONT_EN_B = os.path.join(os.environ.get("WINDIR", r"C:\Windows"), "Fonts", "segoeuib.ttf")


def ffmpeg_path() -> str:
    try:
        import imageio_ffmpeg

        return imageio_ffmpeg.get_ffmpeg_exe()
    except Exception:
        return os.environ.get("FFMPEG", "ffmpeg")


@lru_cache(maxsize=512)
def f_zh(size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(FONT_ZH, size)


@lru_cache(maxsize=512)
def f_zhb(size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(FONT_ZH_B, size)


@lru_cache(maxsize=512)
def f_mono(size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(FONT_MONO, size)


@lru_cache(maxsize=512)
def f_monob(size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(FONT_MONO_B, size)


@lru_cache(maxsize=64)
def f_enb(size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(FONT_EN_B, size)


def clamp(v: float, lo: float = 0.0, hi: float = 1.0) -> float:
    return max(lo, min(hi, v))


def ease_out(p: float) -> float:
    p = clamp(p)
    return 1 - (1 - p) ** 3


def ease_in_out(p: float) -> float:
    p = clamp(p)
    return p * p * (3 - 2 * p)


def lerp(a: float, b: float, p: float) -> float:
    return a + (b - a) * p


def a(c, alpha: int):
    return (c[0], c[1], c[2], int(clamp(alpha, 0, 255)))


def text(draw: ImageDraw.ImageDraw, xy, s: str, font, fill=TEXT, anchor="la", alpha=255):
    draw.text(xy, s, font=font, fill=a(fill, alpha), anchor=anchor)


def text_w(draw: ImageDraw.ImageDraw, s: str, font) -> float:
    return draw.textlength(s, font=font)


def round_rect(draw, box, radius, fill=None, outline=None, width=1):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def shadow_rect(draw, box, radius, spread=10):
    x0, y0, x1, y1 = box
    for off, alpha in ((spread, 55), (spread * 0.55, 65), (3, 75)):
        round_rect(
            draw,
            (x0 - off * 0.15, y0 + off * 0.45, x1 + off * 0.15, y1 + off * 0.55),
            radius,
            fill=(0, 0, 0, int(alpha)),
        )


@dataclass
class Ctx:
    w: int
    h: int
    portrait: bool

    @property
    def land(self) -> bool:
        return not self.portrait

    @property
    def cx(self) -> int:
        return self.w // 2

    @property
    def caption_y(self) -> int:
        return self.h - (170 if self.land else 420)

    @property
    def sub_y(self) -> int:
        return self.caption_y + (62 if self.land else 58)

    def fs(self, n: float) -> int:
        return max(11, int(round(n * (1.0 if self.land else 0.92))))


# Background ------------------------------------------------------------
def make_bg(w: int, h: int, portrait: bool) -> Image.Image:
    yy, xx = np.mgrid[0:h, 0:w].astype(np.float32)
    p = (yy / max(1, h - 1))[:, :, None]
    top = np.array(BG_TOP, dtype=np.float32)
    bot = np.array(BG_BOTTOM, dtype=np.float32)
    base = top + (bot - top) * p

    def glow(cx, cy, sigma, color, strength):
        d2 = (xx - cx) ** 2 + (yy - cy) ** 2
        g = np.exp(-d2 / (2 * sigma * sigma)) * strength
        return g[:, :, None] * np.array(color, dtype=np.float32)

    base += glow(w * 0.18, h * 0.10, w * 0.32, (31, 117, 145), 0.36)
    base += glow(w * 0.88, h * 0.84, w * 0.36, (45, 52, 140), 0.32)
    base += glow(w * 0.52, h * 0.50, w * 0.55, (16, 32, 58), 0.24)

    # subtle noise
    rng = np.random.default_rng(7)
    base += rng.normal(0, 1.8, size=(h, w, 1))
    img = Image.fromarray(np.clip(base, 0, 255).astype(np.uint8), "RGB").convert("RGBA")

    overlay = Image.new("RGBA", (w, h), (0, 0, 0, 0))
    d = ImageDraw.Draw(overlay)
    step = 64 if not portrait else 54
    for x in range(0, w, step):
        d.line((x, 0, x, h), fill=(120, 170, 220, 8), width=1)
    for y in range(0, h, step):
        d.line((0, y, w, y), fill=(120, 170, 220, 8), width=1)

    # vignette
    vig = Image.new("L", (w, h), 0)
    dv = ImageDraw.Draw(vig)
    dv.ellipse((-w * 0.25, -h * 0.25, w * 1.25, h * 1.25), fill=255)
    vig = vig.filter(ImageFilter.GaussianBlur(max(40, int(min(w, h) * 0.08))))
    dark = Image.new("RGBA", (w, h), (0, 0, 0, 255))
    dark.putalpha(Image.eval(vig, lambda v: int((255 - v) * 0.55)))
    overlay = Image.alpha_composite(overlay, dark)
    return Image.alpha_composite(img, overlay).convert("RGB")


# Logo ------------------------------------------------------------------
@lru_cache(maxsize=8)
def logo_mark(size: int) -> Image.Image:
    src = Image.open(os.path.join(ASSETS, "Tr.png")).convert("RGB").resize((size, size), Image.LANCZOS)
    arr = np.asarray(src).astype(np.int16)
    alpha = 255 - arr.min(axis=2)
    alpha = np.clip((alpha - 6) * 1.55, 0, 255).astype(np.uint8)
    mask = Image.fromarray(alpha, "L")

    # cyan -> blue vertical gradient
    out = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    px = np.zeros((size, size, 4), dtype=np.uint8)
    top = np.array((120, 240, 255), dtype=np.float32)
    bot = np.array((66, 120, 255), dtype=np.float32)
    for y in range(size):
        c = top + (bot - top) * (y / max(1, size - 1))
        px[y, :, 0] = int(c[0])
        px[y, :, 1] = int(c[1])
        px[y, :, 2] = int(c[2])
    px[:, :, 3] = np.asarray(mask)
    return Image.fromarray(px, "RGBA")


def draw_logo(layer: Image.Image, cx: int, cy: int, size: int, alpha: int = 255, glow: bool = True):
    if glow:
        gr = int(size * 1.05)
        g = Image.new("RGBA", (gr * 2, gr * 2), (0, 0, 0, 0))
        gd = ImageDraw.Draw(g)
        gd.ellipse((gr * 0.35, gr * 0.35, gr * 1.65, gr * 1.65), fill=a(ACCENT, 28))
        g = g.filter(ImageFilter.GaussianBlur(int(size * 0.10)))
        layer.alpha_composite(g, (cx - gr, cy - gr))
    mark = logo_mark(size)
    if alpha < 255:
        mark = mark.copy()
        mark.putalpha(Image.eval(mark.getchannel("A"), lambda v: int(v * alpha / 255)))
    layer.alpha_composite(mark, (cx - size // 2, cy - size // 2))


# Shared UI -------------------------------------------------------------
def scene_label(layer: Image.Image, ctx: Ctx, idx: int, label: str, t: float):
    d = ImageDraw.Draw(layer)
    x = 80 if ctx.land else 60
    y = 64 if ctx.land else 86
    text(d, (x, y), f"{idx:02d}", f_monob(ctx.fs(22)), ACCENT, alpha=220)
    text(d, (x + ctx.fs(38), y + 2), label.upper(), f_mono(ctx.fs(16)), MUTED, alpha=200)
    d.line((x, y + ctx.fs(36), x + ctx.fs(210), y + ctx.fs(36)), fill=a(BORDER, 150), width=1)


def progress_dots(layer: Image.Image, ctx: Ctx, current: int):
    d = ImageDraw.Draw(layer)
    n = 9
    gap = 22 if ctx.land else 17
    bw = 14 if ctx.land else 10
    x = ctx.w - (80 if ctx.land else 60) - (n - 1) * gap
    y = 78 if ctx.land else 98
    for i in range(n):
        col = ACCENT if i == current else DIM
        alpha = 220 if i == current else 85
        round_rect(d, (x + i * gap, y, x + i * gap + bw, y + 3), 2, fill=a(col, alpha))


def caption(layer: Image.Image, ctx: Ctx, title: str, sub: str, t: float, start: float = 0.0):
    p = ease_out(clamp((t - start) / 0.55))
    if p <= 0:
        return
    d = ImageDraw.Draw(layer)
    y = ctx.caption_y - int(14 * (1 - p))
    alpha = int(255 * p)
    # accent rule
    rw = 34
    round_rect(d, (ctx.cx - rw // 2, y - 26, ctx.cx + rw // 2, y - 22), 2, fill=a(ACCENT, alpha))
    text(d, (ctx.cx, y), title, f_zhb(ctx.fs(45)), TEXT, anchor="ma", alpha=alpha)
    if sub:
        text(d, (ctx.cx, ctx.sub_y), sub, f_zh(ctx.fs(23)), MUTED, anchor="ma", alpha=int(alpha * 0.92))


def terminal(layer: Image.Image, ctx: Ctx, box, title: str) -> tuple[int, int, int, int]:
    x0, y0, x1, y1 = box
    d = ImageDraw.Draw(layer)
    shadow_rect(d, box, 18, 16)
    round_rect(d, box, 18, fill=a(PANEL, 246), outline=a(BORDER, 220), width=1)
    th = 46 if ctx.land else 42
    round_rect(d, (x0, y0, x1, y0 + th), 18, fill=a(PANEL_2, 255))
    d.rectangle((x0, y0 + th - 1, x1, y0 + th), fill=a(BORDER, 200))
    # window dots
    for i, c in enumerate((RED, AMBER, GREEN)):
        cx = x0 + 24 + i * 20
        d.ellipse((cx - 5, y0 + th // 2 - 5, cx + 5, y0 + th // 2 + 5), fill=a(c, 210))
    text(d, (x0 + 100, y0 + th // 2), title, f_mono(ctx.fs(17)), MUTED, anchor="lm")
    text(d, (x1 - 24, y0 + th // 2), "×", f_mono(ctx.fs(18)), DIM, anchor="rm")
    # thin status strip keeps the window from feeling bottom-heavy
    d.line((x0 + 24, y1 - 40, x1 - 24, y1 - 40), fill=a(BORDER, 130), width=1)
    text(d, (x0 + 24, y1 - 30), "tr 2.2.1", f_mono(ctx.fs(13)), (72, 88, 112))
    text(d, (x1 - 24, y1 - 30), "UTF-8", f_mono(ctx.fs(13)), (72, 88, 112), anchor="ra")
    return x0 + 32, y0 + th + 30, x1 - x0 - 64, y1 - y0 - th - 52


def is_cjk(ch: str) -> bool:
    o = ord(ch)
    return 0x2E80 <= o <= 0x9FFF or 0xF900 <= o <= 0xFAFF or 0xFF00 <= o <= 0xFFEF


def draw_mixed(draw: ImageDraw.ImageDraw, x: int, y: int, s: str, size: int, fill=TEXT, alpha=255) -> int:
    """Draw ASCII with the mono font and CJK with the Chinese font."""
    cx = float(x)
    for ch in s:
        font = f_zh(size) if is_cjk(ch) else f_mono(size)
        draw.text((cx, y), ch, font=font, fill=a(fill, alpha))
        cx += draw.textlength(ch, font=font)
    return int(cx)


def prompt(draw: ImageDraw.ImageDraw, x: int, y: int, cmd: str, typed: int, fs: int, blink: bool):
    font = f_mono(fs)
    text(draw, (x, y), "$", f_monob(fs), ACCENT)
    px = x + int(text_w(draw, "$ ", font))
    shown = cmd[:typed]
    if any(is_cjk(c) for c in shown):
        endx = draw_mixed(draw, px, y, shown, fs, TEXT)
    else:
        text(draw, (px, y), shown, font, TEXT)
        endx = px + int(text_w(draw, shown, font))
    if typed < len(cmd) or blink:
        cx = endx + 3
        draw.rectangle((cx, y + fs * 0.10, cx + fs * 0.52, y + fs * 1.22), fill=a(ACCENT, 230))


def chip(layer: Image.Image, ctx: Ctx, xy, label: str, fg=ACCENT, bg=(24, 48, 55), alpha=255, pad=14):
    d = ImageDraw.Draw(layer)
    f = f_zhb(ctx.fs(19))
    w = int(text_w(d, label, f)) + pad * 2
    h = ctx.fs(38)
    x, y = xy
    round_rect(d, (x, y, x + w, y + h), h // 2, fill=a(bg, int(alpha * 0.9)), outline=a(fg, int(alpha * 0.5)))
    text(d, (x + pad, y + h // 2), label, f, fg, anchor="lm", alpha=alpha)
    return w, h


# Scenes ---------------------------------------------------------------
SCENES = [
    ("intro", 0.0, 3.5),
    ("cli", 3.5, 10.5),
    ("arrow", 10.5, 18.0),
    ("tui", 18.0, 27.0),
    ("file", 27.0, 35.0),
    ("offline", 35.0, 41.0),
    ("features", 41.0, 48.0),
    ("install", 48.0, 54.0),
    ("end", 54.0, 60.0),
]


def scene_intro(layer, ctx: Ctx, t: float):
    p = ease_out(clamp(t / 1.0))
    cx = ctx.cx
    cy = int(ctx.h * (0.35 if ctx.land else 0.34))
    size = 200 if ctx.land else 240
    draw_logo(layer, cx, cy, int(size * lerp(0.88, 1.0, p)), alpha=int(255 * p))
    # orbit ring
    d = ImageDraw.Draw(layer)
    rr = size * 0.86
    d.arc((cx - rr, cy - rr, cx + rr, cy + rr), 205, 335, fill=a(ACCENT, int(60 * p)), width=2)

    ty = cy + (150 if ctx.land else 190)
    text(d, (cx, ty), "Tr", f_enb(ctx.fs(102)), TEXT, anchor="ma", alpha=int(255 * p))
    text(d, (cx, ty + ctx.fs(118)), "终端里的 AI 翻译器", f_zhb(ctx.fs(38)), (198, 230, 240), anchor="ma", alpha=int(255 * p))
    text(d, (cx, ty + ctx.fs(172)), "CLI + TUI  ·  纯 Go  ·  默认 AI 接口", f_zh(ctx.fs(21)), MUTED, anchor="ma", alpha=int(230 * p))
    # version chip, centered using its measured width
    lbl = "v2.2.1"
    cw = int(text_w(d, lbl, f_zhb(ctx.fs(19)))) + 28
    chip(layer, ctx, (cx - cw // 2, ty + ctx.fs(220)), lbl, fg=ACCENT, bg=(20, 42, 50), alpha=int(210 * p))


def scene_cli(layer, ctx: Ctx, t: float):
    d = ImageDraw.Draw(layer)
    scene_label(layer, ctx, 1, "CLI", t)
    progress_dots(layer, ctx, 1)
    if ctx.land:
        box = (330, 205, 1590, 765)
    else:
        box = (60, 470, 1020, 1140)
    x, y, cw, ch = terminal(layer, ctx, box, "tr — PowerShell")
    fs = ctx.fs(29 if ctx.land else 25)
    cmd = 'tr "Hello, world!"'
    cps = 20
    typed = int(clamp((t - 0.8) * cps, 0, len(cmd)))
    prompt(d, x, y, cmd, typed, fs, blink=(int(t * 2.2) % 2 == 0))
    if typed == len(cmd):
        text(d, (x, y + fs * 2.1), ">", f_monob(fs), GREEN)
        text(d, (x + fs, y + fs * 2.1), "你好，世界！", f_zhb(fs + 8), TEXT)
        text(d, (x + fs, y + fs * 4.0), "English → 中文", f_zh(fs - 5), MUTED)
        # a couple of quiet examples fill the terminal without competing
        text(d, (x, y + fs * 5.7), '$ echo "hello" | tr', f_mono(fs - 8), (80, 96, 120))
        text(d, (x, y + fs * 6.5), "$ tr -f README.md", f_mono(fs - 8), (80, 96, 120))
    if t > 0.8:
        chip(layer, ctx, (x + cw - 160, y - 2), "CLI", fg=GREEN, bg=(22, 48, 38))
    caption(layer, ctx, "一条命令，直接翻译", "CLI 模式  ·  管道友好  ·  输出干净", t, 0.4)


def scene_arrow(layer, ctx: Ctx, t: float):
    d = ImageDraw.Draw(layer)
    scene_label(layer, ctx, 2, "AUTO DETECT", t)
    progress_dots(layer, ctx, 2)
    # language chips above the terminal
    cy = 150 if ctx.land else 310
    w1, _ = chip(layer, ctx, (ctx.cx - 260, cy), "中文", fg=(190, 230, 255), bg=(26, 38, 62))
    # arrow
    ax0 = ctx.cx - 260 + w1 + 26
    arrow_p = ease_out(clamp((t - 0.3) / 0.5))
    ax1 = ax0 + int(90 * arrow_p)
    d.line((ax0, cy + 19, ax1, cy + 19), fill=a(ACCENT, 220), width=2)
    d.polygon((ax1, cy + 19, ax1 - 10, cy + 13, ax1 - 10, cy + 25), fill=a(ACCENT, 220))
    chip(layer, ctx, (ax1 + 26, cy), "English", fg=ACCENT, bg=(20, 48, 54))

    if ctx.land:
        box = (360, 275, 1560, 760)
    else:
        box = (60, 500, 1020, 1140)
    x, y, cw, ch = terminal(layer, ctx, box, "tr — PowerShell")
    fs = ctx.fs(29 if ctx.land else 25)
    cmd = "tr 你好 -> en"
    typed = int(clamp((t - 0.7) * 18, 0, len(cmd)))
    prompt(d, x, y, cmd, typed, fs, blink=(int(t * 2.2) % 2 == 0))
    if typed == len(cmd):
        text(d, (x, y + fs * 2.1), ">", f_monob(fs), GREEN)
        text(d, (x + fs, y + fs * 2.1), "Hello", f_mono(fs + 8), TEXT)
        text(d, (x + fs, y + fs * 4.0), "AUTO-DETECT  →  en", f_mono(fs - 8), MUTED)
        text(d, (x, y + fs * 5.5), "source: auto-detect (中文)", f_zh(fs - 8), (78, 96, 120))
        text(d, (x, y + fs * 6.3), "target: en", f_mono(fs - 8), (78, 96, 120))
    caption(layer, ctx, "自动识别语言，指定目标", "`->` 只作用于本次执行，不写配置", t, 0.5)


def scene_tui(layer, ctx: Ctx, t: float):
    d = ImageDraw.Draw(layer)
    scene_label(layer, ctx, 3, "TUI", t)
    progress_dots(layer, ctx, 3)
    if ctx.land:
        box = (250, 175, 1670, 805)
    else:
        box = (50, 430, 1030, 1320)
    shadow_rect(d, box, 18, 16)
    round_rect(d, box, 18, fill=a(PANEL, 248), outline=a(BORDER, 220))
    x0, y0, x1, y1 = box
    # header
    text(d, (x0 + 30, y0 + 28), "Tr", f_enb(ctx.fs(25)), ACCENT)
    text(d, (x0 + 80, y0 + 30), "en → zh", f_mono(ctx.fs(20)), MUTED)
    d.ellipse((x0 + 190, y0 + 38, x0 + 200, y0 + 48), fill=a(GREEN, 230))
    text(d, (x0 + 210, y0 + 31), "在线", f_zh(ctx.fs(19)), MUTED)
    text(d, (x1 - 30, y0 + 31), "Ctrl+Q 退出", f_zh(ctx.fs(17)), DIM, anchor="ra")

    pad = 28 if ctx.land else 22
    gap = 18 if ctx.land else 16
    top = y0 + 78
    bottom = y1 - 76
    pw = (x1 - x0 - pad * 2 - gap) // 2
    left = (x0 + pad, top, x0 + pad + pw, bottom)
    right = (left[2] + gap, top, x1 - pad, bottom)
    focus = int(t / 4) % 2
    for i, (bx, label, code) in enumerate(((left, "原文", "en"), (right, "译文", "zh"))):
        col = ACCENT if i == focus else BORDER
        round_rect(d, bx, 12, fill=a((10, 17, 29), 255), outline=a(col, 220 if i == focus else 160))
        text(d, (bx[0] + 18, bx[1] + 16), f"{label} ({code})", f_zh(ctx.fs(17)), MUTED)
    # left input
    lx, ly = left[0] + 18, left[1] + 58
    if t > 0.8:
        shown = "hello world"[: int(clamp((t - 0.8) * 20, 0, 11))]
        text(d, (lx, ly), shown, f_mono(ctx.fs(25 if ctx.land else 21)), TEXT)
        if shown != "hello world" or int(t * 2.2) % 2 == 0:
            cw = text_w(d, shown, f_mono(ctx.fs(25 if ctx.land else 21)))
            d.rectangle((lx + cw + 2, ly + 2, lx + cw + 12, ly + ctx.fs(30)), fill=a(ACCENT, 220))
    # right result
    if t > 2.2:
        text(d, (right[0] + 18, right[1] + 58), "你好，世界", f_zhb(ctx.fs(25 if ctx.land else 21)), TEXT)
        text(d, (right[0] + 18, right[1] + 110), "在线 AI · 缓存未命中", f_zh(ctx.fs(15)), MUTED)
    # footer keys
    keys = "Ctrl+T 翻译    Ctrl+←/→ 切语言    Tab 焦点    Esc 命令行"
    text(d, (ctx.cx, y1 - 28), keys, f_zh(ctx.fs(16 if ctx.land else 14)), DIM, anchor="md")
    caption(layer, ctx, "左右对照的 TUI", "多行输入  ·  语言切换  ·  内建系统命令行", t, 0.5)


def scene_file(layer, ctx: Ctx, t: float):
    d = ImageDraw.Draw(layer)
    scene_label(layer, ctx, 4, "FILE", t)
    progress_dots(layer, ctx, 4)
    if ctx.land:
        box = (330, 235, 1590, 735)
    else:
        box = (60, 470, 1020, 1110)
    x, y, cw, ch = terminal(layer, ctx, box, "tr — PowerShell")
    fs = ctx.fs(29 if ctx.land else 24)
    cmd = "tr -f README.md docs/zh.md"
    typed = int(clamp((t - 0.7) * 22, 0, len(cmd)))
    prompt(d, x, y, cmd, typed, fs, blink=(int(t * 2.2) % 2 == 0))
    if typed == len(cmd):
        # files
        chip(layer, ctx, (x, y + fs * 1.9), "README.md", fg=(198, 222, 255), bg=(24, 34, 54))
        text(d, (x + 170, y + fs * 1.9 + 8), "→", f_zhb(ctx.fs(22)), ACCENT)
        chip(layer, ctx, (x + 205, y + fs * 1.9), "docs/zh.md", fg=GREEN, bg=(22, 44, 36))
        # progress
        p = ease_out(clamp((t - 2.0) / 2.4))
        bx0, by0 = x, y + fs * 3.9
        bw = cw - 120
        round_rect(d, (bx0, by0, bx0 + bw, by0 + 10), 5, fill=a((30, 44, 66), 255))
        round_rect(d, (bx0, by0, bx0 + int(bw * p), by0 + 10), 5, fill=a(ACCENT, 255))
        label = "翻译中 3/8" if p < 0.99 else "完成 · 已写入 docs/zh.md"
        text(d, (bx0, by0 + 24), label, f_zh(ctx.fs(18)), MUTED)
        if p >= 0.99:
            draw_check(d, bx0 + bw - 24, by0 - 14, 24)
    caption(layer, ctx, "整篇文件翻译，保留格式", "按行分块  ·  保留缩进 / BOM / CRLF", t, 0.5)


def scene_offline(layer, ctx: Ctx, t: float):
    d = ImageDraw.Draw(layer)
    scene_label(layer, ctx, 5, "OFFLINE", t)
    progress_dots(layer, ctx, 5)
    if ctx.land:
        box = (440, 250, 1480, 690)
    else:
        box = (90, 500, 990, 1100)
    x, y, cw, ch = terminal(layer, ctx, box, "tr — PowerShell")
    fs = ctx.fs(29 if ctx.land else 25)
    cmd = 'tr -o "hello"'
    typed = int(clamp((t - 0.6) * 20, 0, len(cmd)))
    prompt(d, x, y, cmd, typed, fs, blink=(int(t * 2.2) % 2 == 0))
    if typed == len(cmd):
        text(d, (x, y + fs * 2.2), ">", f_monob(fs), GREEN)
        text(d, (x + fs, y + fs * 2.2), "你好", f_zhb(fs + 8), TEXT)
        chip(layer, ctx, (x, y + fs * 4.1), "OFFLINE · 0 token", fg=GREEN, bg=(20, 44, 34))
        text(d, (x + 290, y + fs * 4.3), "不联网", f_zh(ctx.fs(17)), MUTED)
    caption(layer, ctx, "离线也能用", "缓存 + 内置词典  ·  不联网、不消耗 token", t, 0.5)


def draw_check(draw: ImageDraw.ImageDraw, x: int, y: int, size: int, color=GREEN):
    w = max(2, size // 9)
    draw.line(
        (x, y + size * 0.55, x + size * 0.38, y + size * 0.9, x + size, y + size * 0.12),
        fill=a(color, 245),
        width=w,
        joint="curve",
    )


def icon_terminal(d, box, col):
    x0, y0, x1, y1 = box
    round_rect(d, box, 8, outline=a(col, 220), width=2)
    d.line((x0, y0 + 14, x1, y0 + 14), fill=a(col, 160), width=1)
    for i, c in enumerate((RED, AMBER, GREEN)):
        d.ellipse((x0 + 7 + i * 9, y0 + 5, x0 + 11 + i * 9, y0 + 9), fill=a(c, 180))
    d.line((x0 + 12, y0 + 26, x0 + 22, y0 + 34), fill=a(col, 230), width=2)
    d.line((x0 + 22, y0 + 26, x0 + 12, y0 + 34), fill=a(col, 230), width=2)
    d.line((x0 + 29, y0 + 34, x0 + 44, y0 + 34), fill=a(col, 230), width=2)


def icon_doc(d, box, col):
    x0, y0, x1, y1 = box
    round_rect(d, (x0 + 4, y0, x1 - 4, y1), 6, outline=a(col, 220), width=2)
    for i in range(3):
        d.line((x0 + 13, y0 + 14 + i * 9, x1 - 13, y0 + 14 + i * 9), fill=a(col, 150), width=1)
    d.line((x1 - 16, y0, x1 - 16, y0 + 10), fill=a(col, 220), width=2)


def icon_globe(d, box, col):
    x0, y0, x1, y1 = box
    d.ellipse(box, outline=a(col, 220), width=2)
    d.ellipse((x0 + 11, y0, x1 - 11, y1), outline=a(col, 150), width=1)
    d.line((x0, (y0 + y1) // 2, x1, (y0 + y1) // 2), fill=a(col, 150), width=1)


def icon_monitor(d, box, col):
    x0, y0, x1, y1 = box
    round_rect(d, (x0, y0, x1, y1 - 8), 5, outline=a(col, 220), width=2)
    d.line((ctx_mid(x0, x1), y1 - 8, ctx_mid(x0, x1), y1), fill=a(col, 220), width=2)
    d.line((x0 + 12, y1, x1 - 12, y1), fill=a(col, 220), width=2)


def ctx_mid(a_, b_):
    return (a_ + b_) // 2


def scene_features(layer, ctx: Ctx, t: float):
    d = ImageDraw.Draw(layer)
    scene_label(layer, ctx, 6, "FEATURES", t)
    progress_dots(layer, ctx, 6)
    text(d, (ctx.cx, 150 if ctx.land else 310), "为终端而生", f_zhb(ctx.fs(40)), TEXT, anchor="ma")
    text(d, (ctx.cx, 200 if ctx.land else 370), "CLI 与 TUI 共用同一套缓存和语言设置", f_zh(ctx.fs(20)), MUTED, anchor="ma")

    cards = [
        ("CLI + TUI", "命令与交互界面共用", icon_terminal, ACCENT),
        ("文件翻译", "整篇处理，保留结构", icon_doc, GREEN),
        ("多语言界面", "中文 / English / 日本語", icon_globe, ACCENT_2),
        ("跨平台", "Windows / Linux / macOS", icon_monitor, AMBER),
    ]
    if ctx.land:
        cw, chh = 500, 180
        gapx, gapy = 30, 26
        total_w = cw * 2 + gapx
        x0 = (ctx.w - total_w) // 2
        y0 = 300
    else:
        cw, chh = 900, 160
        gapx, gapy = 0, 24
        x0 = (ctx.w - cw) // 2
        y0 = 520
    for i, (title, sub, icon, col) in enumerate(cards):
        p = ease_out(clamp((t - 0.5 - i * 0.18) / 0.5))
        if p <= 0:
            continue
        if ctx.land:
            row, col_i = divmod(i, 2)
            x = x0 + col_i * (cw + gapx)
            y = y0 + row * (chh + gapy) + int(18 * (1 - p))
        else:
            x = x0
            y = y0 + i * (chh + gapy) + int(18 * (1 - p))
        box = (x, y, x + cw, y + chh)
        shadow_rect(d, box, 16, 10)
        round_rect(d, box, 16, fill=a(PANEL, 245), outline=a(BORDER, 200))
        icon(ImageDraw.Draw(layer), (x + 28, y + 44, x + 28 + 46, y + 44 + 46), col)
        text(d, (x + 96, y + 38), title, f_zhb(ctx.fs(25)), TEXT, alpha=int(255 * p))
        text(d, (x + 96, y + 78), sub, f_zh(ctx.fs(17)), MUTED, alpha=int(230 * p))
        # accent bottom line
        d.line((x + 28, y + chh - 24, x + 28 + int((cw - 56) * p), y + chh - 24), fill=a(col, 200), width=2)
    caption(layer, ctx, "一个工具，覆盖日常翻译", "CLI  ·  TUI  ·  文件  ·  离线", t, 0.3)


def scene_install(layer, ctx: Ctx, t: float):
    d = ImageDraw.Draw(layer)
    scene_label(layer, ctx, 7, "INSTALL", t)
    progress_dots(layer, ctx, 7)
    if ctx.land:
        box = (300, 260, 1620, 660)
    else:
        box = (60, 480, 1020, 920)
    x, y, cw, ch = terminal(layer, ctx, box, "tr — PowerShell")
    fs = ctx.fs(25 if ctx.land else 21)
    cmd = "go install github.com/Qiuxile/Tr@latest"
    typed = int(clamp((t - 0.7) * 30, 0, len(cmd)))
    prompt(d, x, y, cmd, typed, fs, blink=(int(t * 2.2) % 2 == 0))
    if typed == len(cmd):
        draw_check(d, x, y + fs * 2.1, fs)
        text(d, (x + fs * 1.4, y + fs * 2.2), "installed", f_mono(fs), GREEN)
        text(d, (x + fs * 1.4, y + fs * 3.5), "Tr v2.2.1  ·  Windows installer available", f_mono(fs - 7), MUTED)
    text(d, (ctx.cx, (ctx.h - 240) if ctx.land else 1260), "github.com/Qiuxile/Tr", f_mono(ctx.fs(24)), ACCENT, anchor="ma")
    caption(layer, ctx, "安装只要一行", "或使用 Windows 安装包", t, 0.5)


def scene_end(layer, ctx: Ctx, t: float):
    p = ease_out(clamp(t / 0.8))
    cx = ctx.cx
    cy = int(ctx.h * (0.34 if ctx.land else 0.33))
    size = 170 if ctx.land else 210
    draw_logo(layer, cx, cy, size, alpha=int(255 * p))
    d = ImageDraw.Draw(layer)
    text(d, (cx, cy + 140 if ctx.land else cy + 170), "Tr", f_enb(ctx.fs(86)), TEXT, anchor="ma", alpha=int(255 * p))
    text(d, (cx, cy + 250 if ctx.land else cy + 300), "终端里的 AI 翻译器", f_zhb(ctx.fs(32)), (198, 230, 240), anchor="ma", alpha=int(255 * p))
    text(d, (cx, cy + 315 if ctx.land else cy + 380), "github.com/Qiuxile/Tr", f_mono(ctx.fs(23)), ACCENT, anchor="ma", alpha=int(255 * p))
    text(d, (cx, cy + 365 if ctx.land else cy + 435), "Apache-2.0  ·  Go 1.26+", f_zh(ctx.fs(18)), DIM, anchor="ma", alpha=int(220 * p))


SCENE_FUNCS = {
    "intro": scene_intro,
    "cli": scene_cli,
    "arrow": scene_arrow,
    "tui": scene_tui,
    "file": scene_file,
    "offline": scene_offline,
    "features": scene_features,
    "install": scene_install,
    "end": scene_end,
}


def scene_alpha(t: float, start: float, end: float) -> float:
    local = t - start
    dur = end - start
    fade = 0.32
    if local < 0 or local >= dur:
        return 0.0
    if local < fade:
        return ease_out(local / fade)
    if local > dur - fade:
        return ease_out((dur - local) / fade)
    return 1.0


def render_frame(t: float, ctx: Ctx, bg: Image.Image) -> Image.Image:
    frame = bg.copy()
    for name, start, end in SCENES:
        if start <= t < end:
            alpha = scene_alpha(t, start, end)
            if alpha <= 0:
                continue
            layer = Image.new("RGBA", (ctx.w, ctx.h), (0, 0, 0, 0))
            SCENE_FUNCS[name](layer, ctx, t - start)
            if alpha < 1:
                layer.putalpha(Image.eval(layer.getchannel("A"), lambda v: int(v * alpha)))
            frame = Image.alpha_composite(frame.convert("RGBA"), layer).convert("RGB")
            break
    # subtle top/bottom vignette for legibility
    return frame


def render_still(t: float, portrait: bool, out: str):
    w, h = (1080, 1920) if portrait else (1920, 1080)
    ctx = Ctx(w, h, portrait)
    bg = make_bg(w, h, portrait)
    frame = render_frame(t, ctx, bg)
    ImageDraw.Draw(frame)
    os.makedirs(os.path.dirname(out), exist_ok=True)
    frame.save(out)
    print("wrote", out)


def render_video(portrait: bool, out: str, audio: str | None, seconds: float = DURATION):
    w, h = (1080, 1920) if portrait else (1920, 1080)
    ctx = Ctx(w, h, portrait)
    bg = make_bg(w, h, portrait)
    os.makedirs(os.path.dirname(out), exist_ok=True)
    log_path = os.path.join(VIDEO_DIR, "ffmpeg.log")
    log = open(log_path, "ab")
    cmd = [
        ffmpeg_path(),
        "-y",
        "-f",
        "rawvideo",
        "-vcodec",
        "rawvideo",
        "-s",
        f"{w}x{h}",
        "-pix_fmt",
        "rgb24",
        "-r",
        str(FPS),
        "-i",
        "pipe:0",
    ]
    if audio and os.path.exists(audio):
        cmd += ["-i", audio]
    cmd += [
        "-c:v",
        "libx264",
        "-preset",
        "medium",
        "-crf",
        "18",
        "-pix_fmt",
        "yuv420p",
        "-movflags",
        "+faststart",
    ]
    if audio and os.path.exists(audio):
        cmd += ["-c:a", "aac", "-b:a", "128k"]
    cmd += [out]
    print("ffmpeg:", " ".join(cmd), flush=True)
    proc = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=log, stderr=log)
    total = int(seconds * FPS)
    written = 0
    try:
        for i in range(total):
            t = i / FPS
            frame = render_frame(t, ctx, bg)
            proc.stdin.write(frame.tobytes())
            written += 1
            if i % 150 == 0:
                print(f"{os.path.basename(out)}: {i}/{total}", flush=True)
        print(f"{os.path.basename(out)}: wrote {written}/{total} frames", flush=True)
        proc.stdin.close()
        proc.wait()
    finally:
        log.close()
    if proc.returncode != 0:
        raise SystemExit(f"ffmpeg failed; see {log_path}")


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--still", type=float)
    ap.add_argument("--portrait", action="store_true")
    ap.add_argument("--out", default="")
    ap.add_argument("--all", action="store_true")
    ap.add_argument("--clip", action="store_true", help="render a 6s 16:9 test clip")
    args = ap.parse_args()
    if args.still is not None:
        out = args.out or os.path.join(OUTPUT, f"still-{args.still:.1f}{'-p' if args.portrait else ''}.png")
        render_still(args.still, args.portrait, out)
        return
    if args.clip:
        render_video(False, os.path.join(VIDEO_DIR, "clip-6s.mp4"), os.path.join(VIDEO_DIR, "audio.wav"), 6.0)
        return
    if args.all:
        audio = os.path.join(VIDEO_DIR, "audio.wav")
        render_video(False, os.path.join(OUTPUT, "tr-intro-60s-16x9.mp4"), audio)
        render_video(True, os.path.join(OUTPUT, "tr-intro-60s-9x16.mp4"), audio)
        return
    ap.print_help()


if __name__ == "__main__":
    main()
