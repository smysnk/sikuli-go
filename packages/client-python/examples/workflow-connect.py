from __future__ import annotations

from bootstrap import ensure_sikuligo_on_path
from sikuligo import Pattern, Screen

ensure_sikuligo_on_path()

# Connect-only workflow (requires sikuligo already running).
screen = Screen.connect()
try:
    match = screen.click(Pattern("assets/pattern.png").exact())
    print(f"clicked match target at ({match.target_x}, {match.target_y})")
finally:
    screen.close()
