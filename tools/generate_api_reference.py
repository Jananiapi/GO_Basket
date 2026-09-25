#!/usr/bin/env python3
"""Build the complete client reference from reviewed semantics and smoke responses."""
from pathlib import Path
import json
import re
from api_reference_data import DATA

root = Path(__file__).resolve().parents[1]
routes = json.loads((root/'docs/api-inventory.json').read_text())
smoke = json.loads((root/'docs/smoke-results.json').read_text())
responses = {(r['method'],r['path']):r for r in smoke['routes']}
assert {r['go_action'] for r in routes} == set(DATA), 'Operation documentation differs from route inventory'
assert all(responses.get((r['method'],r['path']),{}).get('smoke')=='pass' for r in routes), 'Missing passing response fixture'
lines = ['# Basket Order Module — complete API reference', '',
         'All **36 active endpoints** in the Go service. Request rules are reviewed against the handlers; response examples come from the local smoke test. Examples use synthetic users, contracts, IDs and dates. Replace them with your deployment values. Examples are independent illustrations, not a sequence to paste unchanged.', '',
         'Companion files: [OpenAPI 3.1](openapi.json), [route inventory](API.md), [smoke report](SMOKE_REPORT.md), [compatibility notes](PARITY.md), [setup guide](../README.md).', '',
         '## Connection and authentication', '',
         '- Default base URL: `http://localhost:9009`. Prepend `server.base_path` if configured. Route spelling and capitalization are significant, including `/V1`.',
         '- JSON requests use `Content-Type: application/json`. All examples below are relative to the base URL.',
         '- Normal routes: `Authorization: Bearer <JWT>`. The authenticated user comes from the configured claim and is uppercased. OIDC or HMAC verification is configurable; OIDC requires user-info lookup by default.',
         '- Admin creation routes: `Authorization: <ADMIN_TOKEN>` (raw value, without adding `Bearer`) plus a valid vendor `apiKey` in the JSON body.',
         '- `DELETE /basketorderapi/deleteExpiredBasket` is public in the supplied `auth.public_paths` configuration. Removing that exception makes JWT authentication required.',
         '- Each route’s module switch must be true. Disabled modules return HTTP 404. The notifications and scheduler switches control background features rather than additional API routes.', '',
         'For the cURL examples, set these variables in your shell:', '', '```sh',
         "export BASE_URL='http://localhost:9009'", "export ACCESS_TOKEN='your-jwt'", "export ADMIN_TOKEN='your-raw-admin-token'", '```', '',
         'No API key or password from the original Java deployment is included in these examples. Execution/investment commands submit orders when pointed at configured live services. Use `make smoke` for the isolated dry run.', '',
         '## Response and input conventions', '',
         'Most endpoints return this envelope; `result` can be an array or null:', '', '```json',
         '{"status":"Ok","message":"Success","result":[]}', '```', '',
         'Business failures usually still return HTTP 200:', '', '```json',
         '{"status":"Not ok","message":"Invalid Parameter","result":[]}', '```', '',
         '- Check both `status` and `message`. Some preserved legacy outcomes use status `Ok` with messages such as `Invalid basket` or `No records found`.',
         '- Execute and both invest routes return an **array of envelopes**, including business failures. They return a summary message rather than individual broker order details.',
         '- Missing/invalid credentials return HTTP 401: generally an envelope with null fields, or `[]` for execute/invest. Certain session/upstream authorization failures also return 401.',
         '- Unregistered/disabled routes return 404. Unhandled panics return 500; most handled database/upstream errors return HTTP 200 with message `Failed`.',
         '- Category header returns HTTP 204 without a body. `/token` returns HTML; `/token/logout` returns plain text.',
         '- DTO field names are case-sensitive. Unknown properties are ignored. Send quantities/prices as strings where shown; Jackson-compatible numeric-string coercions are also supported.',
         '- The request property is **`disClosedQty`**. The differently cased `disclosedQty` is ignored on these DTOs; the service maps to that spelling only in its upstream order payload.',
         '- Timestamp input accepts epoch milliseconds or supported date/date-time strings (use `YYYY-MM-DD` for expiry examples). Date-only and timezone-free date-times are parsed as UTC; cleanup compares against local midnight in `business.timezone`. Basket dates serialize as epoch milliseconds; thematic/report outputs retain their endpoint-specific date/string forms shown below.',
         '- Empty, null and absent values are not interchangeable. Required string fields must be nonblank. Send exactly one JSON value as the body. Research POST endpoints require an object even when no filters are needed (`{}`).',
         '- No pagination query parameters are implemented. Query parameters are not needed for the documented routes. Default request-body limit is 2 MiB (`server.max_body_bytes`).', '',
         '## Endpoint index', '', '| Method | Endpoint | Purpose | Module |', '|---|---|---|---|']
for r in routes:
    d=DATA[r['go_action']]
    lines.append(f"| {r['method']} | [{r['path']}](#op-{r['go_action'].lower()}) | {d['title']} | `{r['module']}` |")
