#!/usr/bin/env python3
"""
lupo-proxy-shim.py — bridges Lupo C2 and the Messenger tunneling library.

Communication protocol: newline-delimited JSON on stdin (commands in) and
stdout (events out).  stderr is intentionally unused so Go can read clean
JSON from stdout without interference.

Commands (Go → shim):
    {"cmd": "status"}
    {"cmd": "socks",        "port": 1080}
    {"cmd": "stop_socks",   "port": 1080}
    {"cmd": "forward",      "config": "lhost:lport:dhost:dport"}
    {"cmd": "stop_forward", "config": "lhost:lport:dhost:dport"}
    {"cmd": "stop"}

Events (shim → Go):
    {"event": "ready",              "enc_key": "...", "server_url": "...", "server_port": N}
    {"event": "messenger_connected","messenger_id": "..."}
    {"event": "socks_started",      "host": "127.0.0.1", "port": N}
    {"event": "forward_started",    "config": "..."}
    {"event": "client_built",       "path": "..."}
    {"event": "status",             "messengers_connected": N, "forwarders": [...]}
    {"event": "error",              "message": "..."}
    {"event": "stopped"}
"""

import argparse
import asyncio
import json
import subprocess
import sys
from pathlib import Path

from messenger.engine import Engine
from messenger.generator import generate_encryption_key, generate_hash
from messenger.http_ws_server import HTTPWSServer
from messenger.forwarders import (
    LocalPortForwarder,
    SocksProxy,
    InvalidConfigError,
)
from messenger.messengers import Messenger


# ---------------------------------------------------------------------------
# Minimal UpdateCLI — discards all interactive output so stdout stays clean
# for JSON events.
# ---------------------------------------------------------------------------

class SilentCLI:
    debug_level = 0

    def display(self, _stdout, _status="standard", reprompt=False, debug_level=0):
        pass


# ---------------------------------------------------------------------------
# Engine subclass that fires a callback when a new Messenger client connects.
# ---------------------------------------------------------------------------

class LupoEngine(Engine):
    def __init__(self, messengers, update_cli, encryption_key_hash, on_connect):
        super().__init__(messengers, update_cli, encryption_key_hash)
        self._on_connect = on_connect

    def add_messenger(self, messenger: Messenger):
        result = super().add_messenger(messenger)
        self._on_connect(messenger.identifier)
        return result


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def emit(event: dict) -> None:
    """Write a single JSON event line to stdout and flush immediately."""
    print(json.dumps(event), flush=True)


def build_status(messengers: list) -> dict:
    forwarders = []
    for m in messengers:
        for f in m.forwarders:
            forwarders.append({
                "messenger_id": m.identifier,
                "type": f.NAME,
                "listening_host": str(f.listening_host),
                "listening_port": str(f.listening_port),
                "destination_host": str(f.destination_host),
                "destination_port": str(f.destination_port),
            })
    return {
        "event": "status",
        "messengers_connected": len(messengers),
        "forwarders": forwarders,
    }


# ---------------------------------------------------------------------------
# Async stdin command reader
# ---------------------------------------------------------------------------

