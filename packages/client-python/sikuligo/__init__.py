from .client import (
    Sikuli,
    SikuliError,
    gray_image_from_png,
    gray_image_from_rows,
    pattern_from_png,
    screen_query_options,
)
from .sikulix import LaunchMeta, Match, Pattern, Region, Screen as _Screen


def Screen(**kwargs):
    return _Screen.start(**kwargs)


Screen.start = _Screen.start
Screen.connect = _Screen.connect
Screen.spawn = _Screen.spawn
Screen.auto = _Screen.auto

__all__ = [
    "Sikuli",
    "SikuliError",
    "LaunchMeta",
    "Match",
    "Pattern",
    "Region",
    "Screen",
    "gray_image_from_png",
    "gray_image_from_rows",
    "pattern_from_png",
    "screen_query_options",
]