for i,r in enumerate(routes,1):
    a=r['go_action'];d=DATA[a];sample=responses[(r['method'],r['path'])]
    admin=a in ['adminCreate','adminTemp'];public=a=='deleteExpired'
    auth='Raw admin token + vendor apiKey' if admin else 'Public by default; configurable JWT protection' if public else 'Bearer JWT'
    lines += ['', f'<a id="op-{a.lower()}"></a>', '', f"## {i}. {d['title']}", '', f"**`{r['method']} {r['path']}`**", '', f"Authentication: {auth}. Module: `modules.{r['module']}`.", '', d['notes'], '']
    for p in re.findall(r'\{(\w+)\}',r['path']):
        lines += [f'Path parameter: `{p}` — integer '+('personal basket ID.' if p=='basketId' else 'sector report ID.' if a=='sector' else 'thematic basket ID.'), '']
    if d['body'] is not None:
        lines += ['### Request', '']
        if d['fields']:
            lines += ['| Field | JSON type | Requirement | Meaning |','|---|---|---|---|']
            for f,typ,required,desc in d['fields']:
                lines.append(f'| `{f}` | {typ} | {required} | {desc} |')
            lines += ['']
        lines += ['```json',json.dumps(d['body'],indent=2),'```','']
    else:
        lines += ['Request body: none.', '']
    lines += ['### cURL', '', '```sh']
    path=r['path'].replace('{basketId}','2').replace('{id}','1')
    parts=[f'curl --request {r["method"]} "$BASE_URL{path}"']
    if not public:
        parts.append('  --header "Authorization: '+('$ADMIN_TOKEN' if admin else 'Bearer $ACCESS_TOKEN')+'"')
    if d['body'] is not None:
        parts += ["  --header 'Content-Type: application/json'", "  --data-raw '"+json.dumps(d['body'],separators=(',',':'))+"'"]
    lines += [' \\\n'.join(parts), '```', '', f"### Response example — HTTP {sample['http_status']}", '']
    raw=sample['response']
    if raw:
        try:
            value=json.loads(raw);lang='json';raw=json.dumps(value,indent=2)
        except json.JSONDecodeError:
            lang='html' if a=='token' else 'text';raw=raw.strip()
        lines += [f'```{lang}',raw,'```','']
    else:
        lines += ['No response body.', '']
    if d['errors']:
        lines += ['Common business messages: '+', '.join('`'+e+'`' for e in d['errors'])+'. See the shared response conventions for HTTP status handling.', '']
lines += ['## Shared Scrip request fields', '',
          'Used by add/update, admin distribution, personal execution and thematic investment. `scrips` is one object for add/update-one; an array for update-list, execution, investment and admin creation. Review uses only token and ltp from this DTO.', '',
          'For add/update, required trading strings are exchange, token, qty, price, product, transType, priceType, orderType, ret and source; qty must parse as a positive integer. Execution/admin/investment also require tradingSymbol but preserve the source’s presence-only quantity validation. For thematic investment, top-level source replaces nested source. The broker may impose additional value constraints.', '',
          '| Field | JSON type | Usage |', '|---|---|---|']
usage={
'id':'Required positive existing scrip ID for updates; not an order number.',
'sortOrder':'Optional ordering metadata.',
'exchange':'Required trading field; e.g. NSE, BSE, NFO, CDS, MCX. Admin allowed exchanges are configured separately.',
'token':'Required contract token; e.g. "2188".',
'tradingSymbol':'Required for execution/admin/investment; add/update derive the saved value from the contract cache.',
'qty':'Required quantity string. Add/update require a positive integer.',
'price':'Required price string; broker validation still applies.',
'product':'Required, e.g. CNC. No local enum enforcement.',
'transType':'Required, convention BUY/SELL. No local enum enforcement.',
'priceType':'Required, e.g. MKT/LMT. No local enum enforcement.',
'orderType':'Required, e.g. Regular. Thematic orders can become AMO.',
'ret':'Required retention, e.g. DAY.',
'source':'Required; e.g. WEB. Investment uses the top-level source instead.',
'ltp':'Review input price; omitted numeric value defaults to zero. Numeric strings are accepted.',
'version':'Recommendation version used by V1 execution/holdings.',
'lotSize':'Optional string. Admin derives contract lot size; retrieval refreshes it; update can store the supplied value. Distinct from top-level review lotSize and investment lots.',
'disClosedQty':'Optional disclosed quantity; exact casing required.',
'expiry':'Contract expiry. Add/update replace it using the contract cache.',
'expiryDate':'Optional string validity metadata; distinct from top-level admin expiryDate timestamp.',
'validityDays':'Optional validity string. Update retains existing value when blank.',
'createdBy':'Accepted legacy field; personal basket writes use authenticated identity.',
'formattedInsName':'Contract display name; personal add/update derive it from cache.',
'weekTag':'Contract week metadata; personal add/update derive it from cache.',
'userSegment':'Accepted legacy metadata; not used by active order payload mapping.',
'weightage':'Accepted legacy metadata; review uses stored thematic weightage.',
}
models=(root/'internal/model/java_models.go').read_text()
block=re.search(r'type ScripRequestModel struct \{(.*?)\n\}',models,re.S)[1]
for typ,name in re.findall(r'\w+\s+(\S+)\s+`json:"([^",]+)',block):
    jtype='timestamp' if 'Timestamp' in typ else 'string' if 'string' in typ else 'integer' if 'int' in typ else 'number'
    lines.append(f'| `{name}` | {jtype} | {usage.get(name,"Optional trading value, forwarded to upstream orders when supplied.")} |')
lines += ['', '## Compatibility and verification', '',
          'The complete response bodies above retain null fields and source spellings. The smoke fixture validates these shapes for its seeded data; it is not proof of identical results for every production input. `java/api.txt` contains eight examples and differs from some current Java DTOs; the Go service follows the current source except for documented corrections. See [PARITY.md](PARITY.md).', '',
          'There are no additional public routes for scheduler jobs, direct notification sending, vendor provisioning or thematic master creation in the supplied controller inventory. These depend on configured services, stored data or internal repository functions.', '',
          'Regenerate after changing handlers/examples: run `make smoke`, then `make audit`. The documentation generator fails when any registered route is missing reviewed documentation or a passing response fixture.', '']
(root/'docs/API_REFERENCE.md').write_text('\n'.join(lines))
print(f'API_REFERENCE.md: {len(routes)} endpoints with request rules, cURL and recorded responses.')
