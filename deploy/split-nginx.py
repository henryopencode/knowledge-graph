#!/usr/bin/env python3
"""Detach only the graph/API virtual hosts from a legacy unified config."""

import re
import sys
from pathlib import Path

HOSTS = {"api.henryweb.site", "knowledge.henryweb.site"}


def detach(source: str) -> str:
    def server(match: re.Match) -> str:
        block = match.group()
        names = re.search(r"server_name\s+([^;]+);", block)
        if not names:
            return block
        original = names.group(1).split()
        remaining = [name for name in original if name not in HOSTS]
        if remaining == original:
            return block
        if not remaining:
            return ""
        return block[: names.start(1)] + " ".join(remaining) + block[names.end(1) :]

    # The legacy generated configuration has top-level server blocks at column 0.
    result = re.sub(r"(?ms)^server\s*\{.*?^\}\s*", server, source)
    names = re.findall(r"server_name\s+([^;]+);", result)
    if any(HOSTS.intersection(line.split()) for line in names):
        raise ValueError("Unrecognized legacy Nginx layout; original config preserved")
    return result


if __name__ == "__main__":
    path = Path(sys.argv[1])
    source = path.read_text()
    result = detach(source)
    if result != source:
        path.write_text(result)
