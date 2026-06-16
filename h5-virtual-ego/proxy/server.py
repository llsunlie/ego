#!/usr/bin/env python3
"""
ego H5 LLM Proxy Server

Routes:
  OPTIONS /chat  — CORS preflight
  POST    /chat  — Proxy LLM chat request (hides API key)
  GET     /health — Health check

Usage:
  编辑 proxy/.env 配置认证和提示词模板，然后：
  python3 proxy/server.py
"""

import json
import os
import sys
import time
import urllib.request
import urllib.error
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path
from socketserver import ThreadingMixIn
from urllib.parse import urlparse

# ── Load .env ───────────────────────────────────────────

def _load_dotenv():
    """Load key=value pairs from .env into os.environ (existing env vars take priority)."""
    env_file = Path(__file__).resolve().parent / ".env"
    if not env_file.exists():
        return
    with open(env_file, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" not in line:
                continue
            key, _, val = line.partition("=")
            key = key.strip()
            val = val.strip().replace("\\n", "\n")
            if key not in os.environ:
                os.environ[key] = val

_load_dotenv()

# ── Config ──────────────────────────────────────────────
API_KEY = os.environ["ANTHROPIC_AUTH_TOKEN"]
MODEL = os.environ.get("ANTHROPIC_MODEL", "deepseek-v4-flash")
PORT = int(os.environ.get("PROXY_PORT", "8090"))
BASE_URL = os.environ.get("ANTHROPIC_BASE_URL", "https://api.deepseek.com/anthropic")
ANTHROPIC_URL = f"{BASE_URL}/messages"

# ── Threaded server ─────────────────────────────────────

class ThreadingHTTPServer(ThreadingMixIn, HTTPServer):
    daemon_threads = True


# ── HTTP Handler ─────────────────────────────────────────

class ProxyHandler(BaseHTTPRequestHandler):

    def _cors(self):
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        self.send_header("Access-Control-Max-Age", "86400")

    def do_OPTIONS(self):
        self.send_response(204)
        self._cors()
        self.end_headers()

    def do_GET(self):
        path = urlparse(self.path).path
        if path == "/health":
            self._json(200, {"status": "ok", "model": MODEL})
        else:
            self._json(404, {"error": "not found"})

    def do_POST(self):
        path = urlparse(self.path).path
        if path != "/chat":
            self._json(404, {"error": "not found"})
            return

        t0 = time.time()

        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length) if length else b"{}"

        try:
            payload = json.loads(body)
        except json.JSONDecodeError:
            self._json(400, {"error": "invalid json"})
            return

        messages = payload.get("messages", [])
        system = payload.get("system", "")
        model = payload.get("model", MODEL)

        if not messages:
            self._json(400, {"error": "messages required"})
            return

        # ── Log 1: incoming ──────────────────────────
        last_msg = messages[-1] if messages else {}
        user_text = last_msg.get("content", "") if last_msg.get("role") == "user" else ""
        print(f"[proxy] 1/4 ← user: {user_text}", file=sys.stderr)
        req_payload = {
            "model": model,
            "max_tokens": 600,
            "system": system,
            "messages": messages,
        }
        req_body = json.dumps(req_payload, ensure_ascii=False).encode("utf-8")

        # ── Log 2: request to LLM ─────────────────────
        req_log = json.dumps(req_payload, ensure_ascii=False, indent=2)
        print(f"[proxy] 2/4 → LLM req:\n{req_log}", file=sys.stderr)

        api_req = urllib.request.Request(
            ANTHROPIC_URL,
            data=req_body,
            headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {API_KEY}",
                "anthropic-version": "2023-06-01",
            },
        )

        try:
            with urllib.request.urlopen(api_req, timeout=30) as resp:
                raw_body = resp.read().decode("utf-8", errors="replace")
                data = json.loads(raw_body)

                elapsed_ms = int((time.time() - t0) * 1000)
                res_log = json.dumps(data, ensure_ascii=False, indent=2)
                print(f"[proxy] 3/4 ← LLM res ({elapsed_ms}ms):\n{res_log}", file=sys.stderr)

                content = ""
                if data.get("content"):
                    for block in data["content"]:
                        if block.get("type") == "text":
                            content += block.get("text", "")

                if not content.strip():
                    content = "（我刚才想得太久了……你能再说一遍吗？）"

                print(f"[proxy] 4/4 → client: {content}", file=sys.stderr)
                self._json(200, {"content": content})

        except urllib.error.HTTPError as e:
            err_body = e.read().decode("utf-8", errors="replace")
            elapsed_ms = int((time.time() - t0) * 1000)
            print(f"[proxy] 3/4 ✗ LLM error {e.code} ({elapsed_ms}ms):\n{err_body}", file=sys.stderr)
            self._json(502, {"error": f"LLM API error: {e.code}"})
        except Exception as e:
            elapsed_ms = int((time.time() - t0) * 1000)
            print(f"[proxy] 3/4 ✗ failed ({elapsed_ms}ms): {e}", file=sys.stderr)
            self._json(502, {"error": str(e)})

    def _json(self, status, data):
        body = json.dumps(data, ensure_ascii=False).encode("utf-8")
        self.send_response(status)
        self._cors()
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, fmt, *args):
        print(f"[proxy] {args[0]}", file=sys.stderr)


# ── Main ─────────────────────────────────────────────────

def main():
    print(f"[proxy] http://localhost:{PORT}  Model: {MODEL}", file=sys.stderr)
    server = ThreadingHTTPServer(("0.0.0.0", PORT), ProxyHandler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n[proxy] Shutting down", file=sys.stderr)
        server.shutdown()


if __name__ == "__main__":
    main()
