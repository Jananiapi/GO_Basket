#!/usr/bin/env python3
"""Run every API over loopback HTTP and write an auditable dry-run report."""
import datetime
import json
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[1]
command = ["go", "test", "-race", "./internal/app", "-run", "^TestAllAPIsSmoke$", "-count=1", "-json"]
result = subprocess.run(command, cwd=root, text=True, capture_output=True)
events = []
for line in result.stdout.splitlines():
    try:
        events.append(json.loads(line))
    except json.JSONDecodeError:
        pass
outcomes = {e["Test"].split("/", 1)[1]: e["Action"] for e in events
            if e.get("Test", "").startswith("TestAllAPIsSmoke/") and e["Action"] in ("pass", "fail", "skip")}
responses = {}
# go test -json may split a long output line across several events.
for output in "".join(e.get("Output", "") for e in events).splitlines():
    if "SMOKE_RESULT " in output:
        row = json.loads(output.split("SMOKE_RESULT ", 1)[1])
        responses[(row["method"], row["path"])] = row
inventory = json.loads((root / "docs/api-inventory.json").read_text())
rows = []
for route in inventory:
    action = route["go_action"]
    response = responses.get((route["method"], route["path"]), {})
    rows.append({**route, "smoke": outcomes.get(action, "not run"),
                 "http_status": response.get("http_status"),
                 "response": response.get("response"),
                 "unauthorized": outcomes.get("unauthorized_" + action, "not run"),
                 "disabled": outcomes.get("disabled_" + action, "not run")})
passed = sum(r["smoke"] == "pass" for r in rows)
unauthorized = sum(r["unauthorized"] == "pass" for r in rows)
disabled = sum(r["disabled"] == "pass" for r in rows)
ok = result.returncode == 0 and passed == len(inventory) and disabled == len(inventory)
stamp = datetime.datetime.now(datetime.timezone.utc).isoformat(timespec="seconds")
report = {"generated_at_utc": stamp, "command": " ".join(command), "passed": ok,
          "api_cases_passed": passed, "unauthorized_cases_passed": unauthorized,
          "disabled_cases_passed": disabled, "routes": rows}
(root / "docs/smoke-results.json").write_text(json.dumps(report, indent=2) + "\n")
lines = ["# API smoke / dry-run report", "", f"Generated: {stamp}", "",
         f"Result: **{'PASS' if ok else 'FAIL'}** — {passed}/{len(inventory)} API cases, "
         f"{unauthorized} missing-credential checks, {disabled} disabled-module checks.", "",
         "Executed with the race detector over actual loopback HTTP. Uses a disposable in-memory "
         "SQLite database, fixture cache and signed HMAC JWTs. All outbound application HTTP is "
         "restricted to a single local mock server; no live orders or notifications are sent.", "",
         "Checks include populated response content, basket/scrip mutations and execution/reset flags, "
         "research visibility, thematic journals/details and holdings, review/margin amounts, admin "
         "background persistence and push behavior, and expiry cleanup. Expected mock totals: three "
         "bulk-order calls, one derivative-span call, one NEST-margin call, one order-book call, and "
         "one notification call. Temporary admin creation must not send a notification.", "",
         "Category-header HTTP 204 preserves the Java empty response. Token/logout checks validate "
         "the current stateless response only. The expired-basket route is public in the supplied "
         "configuration, so its missing-credential check is intentionally omitted.", "",
         "This is a smoke test, not exhaustive input/branch coverage or production integration "
         "certification. This run does not retest server databases, native Hazelcast, production OIDC, "
         "or live broker/FCM services. See PARITY.md and the other integration tests for those boundaries.", "",
         "Rerun from `go/`: `python3 tools/smoke_report.py` (or `make smoke`). "
         "Full response bodies are in [smoke-results.json](smoke-results.json).", "",
         "| Method | API | HTTP | Smoke | Missing credentials | Disabled module |",
         "|---|---|---|---|---|---|"]
for r in rows:
    auth = "Public (configured)" if r["path"] == "/basketorderapi/deleteExpiredBasket" else r["unauthorized"].upper() + " (401)"
    lines.append(f'| {r["method"]} | `{r["path"]}` | {r["http_status"] or "—"} | {r["smoke"].upper()} | {auth} | {r["disabled"].upper()} (404) |')
if not ok:
    lines += ["", "## Failure output", "", "```text", result.stdout[-20000:], result.stderr, "```"]
(root / "docs/SMOKE_REPORT.md").write_text("\n".join(lines) + "\n")
print(f"{'PASS' if ok else 'FAIL'}: {passed}/{len(inventory)} APIs; {unauthorized} authentication checks; {disabled} module checks")
print("Reports: docs/SMOKE_REPORT.md, docs/smoke-results.json")
if not ok:
    print(result.stdout[-12000:])
    print(result.stderr, file=sys.stderr)
sys.exit(0 if ok else 1)
