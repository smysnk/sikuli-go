# SikuliGO Python Client

This directory contains the Python client for SikuliGO with Sikuli-style `Screen` + `Pattern` APIs.

## Links

- Main repository: [github.com/smysnk/SikuliGO](https://github.com/smysnk/SikuliGO)
- API reference: [smysnk.github.io/SikuliGO/api](https://smysnk.github.io/SikuliGO/api/)
- Client strategy: [smysnk.github.io/SikuliGO/client-strategy](https://smysnk.github.io/SikuliGO/client-strategy)
- Architecture docs: [Port Strategy](https://smysnk.github.io/SikuliGO/port-strategy), [gRPC Strategy](https://smysnk.github.io/SikuliGO/grpc-strategy)

## Quickstart

`init:py-examples` prompts for the target directory, creates `.venv`, installs `requirements.txt`, and copies examples.
Each example bootstraps `sikuligo` into `./.sikuligo/bin` and prepends it to PATH for the process.
To install `sikuligo` onto PATH persistently:

```bash
pipx run sikuligo init:py-examples
cd sikuligo-demo
python3 examples/click.py
```

runs:
```python
from __future__ import annotations
from sikuligo import Pattern, Screen

screen = Screen()
try:
    match = screen.click(Pattern("assets/pattern.png").exact())
    print(f"clicked match target at ({match.target_x}, {match.target_y})")
finally:
    screen.close()
```

## Web Dashboard

Launch with `pipx run`:

```bash
pipx run sikuligo sikuligo -listen 127.0.0.1:50051 -admin-listen :8080
```

Open:

- http://127.0.0.1:8080/dashboard

Additional endpoints:

- http://127.0.0.1:8080/healthz
- http://127.0.0.1:8080/metrics
- http://127.0.0.1:8080/snapshot

Install permanently on PATH:

```bash
pipx run sikuligo install-binary
source ~/.zshrc
# or
source ~/.bash_profile
```
