from __future__ import annotations

import os
import sys
from pathlib import Path


def _venv_python_path(project_root: Path) -> Path:
    if os.name == "nt":
        return project_root / ".venv" / "Scripts" / "python.exe"
    return project_root / ".venv" / "bin" / "python"


def ensure_project_venv_python() -> None:
    if os.environ.get("SIKULI_GO_VENV_REEXEC") == "1":
        return

    script = Path(__file__).resolve()
    project_root = script.parent.parent
    venv_python = _venv_python_path(project_root)
    if not venv_python.exists():
        return

    env = dict(os.environ)
    env["SIKULI_GO_VENV_REEXEC"] = "1"
    os.execve(str(venv_python), [str(venv_python), *sys.argv], env)