async def read_commands(messengers: list, enc_key: str, server_url: str) -> None:
    loop = asyncio.get_event_loop()
    reader = asyncio.StreamReader()
    transport, _ = await loop.connect_read_pipe(
        lambda: asyncio.StreamReaderProtocol(reader), sys.stdin
    )
    cli = SilentCLI()

    while True:
        raw = await reader.readline()
        if not raw:
            break

        try:
            cmd = json.loads(raw.decode().strip())
        except (json.JSONDecodeError, ValueError):
            emit({"event": "error", "message": "invalid JSON command"})
            continue

        action = cmd.get("cmd", "")

        if action == "status":
            emit(build_status(messengers))

        elif action == "socks":
            if not messengers:
                emit({"event": "error", "message": "no messenger connected"})
                continue
            port = int(cmd.get("port", 1080))
            messenger = messengers[-1]
            config = f"127.0.0.1:{port}"
            try:
                forwarder = SocksProxy(messenger, config, cli)
                success = await forwarder.start()
                if success:
                    messenger.forwarders.append(forwarder)
                    emit({"event": "socks_started", "host": "127.0.0.1", "port": port})
                else:
                    emit({"event": "error", "message": f"failed to bind SOCKS on port {port}"})
            except InvalidConfigError as exc:
                emit({"event": "error", "message": str(exc)})

        elif action == "forward":
            if not messengers:
                emit({"event": "error", "message": "no messenger connected"})
                continue
            config = cmd.get("config", "")
            if not config:
                emit({"event": "error", "message": "forward requires a config string"})
                continue
            messenger = messengers[-1]
            try:
                forwarder = LocalPortForwarder(messenger, config, cli)
                success = await forwarder.start()
                if success:
                    messenger.forwarders.append(forwarder)
                    emit({"event": "forward_started", "config": config})
                else:
                    emit({"event": "error", "message": f"failed to start forward: {config}"})
            except InvalidConfigError as exc:
                emit({"event": "error", "message": str(exc)})

        elif action == "build_client":
            output = cmd.get("output", "/tmp/messenger-client.py")
            client_enc_key = cmd.get("enc_key", enc_key)
            client_server_url = cmd.get("server_url", server_url)
            builder_bin = Path(sys.executable).parent / "messenger-builder"
            try:
                result = subprocess.run(
                    [
                        str(builder_bin), "python",
                        "--encryption-key", client_enc_key,
                        "--server-url", client_server_url,
                        "--name", output,
                    ],
                    capture_output=True,
                    text=True,
                )
                if result.returncode == 0:
                    emit({"event": "client_built", "path": output})
                else:
                    msg = (result.stderr or result.stdout or "unknown build error").strip()
                    emit({"event": "error", "message": msg})
            except Exception as exc:
                emit({"event": "error", "message": str(exc)})

        elif action == "stop_socks":
            port = int(cmd.get("port", 0))
            stopped = False
            for m in messengers:
                for f in list(m.forwarders):
                    if isinstance(f, SocksProxy) and f.listening_port == port:
                        await f.stop()
                        m.forwarders.remove(f)
                        stopped = True
                        break
                if stopped:
                    break
            if stopped:
                emit({"event": "socks_stopped", "port": port})
            else:
                emit({"event": "error", "message": f"no SOCKS listener on port {port}"})

        elif action == "stop_forward":
            config = cmd.get("config", "")
            stopped = False
            for m in messengers:
                for f in list(m.forwarders):
                    if (not isinstance(f, SocksProxy) and
                            f"{f.listening_host}:{f.listening_port}:{f.destination_host}:{f.destination_port}" == config):
                        await f.stop()
                        m.forwarders.remove(f)
                        stopped = True
                        break
                if stopped:
                    break
            if stopped:
                emit({"event": "forward_stopped", "config": config})
            else:
                emit({"event": "error", "message": f"no forward matching config: {config}"})

        elif action == "stop":
            emit({"event": "stopped"})
            sys.exit(0)

        else:
            emit({"event": "error", "message": f"unknown command: {action}"})


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

async def run(args: argparse.Namespace) -> None:
    cli = SilentCLI()
    messengers: list = []

    enc_key = args.encryption_key or generate_encryption_key()
    key_hash = generate_hash(enc_key)

    def on_connect(messenger_id: str) -> None:
        emit({"event": "messenger_connected", "messenger_id": messenger_id})

    engine = LupoEngine(messengers, cli, key_hash, on_connect)
    server = HTTPWSServer(cli, engine, ip=args.lhost, port=args.server_port)

    await server.start()

    server_url = f"{args.lhost}:{args.server_port}"
    emit({
        "event": "ready",
        "enc_key": enc_key,
        "server_url": server_url,
        "server_port": args.server_port,
    })

    await read_commands(messengers, enc_key, server_url)


def main() -> None:
    parser = argparse.ArgumentParser(description="Lupo proxy shim for Messenger")
    parser.add_argument("--server-port", type=int, default=8080,
                        help="Port for the Messenger HTTP/WS server")
    parser.add_argument("--lhost", default="0.0.0.0",
                        help="Interface to bind the Messenger server on")
    parser.add_argument("--encryption-key", default=None,
                        help="AES encryption key (auto-generated if omitted)")
    args = parser.parse_args()
    asyncio.run(run(args))


if __name__ == "__main__":
    main()
