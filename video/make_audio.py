#!/usr/bin/env python3
"""Generate an original, low-volume ambient bed + typing clicks for the intro.

No third-party music is used, so the result can be published without a music
license. Replace audio.wav with your own track if you prefer.
"""
import math
import os
import wave

import numpy as np

SR = 44100
DUR = 60.0
N = int(SR * DUR)
rng = np.random.default_rng(2026)


def note(name: str) -> float:
    # equal temperament, A4=440
    names = {"C": 0, "C#": 1, "D": 2, "D#": 3, "E": 4, "F": 5,
             "F#": 6, "G": 7, "G#": 8, "A": 9, "A#": 10, "B": 11}
    if name[-1].isdigit():
        octave = int(name[-1])
        key = name[:-1]
    else:
        octave = 4
        key = name
    midi = 12 * (octave + 1) + names[key]
    return 440.0 * 2 ** ((midi - 69) / 12)


def adsr(n: int, a: float, d: float, s: float, r: float) -> np.ndarray:
    env = np.ones(n)
    ai = min(n, int(a * SR))
    di = min(n, int((a + d) * SR))
    ri = min(n, int(r * SR))
    if ai:
        env[:ai] = np.linspace(0, 1, ai)
    if di > ai:
        env[ai:di] = np.linspace(1, s, di - ai)
    env[di:] = s
    if ri:
        env[-ri:] *= np.linspace(1, 0, ri)
    return env


def add(buf: np.ndarray, start: float, sig: np.ndarray, gain: float = 1.0):
    i = int(start * SR)
    if i < 0:
        sig = sig[-i:]
        i = 0
    j = min(len(buf), i + len(sig))
    if j > i:
        buf[i:j] += sig[: j - i] * gain


# Chord bed: Am - F - C - G, repeated over 8 bars of 7.5s.
chords = [
    ("A2", ["A3", "C4", "E4"]),
    ("F2", ["F3", "A3", "C4"]),
    ("C3", ["C4", "E4", "G4"]),
    ("G2", ["G3", "B3", "D4"]),
]
mix = np.zeros(N, dtype=np.float64)
bar = 7.5
for b in range(8):
    root, notes = chords[b % 4]
    t0 = b * bar
    dur = bar + 1.4
    n = int(dur * SR)
    tt = np.arange(n) / SR
    env = adsr(n, 1.4, 1.6, 0.55, 1.8)
    sig = np.zeros(n)
    for k, nm in enumerate(notes):
        f = note(nm)
        sig += np.sin(2 * math.pi * f * tt + k * 0.8) * (0.55 if k == 0 else 0.32)
        sig += np.sin(2 * math.pi * f * 2 * tt) * 0.08
    # detuned warmth
    sig += np.sin(2 * math.pi * (note(notes[0]) * 0.997) * tt) * 0.12
    sig *= env
    add(mix, t0, sig, 0.16)
    # sub bass
    sub = np.sin(2 * math.pi * note(root) * tt) * adsr(n, 0.8, 1.2, 0.6, 1.5)
    add(mix, t0, sub, 0.14)

# Sparse high plucks, pentatonic, every 1.875s.
pluck_notes = ["A4", "C5", "E5", "G5", "E5", "C5"]
for i in range(int(DUR / 1.875)):
    t0 = i * 1.875 + 0.4
    f = note(pluck_notes[i % len(pluck_notes)])
    n = int(2.4 * SR)
    tt = np.arange(n) / SR
    sig = (np.sin(2 * math.pi * f * tt) + 0.35 * np.sin(2 * math.pi * f * 2.01 * tt))
    sig *= np.exp(-tt * 3.2) * (1 - np.exp(-tt * 40))
    add(mix, t0, sig, 0.035)

# Typing clicks aligned with the terminal scenes.
typing_windows = [
    (4.30, 5.35),   # tr "Hello, world!"
    (11.15, 12.10), # tr 你好 -> en
    (27.65, 28.75), # tr -f README.md docs/zh.md
    (35.55, 36.35), # tr -o "hello"
    (48.65, 50.15), # go install ...
]
for start, end in typing_windows:
    t = start
    while t < end:
        n = int(0.035 * SR)
        tt = np.arange(n) / SR
        click = rng.normal(0, 1, n) * np.exp(-tt * 180)
        click += 0.4 * np.sin(2 * math.pi * 2100 * tt) * np.exp(-tt * 120)
        add(mix, t, click, 0.045)
        t += 0.068

# Soft transition swooshes at scene boundaries.
for t0 in (3.35, 10.35, 17.85, 26.85, 34.85, 40.85, 47.85, 53.85):
    n = int(0.45 * SR)
    tt = np.arange(n) / SR
    noise = rng.normal(0, 1, n)
    # one-pole lowpass
    for _ in range(2):
        noise = np.convolve(noise, np.ones(24) / 24, mode="same")
    env = np.sin(math.pi * np.clip(tt / 0.45, 0, 1)) ** 2
    add(mix, t0, noise * env, 0.012)

# Gentle global fades and soft clip.
fade_in = int(1.0 * SR)
fade_out = int(1.8 * SR)
mix[:fade_in] *= np.linspace(0, 1, fade_in)
mix[-fade_out:] *= np.linspace(1, 0, fade_out)
mix = np.tanh(mix * 1.2) * 0.65
# normalize to a comfortable peak
peak = max(1e-9, float(np.max(np.abs(mix))))
mix = mix / peak * 0.62

# Mono 16-bit WAV.
pcm = (mix * 32767).astype(np.int16)
out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "audio.wav")
with wave.open(out, "wb") as w:
    w.setnchannels(1)
    w.setsampwidth(2)
    w.setframerate(SR)
    w.writeframes(pcm.tobytes())
print("wrote", out, "seconds", DUR)
