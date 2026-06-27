#!/usr/bin/env python3
"""
Loop Engineering Webhook Server
Receives GitHub webhook events and triggers local codebuddy loops.

Usage: python3 loop-webhook.py [port] [webhook_secret]
  port: default 7777
  webhook_secret: GitHub webhook secret for HMAC verification (required for security)

Requires: codebuddy in PATH, gh authenticated, TKEHUB_API_KEY in env
"""

import http.server
import json
import os
import subprocess
import sys
import threading
import hmac
import hashlib
from datetime import datetime

REPO_DIR = "/data/git/req"
LOG_DIR = "/tmp/loop-webhook-logs"
os.makedirs(LOG_DIR, exist_ok=True)

WEBHOOK_SECRET = sys.argv[2] if len(sys.argv) > 2 else os.environ.get("LOOP_WEBHOOK_SECRET", "")


def log(msg):
    ts = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    line = f"[{ts}] {msg}"
    print(line, flush=True)
    with open(f"{LOG_DIR}/webhook.log", "a") as f:
        f.write(line + "\n")


def verify_signature(body, signature_header):
    """Verify GitHub webhook HMAC signature."""
    if not WEBHOOK_SECRET:
        log("WARNING: No webhook secret set, skipping signature verification")
        return True
    if not signature_header:
        return False
    sha_name, signature = signature_header.split("=", 1)
    if sha_name != "sha256":
        return False
    expected = hmac.new(
        WEBHOOK_SECRET.encode(), body, hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)


def run_loop(loop_type, payload):
    """Trigger a codebuddy loop based on the event type."""
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    log_file = f"{LOG_DIR}/{loop_type}-{timestamp}.log"

    prompts = {
        "issue": lambda: (
            f'Use the issue-triager agent to triage issue #{payload.get("issue", {}).get("number", "?")}: '
            f'{payload.get("issue", {}).get("title", "")}. '
            f'Read STATE.md, classify the issue, apply labels with gh issue edit, '
            f'post initial response if needed. Update STATE.md.'
        ),
        "pr": lambda: (
            f'Use the pr-reviewer agent to review PR #{payload.get("pull_request", {}).get("number", "?")} '
            f'(action: {payload.get("action", "")}). '
            f'Read STATE.md, fetch PR diff with gh pr view and gh pr diff, '
            f'apply pr-review skill checklist. Post review with gh pr review. Update STATE.md.'
        ),
        "ci_failure": lambda: (
            f'Use the ci-fixer agent to check for failed CI runs. '
            f'Use gh run list --status failure --limit 3 --workflow ci.yml. '
            f'For each failure, analyze and fix. Update STATE.md.'
        ),
    }

    if loop_type not in prompts:
        return

    prompt = prompts[loop_type]()

    def run():
        with open(log_file, "w") as f:
            f.write(f"Starting {loop_type} loop at {datetime.now()}\n")
            f.write(f"Prompt: {prompt}\n\n")
            f.flush()
            try:
                proc = subprocess.run(
                    ["codebuddy", "-p", "-y",
                     "--allowedTools", "Read,Bash,Grep,Edit,WebFetch",
                     prompt],
                    cwd=REPO_DIR,
                    stdout=f, stderr=subprocess.STDOUT,
                    timeout=600,
                )
                f.write(f"\nExit code: {proc.returncode}\n")
            except subprocess.TimeoutExpired:
                f.write("\nTimeout after 600s\n")
            except Exception as e:
                f.write(f"\nError: {e}\n")

    threading.Thread(target=run, daemon=True).start()
    log(f"Started {loop_type} loop (log: {log_file})")


class WebhookHandler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length)
        event = self.headers.get("X-GitHub-Event", "")
        signature = self.headers.get("X-Hub-Signature-256", "")

        if not verify_signature(body, signature):
            log(f"REJECTED: invalid signature for {event} event")
            self.send_response(403)
            self.end_headers()
            self.wfile.write(b"invalid signature\n")
            return

        payload = json.loads(body) if body else {}
        log_msg = f"Received {event} event"

        if event == "issues" and payload.get("action") in ("opened", "reopened"):
            log_msg += f": issue #{payload.get('issue', {}).get('number')}"
            run_loop("issue", payload)
        elif event == "pull_request" and payload.get("action") in ("opened", "synchronize"):
            log_msg += f": PR #{payload.get('pull_request', {}).get('number')}"
            run_loop("pr", payload)
        elif event == "workflow_run":
            conclusion = payload.get("workflow_run", {}).get("conclusion", "")
            name = payload.get("workflow_run", {}).get("name", "")
            if conclusion == "failure" and name == "CI":
                log_msg += ": CI failed, triggering ci-fixer"
                run_loop("ci_failure", payload)
            else:
                log_msg += f" (conclusion={conclusion}, name={name}, ignored)"
        elif event == "ping":
            log_msg += ": ping received"
        else:
            log_msg += f" (action: {payload.get('action', 'n/a')}, ignored)"

        log(log_msg)
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.end_headers()
        self.wfile.write(b"ok\n")

    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.end_headers()
        self.wfile.write(b"loop-webhook server running\n")

    def log_message(self, format, *args):
        pass  # Suppress default logging


if __name__ == "__main__":
    PORT = int(sys.argv[1]) if len(sys.argv) > 1 else 7777
    server = http.server.HTTPServer(("0.0.0.0", PORT), WebhookHandler)
    log(f"Loop webhook server listening on :{PORT}")
    log(f"Webhook secret: {'set' if WEBHOOK_SECRET else 'NOT SET (insecure!)'}")
    log(f"Repo dir: {REPO_DIR}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        log("Shutting down")
        server.shutdown()
