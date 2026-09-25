# API smoke / dry-run report

Generated: 2026-09-10T09:46:08+00:00

Result: **PASS** — 36/36 API cases, 35 missing-credential checks, 36 disabled-module checks.

Executed with the race detector over actual loopback HTTP. Uses a disposable in-memory SQLite database, fixture cache and signed HMAC JWTs. All outbound application HTTP is restricted to a single local mock server; no live orders or notifications are sent.

Checks include populated response content, basket/scrip mutations and execution/reset flags, research visibility, thematic journals/details and holdings, review/margin amounts, admin background persistence and push behavior, and expiry cleanup. Expected mock totals: three bulk-order calls, one derivative-span call, one NEST-margin call, one order-book call, and one notification call. Temporary admin creation must not send a notification.

Category-header HTTP 204 preserves the Java empty response. Token/logout checks validate the current stateless response only. The expired-basket route is public in the supplied configuration, so its missing-credential check is intentionally omitted.

This is a smoke test, not exhaustive input/branch coverage or production integration certification. This run does not retest server databases, native Hazelcast, production OIDC, or live broker/FCM services. See PARITY.md and the other integration tests for those boundaries.

Rerun from `go/`: `python3 tools/smoke_report.py` (or `make smoke`). Full response bodies are in [smoke-results.json](smoke-results.json).

| Method | API | HTTP | Smoke | Missing credentials | Disabled module |
|---|---|---|---|---|---|
| POST | `/basketorder/add/scrips` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/create` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/delete/scrips` | 200 | PASS | PASS (401) | PASS (404) |
| DELETE | `/basketorder/delete/{basketId}` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/execute` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/basketorder/get` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/basketorder/get/scrips/{basketId}` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/nest/spanmargin` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/rename` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/basketorder/reset/{basketId}` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/spanmargin` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/update/scrips` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorder/update/scrips/list` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorderapi/adminCreate` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/basketorderapi/adminCreate/temp` | 200 | PASS | PASS (401) | PASS (404) |
| DELETE | `/basketorderapi/deleteExpiredBasket` | 200 | PASS | Public (configured) | PASS (404) |
| POST | `/cache/delete/expiry` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/research/get/sector/data/{id}` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/research/getResearchCall` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/research/getResearchWithBasket` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/research/getUniqStatus` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/research/getall/rc/report` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/research/getall/sector/data` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/thematic/basket/get/category/header` | 204 | PASS | PASS (401) | PASS (404) |
| GET | `/thematic/basket/get/{id}` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/thematic/basket/getall` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/thematic/basket/holdings` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/thematic/basket/invest` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/thematic/basket/rebalance/details` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/thematic/basket/report` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/thematic/basket/review` | 200 | PASS | PASS (401) | PASS (404) |
| POST | `/thematic/basket/v1/invest` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/thematic/holdings/get` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/thematic/holdings/get/V1` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/token` | 200 | PASS | PASS (401) | PASS (404) |
| GET | `/token/logout` | 200 | PASS | PASS (401) | PASS (404) |
