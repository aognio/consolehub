#!/usr/bin/env python3
"""
ConsoleHub Python Demo Agent
Demonstrates real-time telemetry streaming over JSON-RPC 2.0 WebSockets.
"""

import json
import time

print("ConsoleHub Python Demo Agent")
print("Target endpoint: ws://localhost:8080/api/v1/rpc/ws")

sample_payload = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "process.register",
    "params": {
        "tenant": "acme",
        "app": "python-replicator",
        "host": {"hostname": "vps-02", "display_name": "Backup VPS", "platform": "linux"},
        "process": {
            "client_run_id": "py-uuid-999",
            "pid": 5812,
            "started_at": "2026-07-26T18:50:00Z",
            "version": "1.0.0",
            "command_line": "python3 agent.py",
            "working_directory": "/srv/python-app"
        }
    }
}

print("Sample Registration Payload:")
print(json.dumps(sample_payload, indent=2))
