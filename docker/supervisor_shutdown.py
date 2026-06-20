#!/usr/bin/env python3
import os
import signal
import sys


def parse_fields(value: str) -> dict[str, str]:
    return dict(item.split(":", 1) for item in value.strip().split() if ":" in item)


while True:
    sys.stdout.write("READY\n")
    sys.stdout.flush()
    header_line = sys.stdin.readline()
    if not header_line:
        break
    headers = parse_fields(header_line)
    payload = sys.stdin.read(int(headers.get("len", "0")))
    sys.stdout.write("RESULT 2\nOK")
    sys.stdout.flush()

    process_name = parse_fields(payload).get("processname")
    if process_name in {"caddy", "backend"}:
        os.kill(os.getppid(), signal.SIGTERM)
        break
