from __future__ import annotations

import os
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

# Auto-detect Docker socket for testcontainers when DOCKER_HOST is not set.
# Rancher Desktop and OrbStack expose the socket at a non-standard path on macOS;
# testcontainers' docker-py client falls back to /var/run/docker.sock by default.
if not os.environ.get("DOCKER_HOST"):
    _CANDIDATE_SOCKETS = [
        Path.home() / ".rd" / "docker.sock",  # Rancher Desktop
        Path.home() / ".orbstack" / "run" / "docker.sock",  # OrbStack
        Path("/var/run/docker.sock"),  # Linux / Docker Desktop (standard)
    ]
    for _sock in _CANDIDATE_SOCKETS:
        if _sock.exists():
            os.environ["DOCKER_HOST"] = f"unix://{_sock}"
            break

# Disable testcontainers Ryuk reaper unconditionally.
# Ryuk requires privileged mounts that are unavailable on Rancher Desktop (macOS)
# and some CI runners. Container cleanup is handled by the Python context manager
# in the fixture, so ryuk is not needed for correctness.
os.environ["TESTCONTAINERS_RYUK_DISABLED"] = "true"
