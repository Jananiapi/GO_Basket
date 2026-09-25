# Basket Order Module — complete API reference

All **36 active endpoints** in the Go service. Request rules are reviewed against the handlers; response examples come from the local smoke test. Examples use synthetic users, contracts, IDs and dates. Replace them with your deployment values. Examples are independent illustrations, not a sequence to paste unchanged.

Companion files: [OpenAPI 3.1](openapi.json), [route inventory](API.md), [smoke report](SMOKE_REPORT.md), [compatibility notes](PARITY.md), [setup guide](../README.md).

## Connection and authentication

- Default base URL: `http://localhost:9009`. Prepend `server.base_path` if configured. Route spelling and capitalization are significant, including `/V1`.
- JSON requests use `Content-Type: application/json`. All examples below are relative to the base URL.
- Normal routes: `Authorization: Bearer <JWT>`. The authenticated user comes from the configured claim and is uppercased. OIDC or HMAC verification is configurable; OIDC requires user-info lookup by default.
- Admin creation routes: `Authorization: <ADMIN_TOKEN>` (raw value, without adding `Bearer`) plus a valid vendor `apiKey` in the JSON body.
- `DELETE /basketorderapi/deleteExpiredBasket` is public in the supplied `auth.public_paths` configuration. Removing that exception makes JWT authentication required.
- Each route’s module switch must be true. Disabled modules return HTTP 404. The notifications and scheduler switches control background features rather than additional API routes.

For the cURL examples, set these variables in your shell:

```sh
export BASE_URL='http://localhost:9009'
export ACCESS_TOKEN='your-jwt'
export ADMIN_TOKEN='your-raw-admin-token'
```

No API key or password from the original Java deployment is included in these examples. Execution/investment commands submit orders when pointed at configured live services. Use `make smoke` for the isolated dry run.

## Response and input conventions

Most endpoints return this envelope; `result` can be an array or null:

```json
{"status":"Ok","message":"Success","result":[]}
```

Business failures usually still return HTTP 200:

```json
{"status":"Not ok","message":"Invalid Parameter","result":[]}
```

- Check both `status` and `message`. Some preserved legacy outcomes use status `Ok` with messages such as `Invalid basket` or `No records found`.
- Execute and both invest routes return an **array of envelopes**, including business failures. They return a summary message rather than individual broker order details.
- Missing/invalid credentials return HTTP 401: generally an envelope with null fields, or `[]` for execute/invest. Certain session/upstream authorization failures also return 401.
- Unregistered/disabled routes return 404. Unhandled panics return 500; most handled database/upstream errors return HTTP 200 with message `Failed`.
- Category header returns HTTP 204 without a body. `/token` returns HTML; `/token/logout` returns plain text.
- DTO field names are case-sensitive. Unknown properties are ignored. Send quantities/prices as strings where shown; Jackson-compatible numeric-string coercions are also supported.
- The request property is **`disClosedQty`**. The differently cased `disclosedQty` is ignored on these DTOs; the service maps to that spelling only in its upstream order payload.
- Timestamp input accepts epoch milliseconds or supported date/date-time strings (use `YYYY-MM-DD` for expiry examples). Date-only and timezone-free date-times are parsed as UTC; cleanup compares against local midnight in `business.timezone`. Basket dates serialize as epoch milliseconds; thematic/report outputs retain their endpoint-specific date/string forms shown below.
- Empty, null and absent values are not interchangeable. Required string fields must be nonblank. Send exactly one JSON value as the body. Research POST endpoints require an object even when no filters are needed (`{}`).
- No pagination query parameters are implemented. Query parameters are not needed for the documented routes. Default request-body limit is 2 MiB (`server.max_body_bytes`).

## Endpoint index

| Method | Endpoint | Purpose | Module |
|---|---|---|---|
| POST | [/basketorder/add/scrips](#op-addscrip) | Add one scrip | `basket` |
| POST | [/basketorder/create](#op-create) | Create a personal basket | `basket` |
| POST | [/basketorder/delete/scrips](#op-deletescrip) | Delete selected basket scrips | `basket` |
| DELETE | [/basketorder/delete/{basketId}](#op-delete) | Delete a personal basket | `basket` |
| POST | [/basketorder/execute](#op-execute) | Execute a personal basket | `basket` |
| GET | [/basketorder/get](#op-get) | List personal baskets | `basket` |
| GET | [/basketorder/get/scrips/{basketId}](#op-scrips) | Get basket scrips | `basket` |
| POST | [/basketorder/nest/spanmargin](#op-nestspan) | Calculate NEST basket margin | `margin` |
| POST | [/basketorder/rename](#op-rename) | Rename a personal basket | `basket` |
| GET | [/basketorder/reset/{basketId}](#op-reset) | Reset basket execution flag | `basket` |
| POST | [/basketorder/spanmargin](#op-span) | Calculate equity and derivative span margin | `margin` |
| POST | [/basketorder/update/scrips](#op-updatescrip) | Update one scrip | `basket` |
| POST | [/basketorder/update/scrips/list](#op-updatescriplist) | Update a list of scrips | `basket` |
| POST | [/basketorderapi/adminCreate](#op-admincreate) | Distribute an admin research basket | `admin` |
| POST | [/basketorderapi/adminCreate/temp](#op-admintemp) | Create temporary admin research baskets | `admin` |
| DELETE | [/basketorderapi/deleteExpiredBasket](#op-deleteexpired) | Delete expired baskets | `admin` |
| POST | [/cache/delete/expiry](#op-expiry) | Delete expired basket scrips | `cache` |
| GET | [/research/get/sector/data/{id}](#op-sector) | Get an active sector report | `research` |
| POST | [/research/getResearchCall](#op-research) | Get research calls | `research` |
| POST | [/research/getResearchWithBasket](#op-researchbasket) | Get research scrips grouped as baskets | `research` |
| GET | [/research/getUniqStatus](#op-status) | List distinct research statuses | `research` |
| POST | [/research/getall/rc/report](#op-reports) | Filter active research/sector reports | `research` |
| GET | [/research/getall/sector/data](#op-sectors) | List active sector reports | `research` |
| GET | [/thematic/basket/get/category/header](#op-category) | Get category header (empty source behavior) | `thematic` |
| GET | [/thematic/basket/get/{id}](#op-thematicdetails) | Get thematic basket details | `thematic` |
| GET | [/thematic/basket/getall](#op-thematicall) | List available thematic baskets | `thematic` |
| GET | [/thematic/basket/holdings](#op-legacyholdings) | Get legacy thematic holdings | `thematic` |
| POST | [/thematic/basket/invest](#op-invest) | Invest in a thematic basket (legacy) | `thematic` |
| POST | [/thematic/basket/rebalance/details](#op-rebalance) | Get thematic rebalance scrips | `thematic` |
| POST | [/thematic/basket/report](#op-view) | Record a thematic basket view | `thematic` |
| POST | [/thematic/basket/review](#op-review) | Preview thematic investment | `thematic` |
| POST | [/thematic/basket/v1/invest](#op-investv1) | Invest in a thematic basket (V1) | `thematic` |
| GET | [/thematic/holdings/get](#op-holdings) | Get thematic holdings and rebalance actions (V0) | `holdings` |
| GET | [/thematic/holdings/get/V1](#op-holdingsv1) | Get thematic holdings and rebalance actions (V1) | `holdings` |
| GET | [/token](#op-token) | Show token identity | `token` |
| GET | [/token/logout](#op-logout) | Return logout response | `token` |

<a id="op-addscrip"></a>

## 1. Add one scrip

**`POST /basketorder/add/scrips`**

Authentication: Bearer JWT. Module: `modules.basket`.

Adds one scrip; maximum business.max_scrips defaults to 20. Returns all basket scrips. Exchange/token/symbol/expiry/format/week-tag are refreshed from the contract cache; add response lotSize is null.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive ID owned by the caller. |
| `scrips` | Scrip | Required | See shared Scrip fields below. A single object, not an array. |

```json
{
  "basketId": 2,
  "scrips": {
    "exchange": "NSE",
    "token": "2188",
    "tradingSymbol": "GENCON-EQ",
    "qty": "1",
    "price": "43.04",
    "product": "CNC",
    "transType": "BUY",
    "priceType": "MKT",
    "orderType": "Regular",
    "ret": "DAY",
    "source": "WEB"
  }
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/add/scrips" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":2,"scrips":{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"1","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB"}}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "id": 2,
      "basketId": 2,
      "lotSize": null,
      "sortOrder": null,
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "1",
      "price": "43.04",
      "expiry": null,
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "triggerPrice": "",
      "disClosedQty": null,
      "mktProtection": null,
      "target": "0",
      "stopLoss": "0",
      "trailingStopLoss": "",
      "formattedInsName": "GENCON-EQ",
      "weekTag": null,
      "validityDays": null,
      "expiryDate": null
    }
  ]
}
```

Common business messages: `Invalid Parameter`, `Invalid basket`, `Basket reached the maximum limits`. See the shared response conventions for HTTP status handling.


<a id="op-create"></a>

## 2. Create a personal basket

**`POST /basketorder/create`**

Authentication: Bearer JWT. Module: `modules.basket`.

Creates the basket and returns the user’s personal basket list, newest ID first.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketName` | string | Required | Nonblank; unique for the authenticated user. |

```json
{
  "basketName": "Smoke personal"
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/create" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketName":"Smoke personal"}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": 2,
      "basketName": "Smoke personal",
      "createdOn": 1789033567235,
      "isExecuted": "0",
      "scripCount": 0
    },
    {
      "basketId": 1,
      "basketName": "Expired fixture",
      "createdOn": 1789033567233,
      "isExecuted": "0",
      "scripCount": 1
    }
  ]
}
```

Common business messages: `Invalid Parameter`, `Basket name already exist`. See the shared response conventions for HTTP status handling.


<a id="op-deletescrip"></a>

## 3. Delete selected basket scrips

**`POST /basketorder/delete/scrips`**

Authentication: Bearer JWT. Module: `modules.basket`.

Deletes matching scrips and returns the remaining scrip array, which can be empty.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive owned basket ID. |
| `scripsId` | integer[] | Required | Nonempty list of scrip IDs belonging to this basket. |

```json
{
  "basketId": 2,
  "scripsId": [
    2
  ]
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/delete/scrips" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":2,"scripsId":[2]}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": []
}
```

Common business messages: `Invalid Parameter`, `Invalid basket`, `No records found`. See the shared response conventions for HTTP status handling.


<a id="op-delete"></a>

## 4. Delete a personal basket

**`DELETE /basketorder/delete/{basketId}`**

Authentication: Bearer JWT. Module: `modules.basket`.

Requires ownership. Deletes the basket and its scrips transactionally; related notifications are deleted when the notification module is enabled. Returns the remaining personal basket list.

Path parameter: `basketId` — integer personal basket ID.

Request body: none.

### cURL

```sh
curl --request DELETE "$BASE_URL/basketorder/delete/2" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": 1,
      "basketName": "Expired fixture",
      "createdOn": 1789033567233,
      "isExecuted": "0",
      "scripCount": 1
    }
  ]
}
```

Common business messages: `Invalid Parameter`, `Invalid basket`. See the shared response conventions for HTTP status handling.


<a id="op-execute"></a>

## 5. Execute a personal basket

**`POST /basketorder/execute`**

Authentication: Bearer JWT. Module: `modules.basket`.

Submits the request’s scrips to the configured bulk-order service, forwarding the bearer token. Does not load the saved basket scrips as the order request. A non-null upstream response marks the basket executed; individual broker rejections are not aggregated. Returns an ARRAY of response envelopes. No automatic order retry.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive owned basket ID. |
| `scrips` | Scrip[] | Required | Nonempty submitted order list; tradingSymbol and source required per scrip. |

```json
{
  "basketId": 2,
  "scrips": [
    {
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "1",
      "price": "43.04",
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "source": "WEB"
    }
  ]
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/execute" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":2,"scrips":[{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"1","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB"}]}'
```

### Response example — HTTP 200

```json
[
  {
    "status": "Ok",
    "message": "Basket executed successfully",
    "result": null
  }
]
```

Common business messages: `Invalid Parameter`, `Invalid basket`, `Failed`. See the shared response conventions for HTTP status handling.


<a id="op-get"></a>

## 6. List personal baskets

**`GET /basketorder/get`**

Authentication: Bearer JWT. Module: `modules.basket`.

Returns basketId, basketName, isExecuted (string "0"/"1"), createdOn (epoch milliseconds), and scripCount. Research-call baskets are excluded. An empty list returns status Ok, message No records found, result null.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/basketorder/get" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": 2,
      "basketName": "Smoke personal",
      "createdOn": 1789033567235,
      "isExecuted": "0",
      "scripCount": 0
    },
    {
      "basketId": 1,
      "basketName": "Expired fixture",
      "createdOn": 1789033567233,
      "isExecuted": "0",
      "scripCount": 1
    }
  ]
}
```


<a id="op-scrips"></a>

## 7. Get basket scrips

**`GET /basketorder/get/scrips/{basketId}`**

Authentication: Bearer JWT. Module: `modules.basket`.

Requires basket ownership. Returns scrips ordered by ID and refreshes lotSize from the contract cache. No scrips: status Ok, message No records found, result null.

Path parameter: `basketId` — integer personal basket ID.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/basketorder/get/scrips/2" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "id": 2,
      "basketId": 2,
      "lotSize": "1",
      "sortOrder": null,
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "1",
      "price": "43.04",
      "expiry": null,
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "triggerPrice": "",
      "disClosedQty": null,
      "mktProtection": null,
      "target": "0",
      "stopLoss": "0",
      "trailingStopLoss": "",
      "formattedInsName": "GENCON-EQ",
      "weekTag": null,
      "validityDays": null,
      "expiryDate": null
    }
  ]
}
```

Common business messages: `Invalid Parameter`, `Invalid basket`. See the shared response conventions for HTTP status handling.


<a id="op-nestspan"></a>

## 8. Calculate NEST basket margin

**`POST /basketorder/nest/spanmargin`**

Authentication: Bearer JWT. Module: `modules.margin`.

Nonempty array body. Uses cached customer credentials for the configured NEST endpoint. business.legacy_nest_first_only defaults to true: only the first leg is sent. Set false to send all legs. Returns upstream spanRequirement as span. Upstream HTTP 401 propagates.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `exchange` | string | Required per item | Exchange. |
| `qty` | string | Required per item | Quantity sent as netQty. |
| `symbol` | string | Supply for upstream | Trading symbol. Token, price and transType are accepted DTO fields but not forwarded by this route. |

```json
[
  {
    "exchange": "NSE",
    "symbol": "GENCON-EQ",
    "qty": "2"
  }
]
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/nest/spanmargin" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '[{"exchange":"NSE","symbol":"GENCON-EQ","qty":"2"}]'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "span": "123.45"
    }
  ]
}
```

Common business messages: `Invalid Parameter`, `Upstream Emsg`. See the shared response conventions for HTTP status handling.


<a id="op-rename"></a>

## 9. Rename a personal basket

**`POST /basketorder/rename`**

Authentication: Bearer JWT. Module: `modules.basket`.

Persists the name and returns the personal basket list.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive ID owned by the caller. |
| `basketName` | string | Required | Nonblank, unique for the user, including the current name. |

```json
{
  "basketId": 2,
  "basketName": "Smoke renamed"
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/rename" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":2,"basketName":"Smoke renamed"}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": 2,
      "basketName": "Smoke renamed",
      "createdOn": 1789033567235,
      "isExecuted": "0",
      "scripCount": 0
    },
    {
      "basketId": 1,
      "basketName": "Expired fixture",
      "createdOn": 1789033567233,
      "isExecuted": "0",
      "scripCount": 1
    }
  ]
}
```

Common business messages: `Invalid Parameter`, `Basket name already exist`, `Invalid basket`. See the shared response conventions for HTTP status handling.


<a id="op-reset"></a>

## 10. Reset basket execution flag

**`GET /basketorder/reset/{basketId}`**

Authentication: Bearer JWT. Module: `modules.basket`.

Sets isExecuted to "0" for an owned basket and returns the personal basket list. This does not cancel broker orders.

Path parameter: `basketId` — integer personal basket ID.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/basketorder/reset/2" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": 2,
      "basketName": "Smoke renamed",
      "createdOn": 1789033567235,
      "isExecuted": "0",
      "scripCount": 1
    },
    {
      "basketId": 1,
      "basketName": "Expired fixture",
      "createdOn": 1789033567233,
      "isExecuted": "0",
      "scripCount": 1
    }
  ]
}
```

Common business messages: `Invalid Parameter`, `Invalid basket`. See the shared response conventions for HTTP status handling.


<a id="op-span"></a>

## 11. Calculate equity and derivative span margin

**`POST /basketorder/spanmargin`**

Authentication: Bearer JWT. Module: `modules.margin`.

The body is an array. Requires the user REST session in cache. Equity margin is qty × price / business.equity_margin_divisor (default 5). Adds derivative span_trade and expo_trade; MCX multiplies quantity by contract lot size. Returns a two-decimal span string. Derivative upstream failure preserves the equity subtotal. Missing session returns HTTP 401.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `exchange` | string | Required per item | NSE/BSE equity is calculated locally; other exchanges use derivative span. |
| `token` | string | Required per item | Contract token; derivatives require a contract-cache entry. |
| `qty` | string | Required per item | Numeric quantity string. |
| `price` | string | Required per item | Numeric price string. |
| `transType` | string | Required per item | BUY/B or SELL/S. |

```json
[
  {
    "exchange": "NSE",
    "token": "2188",
    "qty": "2",
    "price": "43.04",
    "transType": "BUY"
  },
  {
    "exchange": "NFO",
    "token": "999",
    "qty": "2",
    "price": "10",
    "transType": "BUY"
  }
]
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/spanmargin" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '[{"exchange":"NSE","token":"2188","qty":"2","price":"43.04","transType":"BUY"},{"exchange":"NFO","token":"999","qty":"2","price":"10","transType":"BUY"}]'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "span": "142.22"
    }
  ]
}
```

Common business messages: `Invalid Parameter`. See the shared response conventions for HTTP status handling.


<a id="op-updatescrip"></a>

## 12. Update one scrip

**`POST /basketorder/update/scrips`**

Authentication: Bearer JWT. Module: `modules.basket`.

Updates are transactional. Returns status Ok, message Success, result null. This is a full field update, not PATCH; send every required trading field. A missing scrip returns the legacy combination status Ok and message Invalid basket.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive ID owned by the caller. |
| `scrips` | Scrip | Required | See shared Scrip fields below. Each id must identify an existing scrip in this basket. |

```json
{
  "basketId": 2,
  "scrips": {
    "exchange": "NSE",
    "token": "2188",
    "tradingSymbol": "GENCON-EQ",
    "qty": "2",
    "price": "43.04",
    "product": "CNC",
    "transType": "BUY",
    "priceType": "MKT",
    "orderType": "Regular",
    "ret": "DAY",
    "source": "WEB",
    "id": 2
  }
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/update/scrips" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":2,"scrips":{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"2","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB","id":2}}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": null
}
```

Common business messages: `Invalid Parameter`, `Invalid basket`. See the shared response conventions for HTTP status handling.


<a id="op-updatescriplist"></a>

## 13. Update a list of scrips

**`POST /basketorder/update/scrips/list`**

Authentication: Bearer JWT. Module: `modules.basket`.

Updates are transactional. Returns status Ok, message Success, result null. This is a full field update, not PATCH; send every required trading field. A missing scrip returns the legacy combination status Ok and message Invalid basket.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive ID owned by the caller. |
| `scrips` | Scrip[] | Required | See shared Scrip fields below. Each id must identify an existing scrip in this basket. |

```json
{
  "basketId": 2,
  "scrips": [
    {
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "2",
      "price": "43.04",
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "source": "WEB",
      "id": 2
    }
  ]
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorder/update/scrips/list" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":2,"scrips":[{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"2","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB","id":2}]}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": null
}
```

Common business messages: `Invalid Parameter`, `Invalid basket`. See the shared response conventions for HTTP status handling.


<a id="op-admincreate"></a>

## 14. Distribute an admin research basket

**`POST /basketorderapi/adminCreate`**

Authentication: Raw admin token + vendor apiKey. Module: `modules.admin`.

Requires raw admin Authorization header plus vendor apiKey. Exchanges allowed by default: NSE, BSE, NFO, CDS. Quantities multiply by contract lot size. Marks researchIdeas cache and queues per-user basket creation. HTTP success means the campaign was queued, not that all background work succeeded. Persists notification records and sends push for eligible device-mapped users.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `apiKey` | string | Required for authorization | Must identify a vendor with tpp_authorization = 1. |
| `basketName` | string | Required | Nonblank; a local date/time suffix is appended. |
| `expiryDate` | timestamp | Required | Example YYYY-MM-DD; see date conventions. |
| `scrips` | Scrip[] | Required | Non-null array, maximum 20 by default; tradingSymbol required. Source permits an empty array. |
| `userId` | string[] | Optional | If omitted/empty, distribute to all distinct device-mapped users. |
| `description` | string | Optional | Basket description. |
| `pushNotification` | integer | Optional | 1 enables notification processing if the module is enabled. |
| `title` | string | Optional | Push title. |
| `message` | string | Optional | Push/notification message. |

```json
{
  "apiKey": "your-vendor-api-key",
  "basketName": "Research",
  "userId": [
    "USER1"
  ],
  "expiryDate": "2026-09-30",
  "scrips": [
    {
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "1",
      "price": "43.04",
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "source": "WEB"
    }
  ],
  "pushNotification": 1,
  "title": "Research basket",
  "message": "New research basket available"
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorderapi/adminCreate" \
  --header "Authorization: $ADMIN_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"apiKey":"your-vendor-api-key","basketName":"Research","userId":["USER1"],"expiryDate":"2026-09-30","scrips":[{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"1","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB"}],"pushNotification":1,"title":"Research basket","message":"New research basket available"}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": null
}
```

Common business messages: `Your not a vendor`, `Your not authorized by admin`, `Invalid Parameter`, `Scrip is more than maximum size`, `Invalid Exchange`, `Notification queue is full`. See the shared response conventions for HTTP status handling.


<a id="op-admintemp"></a>

## 15. Create temporary admin research baskets

**`POST /basketorderapi/adminCreate/temp`**

Authentication: Raw admin token + vendor apiKey. Module: `modules.admin`.

Requires raw admin Authorization header plus vendor apiKey. Exchanges allowed by default: NSE, BSE, NFO, CDS. Quantities multiply by contract lot size. Marks researchIdeas cache and queues per-user basket creation. HTTP success means the campaign was queued, not that all background work succeeded. Persists notification records for eligible users but sends no push.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `apiKey` | string | Required for authorization | Must identify a vendor with tpp_authorization = 1. |
| `basketName` | string | Required | Nonblank; a local date/time suffix is appended. |
| `expiryDate` | timestamp | Required | Example YYYY-MM-DD; see date conventions. |
| `scrips` | Scrip[] | Required | Non-null array, maximum 20 by default; tradingSymbol required. Source permits an empty array. |
| `userId` | string[] | Optional | If omitted/empty, distribute to all distinct device-mapped users. |
| `description` | string | Optional | Basket description. |
| `pushNotification` | integer | Optional | 1 enables notification processing if the module is enabled. |
| `title` | string | Optional | Push title. |
| `message` | string | Optional | Push/notification message. |

```json
{
  "apiKey": "your-vendor-api-key",
  "basketName": "Research",
  "userId": [
    "USER1"
  ],
  "expiryDate": "2026-09-30",
  "scrips": [
    {
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "1",
      "price": "43.04",
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "source": "WEB"
    }
  ],
  "pushNotification": 1,
  "title": "Research basket",
  "message": "New research basket available"
}
```

### cURL

```sh
curl --request POST "$BASE_URL/basketorderapi/adminCreate/temp" \
  --header "Authorization: $ADMIN_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"apiKey":"your-vendor-api-key","basketName":"Research","userId":["USER1"],"expiryDate":"2026-09-30","scrips":[{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"1","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB"}],"pushNotification":1,"title":"Research basket","message":"New research basket available"}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": null
}
```

Common business messages: `Your not a vendor`, `Your not authorized by admin`, `Invalid Parameter`, `Scrip is more than maximum size`, `Invalid Exchange`, `Notification queue is full`. See the shared response conventions for HTTP status handling.


<a id="op-deleteexpired"></a>

## 16. Delete expired baskets

**`DELETE /basketorderapi/deleteExpiredBasket`**

Authentication: Public by default; configurable JWT protection. Module: `modules.admin`.

Public by default through auth.public_paths; removing the configured exception requires JWT authentication. Deletes baskets whose expiry_date is strictly before local midnight, associated scrips and (if enabled) notifications. Not restricted to the calling user. No body.

Request body: none.

### cURL

```sh
curl --request DELETE "$BASE_URL/basketorderapi/deleteExpiredBasket"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": null
}
```

Common business messages: `No records found`, `Failed`. See the shared response conventions for HTTP status handling.


<a id="op-expiry"></a>

## 17. Delete expired basket scrips

**`POST /cache/delete/expiry`**

Authentication: Bearer JWT. Module: `modules.cache`.

Authenticated maintenance operation across users. Deletes scrips with expiry strictly before local midnight; does not delete their baskets. Returns "<count>-Record Deleted"; zero is a successful result. No body.

Request body: none.

### cURL

```sh
curl --request POST "$BASE_URL/cache/delete/expiry" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "1-Record Deleted",
  "result": null
}
```

Common business messages: `Failed to deleted`. See the shared response conventions for HTTP status handling.


<a id="op-sector"></a>

## 18. Get an active sector report

**`GET /research/get/sector/data/{id}`**

Authentication: Bearer JWT. Module: `modules.research`.

Returns the selected active report in a one-element result array, with name, audit fields and activeStatus.

Path parameter: `id` — integer sector report ID.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/research/get/sector/data/1" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "activeStatus": 1,
      "attachment": null,
      "createdBy": 0,
      "createdOn": 1789033567232,
      "description": null,
      "id": 1,
      "name": "Smoke report",
      "type": "Equity",
      "updatedBy": 0,
      "updatedOn": 1789033567232,
      "url": null
    }
  ]
}
```

Common business messages: `No data found`. See the shared response conventions for HTTP status handling.


<a id="op-research"></a>

## 19. Get research calls

**`POST /research/getResearchCall`**

Authentication: Bearer JWT. Module: `modules.research`.

Send {} for default filters. Active calls must map to the user or ALL. Default closed-call window is business.closed_research_days (7). Explicit status removes that default age filter. Groups by category then subcategory; each call includes scripdetails. Other fields in the legacy request DTO do not add filters here.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `analystName` | string | Optional | Trimmed analyst filter. |
| `status` | string | Optional | Case-insensitive status filter; omitted means open calls plus recently closed calls. |

```json
{
  "analystName": "Smoke Analyst",
  "status": "open"
}
```

### cURL

```sh
curl --request POST "$BASE_URL/research/getResearchCall" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"analystName":"Smoke Analyst","status":"open"}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "Equity": {
        "Intraday": [
          {
            "analystName": "Smoke Analyst",
            "attachement": "",
            "basketId": 1,
            "createdOn": "2026-09-10 10:00:00",
            "description": null,
            "expiryDate": null,
            "investmentDuration": null,
            "remarks": null,
            "scripdetails": [
              {
                "disClosedQty": null,
                "exchange": "NSE",
                "expiry": "2026-09-30",
                "formattedInsName": null,
                "lotSize": null,
                "mktProtection": null,
                "orderType": null,
                "priceLowerBound": null,
                "priceType": null,
                "priceUpperBound": null,
                "product": null,
                "qty": "2",
                "ret": null,
                "scripId": 1,
                "stopLossLowerBound": null,
                "stopLossUpperBound": null,
                "targetLowerBound": null,
                "targetUpperBound": null,
                "token": "2188",
                "tradingSymbol": null,
                "trailingStopLoss": null,
                "transType": null,
                "triggerPrice": null,
                "validityDays": null,
                "weekTag": null
              }
            ],
            "shortDesc": null,
            "sortOrder": 0,
            "speclizationTag": null,
            "status": "open"
          }
        ]
      }
    }
  ]
}
```

Common business messages: `No data found for this user`, `No data found`, `Invalid Parameter`. See the shared response conventions for HTTP status handling.


<a id="op-researchbasket"></a>

## 20. Get research scrips grouped as baskets

**`POST /research/getResearchWithBasket`**

Authentication: Bearer JWT. Module: `modules.research`.

A JSON object body is required; {} is sufficient. Uses active research calls visible to the user or ALL. Groups flattened scrip entries by category/subcategory, adding basketId to each scrip. Calls without scrips are skipped. Request filter fields are not applied by this handler.

### Request

```json
{}
```

### cURL

```sh
curl --request POST "$BASE_URL/research/getResearchWithBasket" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "Equity": {
        "Intraday": [
          {
            "basketId": 1,
            "disClosedQty": null,
            "exchange": "NSE",
            "expiry": "2026-09-30",
            "expiryDate": null,
            "formattedInsName": null,
            "lotSize": null,
            "mktProtection": null,
            "orderType": null,
            "priceType": null,
            "product": null,
            "qty": "2",
            "ret": null,
            "token": "2188",
            "tradingSymbol": null,
            "trailingStopLoss": null,
            "transType": null,
            "triggerPrice": null,
            "validityDays": null,
            "weekTag": null
          }
        ]
      }
    }
  ]
}
```

Common business messages: `No data found`, `Invalid Parameter`. See the shared response conventions for HTTP status handling.


<a id="op-status"></a>

## 21. List distinct research statuses

**`GET /research/getUniqStatus`**

Authentication: Bearer JWT. Module: `modules.research`.

Returns distinct statuses from the research master table. This list is not restricted by user mappings or active status.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/research/getUniqStatus" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    "open"
  ]
}
```

Common business messages: `No data found`. See the shared response conventions for HTTP status handling.


<a id="op-reports"></a>

## 22. Filter active research/sector reports

**`POST /research/getall/rc/report`**

Authentication: Bearer JWT. Module: `modules.research`.

Returns report summaries in result; an empty result array is successful.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketType` | string | Optional | Matches the report type column; {} returns all active reports. |

```json
{
  "basketType": "Equity"
}
```

### cURL

```sh
curl --request POST "$BASE_URL/research/getall/rc/report" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketType":"Equity"}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "attachment": null,
      "createdOn": 1789033567232,
      "description": null,
      "id": 1,
      "title": "Smoke report",
      "type": "Equity",
      "url": null
    }
  ]
}
```

Common business messages: `Invalid Parameter`. See the shared response conventions for HTTP status handling.


<a id="op-sectors"></a>

## 23. List active sector reports

**`GET /research/getall/sector/data`**

Authentication: Bearer JWT. Module: `modules.research`.

Returns active reports with id, title, attachment, description, type, url and createdOn.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/research/getall/sector/data" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "attachment": null,
      "createdOn": 1789033567232,
      "description": null,
      "id": 1,
      "title": "Smoke report",
      "type": "Equity",
      "url": null
    }
  ]
}
```

Common business messages: `No data found`. See the shared response conventions for HTTP status handling.


<a id="op-category"></a>

## 24. Get category header (empty source behavior)

**`GET /thematic/basket/get/category/header`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Returns HTTP 204 with no body because the supplied Java service has no category-header implementation.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/thematic/basket/get/category/header" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 204

No response body.


<a id="op-thematicdetails"></a>

## 25. Get thematic basket details

**`GET /thematic/basket/get/{id}`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Returns current-source metadata, latest-version scrips and associated active documents. minInvstAmt maps to total_invst_amt. The source checks basket existence here, not catalog visibility. Response field names include subcategory, analyst, risk and minInvstAmt; see example.

Path parameter: `id` — integer thematic basket ID.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/thematic/basket/get/1" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "analyst": null,
      "basketName": "Smoke thematic",
      "benchmark": null,
      "category": null,
      "createdOn": "2026-09-10",
      "curReturn": null,
      "documents": null,
      "expectedReturn": null,
      "expiryDate": null,
      "exposure": null,
      "investmentDuration": null,
      "launchDate": "2026-09-10 15:16:07",
      "longDescription": null,
      "methodology": null,
      "minInvestmentAmount": null,
      "minInvstAmt": 100,
      "overallWeightage": null,
      "rationale": null,
      "reviewFrequency": null,
      "risk": null,
      "riskDescription": null,
      "scrips": [
        {
          "exchange": "NSE",
          "formattedInsName": "GENCON-EQ",
          "holdQty": null,
          "marketCap": null,
          "orderType": "Regular",
          "pdc": null,
          "price": "43.04",
          "priceType": "MKT",
          "qty": "2",
          "token": "2188",
          "tradingSymbol": "GENCON-EQ",
          "transType": "BUY",
          "version": 1,
          "weightage": "100"
        }
      ],
      "shortDescription": null,
      "subcategory": null,
      "tag": null,
      "tagDescription": null,
      "type": null
    }
  ]
}
```

Common business messages: `Basket not found for ID: <id>`. See the shared response conventions for HTTP status handling.


<a id="op-thematicall"></a>

## 26. List available thematic baskets

**`GET /thematic/basket/getall`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Requires active status, action_type SEND_NOW, status Open, and a user/ALL mapping. Scrips use the latest basket version. basketId is a string here. Contract cache supplies pdc when available.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/thematic/basket/getall" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": "1",
      "basketName": "Smoke thematic",
      "category": null,
      "createdOn": "2026-09-10",
      "curReturn": null,
      "expiryDate": null,
      "longDescription": null,
      "minInvestmentAmount": null,
      "scrips": [
        {
          "exchange": "NSE",
          "pdc": null,
          "price": "43.04",
          "qty": "2",
          "token": "2188"
        }
      ],
      "shortDescription": null,
      "subCategory": null,
      "tag": null,
      "totalInvstAmt": 100
    }
  ]
}
```

Common business messages: `No data found for this user`. See the shared response conventions for HTTP status handling.


<a id="op-legacyholdings"></a>

## 27. Get legacy thematic holdings

**`GET /thematic/basket/holdings`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Aggregates legacy execution journals containing broker order numbers. Returns basketName, userId, investedAmount, createdDate, executedDate, string lotSize and scripList. V1 journals have a different shape and are skipped by this parser. Empty result is successful.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/thematic/basket/holdings" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketName": "Smoke thematic",
      "createdDate": "10 Sep 2026 03:16 PM",
      "executedDate": "10 Sep 2026 03:16 PM",
      "investedAmount": 43.04,
      "lotSize": "1",
      "scripList": [
        {
          "exchange": "NSE",
          "qty": "1",
          "token": "2188",
          "tradingSymbol": "GENCON-EQ"
        }
      ],
      "userId": "USER1"
    }
  ]
}
```


<a id="op-invest"></a>

## 28. Invest in a thematic basket (legacy)

**`POST /thematic/basket/invest`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Forwards bearer token and orders tagged TBK:<basketId>. Outside configured market days/hours (default weekdays 09:15–15:30 Asia/Kolkata), orderType becomes AMO. Returns an ARRAY of response envelopes. Persists the legacy journal used by the legacy holdings endpoint. No automatic retry; a database error after broker submission requires reconciliation.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive existing thematic basket ID. |
| `scrips` | Scrip[] | Required | Nonempty; tradingSymbol required. Use version for V1 recommendation tracking. |
| `source` | string | Required | Copies to every submitted scrip, overriding nested source. |
| `lots` | integer | Required in V1 | Positive in V1. Legacy stores 1 regardless of submitted lots. |
| `basketAction` | string | Required in V1 | BUY or SELL drives inactivation logic; handler checks nonblank, not an enum. |
| `investmentAmount` | number | Optional | Recorded as submitted; not recomputed from prices. |

```json
{
  "basketId": 1,
  "lots": 2,
  "basketAction": "BUY",
  "source": "WEB",
  "investmentAmount": 43.04,
  "scrips": [
    {
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "1",
      "price": "43.04",
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "source": "WEB",
      "version": 1
    }
  ]
}
```

### cURL

```sh
curl --request POST "$BASE_URL/thematic/basket/invest" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":1,"lots":2,"basketAction":"BUY","source":"WEB","investmentAmount":43.04,"scrips":[{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"1","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB","version":1}]}'
```

### Response example — HTTP 200

```json
[
  {
    "status": "Ok",
    "message": "Basket executed successfully",
    "result": null
  }
]
```

Common business messages: `Invalid Parameter`, `Invalid basket`, `Failed`. See the shared response conventions for HTTP status handling.


<a id="op-rebalance"></a>

## 29. Get thematic rebalance scrips

**`POST /thematic/basket/rebalance/details`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Returns the latest version from the rebalance scrip table. weightage, marketCap, version and pdc are null in this response. No order placement.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Supply | Existing basket with rebalance_available = 1 and a prior execution by this user. |

```json
{
  "basketId": 1
}
```

### cURL

```sh
curl --request POST "$BASE_URL/thematic/basket/rebalance/details" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":1}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "exchange": "NSE",
      "formattedInsName": "GENCON-EQ",
      "holdQty": null,
      "marketCap": null,
      "orderType": "Regular",
      "pdc": null,
      "price": "43.04",
      "priceType": "MKT",
      "qty": "2",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "transType": "BUY",
      "version": null,
      "weightage": null
    }
  ]
}
```

Common business messages: `Rebalance not available for this user or basket.`, `No rebalance data available.`. See the shared response conventions for HTTP status handling.


<a id="op-view"></a>

## 30. Record a thematic basket view

**`POST /thematic/basket/report`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Inserts a user execution/view record. Repeated requests insert repeated records. Returns message Basket updated successfully.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Supply | Thematic basket ID; no existence/positivity validation in this handler. |
| `isViewed` | integer | Supply | Use 1 to mark viewed; handler does not constrain values. |

```json
{
  "basketId": 1,
  "isViewed": 1
}
```

### cURL

```sh
curl --request POST "$BASE_URL/thematic/basket/report" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":1,"isViewed":1}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Basket updated successfully",
  "result": null
}
```

Common business messages: `Invalid Parameter`. See the shared response conventions for HTTP status handling.


<a id="op-review"></a>

## 31. Preview thematic investment

**`POST /thematic/basket/review`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Does not place orders. Quantity = stored recommended quantity × lotSize. Prices, total and weights use two-decimal rounding. Returns scrips, totalInvested, investmentAmount, balance "0.00" and configured bufferPercentage. All versions are queried; duplicate tokens across versions cause Failed.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Supply | Existing thematic basket. |
| `lotSize` | integer | Supply | Quantity multiplier for review; this is not the lots field. Handler does not enforce positivity. |
| `scrips` | object[] | Required | Nonempty list of token and ltp. Tokens must exist in basket; omitted ltp defaults to zero. |

```json
{
  "basketId": 1,
  "lotSize": 3,
  "scrips": [
    {
      "token": "2188",
      "ltp": 43.045
    }
  ]
}
```

### cURL

```sh
curl --request POST "$BASE_URL/thematic/basket/review" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":1,"lotSize":3,"scrips":[{"token":"2188","ltp":43.045}]}'
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "balance": "0.00",
      "bufferPercentage": "10",
      "investmentAmount": "258.27",
      "scrips": [
        {
          "adjWeightage": "100.00",
          "exchange": "NSE",
          "formattedInsName": "GENCON-EQ",
          "orderType": "Regular",
          "pdc": null,
          "price": "43.05",
          "priceType": "MKT",
          "qty": 6,
          "token": "2188",
          "tradingSymbol": "GENCON-EQ",
          "transType": "BUY",
          "weightage": "100.00"
        }
      ],
      "totalInvested": "258.27"
    }
  ]
}
```

Common business messages: `Basket not found.`, `Scrip details (LTP) must be provided in request.`, `No scrips found for this basket.`, `Scrip not found in basket: <token>`, `Failed`. See the shared response conventions for HTTP status handling.


<a id="op-investv1"></a>

## 32. Invest in a thematic basket (V1)

**`POST /thematic/basket/v1/invest`**

Authentication: Bearer JWT. Module: `modules.thematic`.

Forwards bearer token and orders tagged TBK:<basketId>. Outside configured market days/hours (default weekdays 09:15–15:30 Asia/Kolkata), orderType becomes AMO. Returns an ARRAY of response envelopes. Persists execution journal, user master, per-scrip details, BUY/SELL inactivation and queues order-book refresh. No automatic retry; a database error after broker submission requires reconciliation.

### Request

| Field | JSON type | Requirement | Meaning |
|---|---|---|---|
| `basketId` | integer | Required | Positive existing thematic basket ID. |
| `scrips` | Scrip[] | Required | Nonempty; tradingSymbol required. Use version for V1 recommendation tracking. |
| `source` | string | Required | Copies to every submitted scrip, overriding nested source. |
| `lots` | integer | Required in V1 | Positive in V1. Legacy stores 1 regardless of submitted lots. |
| `basketAction` | string | Required in V1 | BUY or SELL drives inactivation logic; handler checks nonblank, not an enum. |
| `investmentAmount` | number | Optional | Recorded as submitted; not recomputed from prices. |

```json
{
  "basketId": 1,
  "lots": 2,
  "basketAction": "BUY",
  "source": "WEB",
  "investmentAmount": 43.04,
  "scrips": [
    {
      "exchange": "NSE",
      "token": "2188",
      "tradingSymbol": "GENCON-EQ",
      "qty": "1",
      "price": "43.04",
      "product": "CNC",
      "transType": "BUY",
      "priceType": "MKT",
      "orderType": "Regular",
      "ret": "DAY",
      "source": "WEB",
      "version": 1
    }
  ]
}
```

### cURL

```sh
curl --request POST "$BASE_URL/thematic/basket/v1/invest" \
  --header "Authorization: Bearer $ACCESS_TOKEN" \
  --header 'Content-Type: application/json' \
  --data-raw '{"basketId":1,"lots":2,"basketAction":"BUY","source":"WEB","investmentAmount":43.04,"scrips":[{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"1","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","source":"WEB","version":1}]}'
```

### Response example — HTTP 200

```json
[
  {
    "status": "Ok",
    "message": "Basket executed successfully",
    "result": null
  }
]
```

Common business messages: `Invalid Parameter`, `Invalid basket`, `Failed`, `Invalid Parameter basketAction`, `Invalid Parameter lots`. See the shared response conventions for HTTP status handling.


<a id="op-holdings"></a>

## 33. Get thematic holdings and rebalance actions (V0)

**`GET /thematic/holdings/get`**

Authentication: Bearer JWT. Module: `modules.holdings`.

Returns user basket holdings with scripList and rebalancedAction values Add New, Add More, Reduce, Exit or null. V0 compares current quantities to latest recommended quantities using EXECUTED records. Both preserve EXECUTED-only initial basket discovery/lot calculation; baskets with only complete records can be absent. Empty result is successful.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/thematic/holdings/get" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": 1,
      "basketName": "Smoke thematic",
      "currentVersion": 1,
      "executedDate": "2026-09-10 15:16:07.2",
      "investedAmount": 43.04,
      "isRebalanced": 0,
      "lotSize": 2,
      "rebalancedVersion": 1,
      "scripList": [
        {
          "currentQty": 1,
          "exchange": "NSE",
          "executedOn": "2026-09-10 15:16:07.2",
          "executedPrice": 43.04,
          "orgRecoQty": 2,
          "rebalancedAction": null,
          "rebalancedQty": 4,
          "recommendedQty": 2,
          "recommendedVersion": 1,
          "token": "2188",
          "tradingSymbol": "GENCON-EQ",
          "version": 1
        }
      ],
      "userId": "USER1"
    }
  ]
}
```


<a id="op-holdingsv1"></a>

## 34. Get thematic holdings and rebalance actions (V1)

**`GET /thematic/holdings/get/V1`**

Authentication: Bearer JWT. Module: `modules.holdings`.

Returns user basket holdings with scripList and rebalancedAction values Add New, Add More, Reduce, Exit or null. V1 compares original recommended quantities to latest recommendations and excludes sold/inactive masters. Aggregation includes EXECUTED/complete records. Both preserve EXECUTED-only initial basket discovery/lot calculation; baskets with only complete records can be absent. Empty result is successful.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/thematic/holdings/get/V1" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```json
{
  "status": "Ok",
  "message": "Success",
  "result": [
    {
      "basketId": 1,
      "basketName": "Smoke thematic",
      "currentVersion": 1,
      "executedDate": "2026-09-10 15:16:07.2",
      "investedAmount": 43.04,
      "isRebalanced": 0,
      "lotSize": 2,
      "rebalancedVersion": 1,
      "scripList": [
        {
          "currentQty": 1,
          "exchange": "NSE",
          "executedOn": "2026-09-10 15:16:07.2",
          "executedPrice": 43.04,
          "orgRecoQty": 2,
          "rebalancedAction": null,
          "rebalancedQty": 4,
          "recommendedQty": 2,
          "recommendedVersion": 1,
          "token": "2188",
          "tradingSymbol": "GENCON-EQ",
          "version": 1
        }
      ],
      "userId": "USER1"
    }
  ]
}
```


<a id="op-token"></a>

## 35. Show token identity

**`GET /token`**

Authentication: Bearer JWT. Module: `modules.token`.

Returns escaped HTML with verified username, scopes and refresh_token: false. Does not expose the raw token or create an interactive login session.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/token" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```html
<html><body><ul><li>username: USER1</li><li>scopes: basket</li><li>refresh_token: false</li></ul></body></html>
```


<a id="op-logout"></a>

## 36. Return logout response

**`GET /token/logout`**

Authentication: Bearer JWT. Module: `modules.token`.

Returns plain text You are logged out. The client must discard its bearer token. This does not revoke provider tokens or implement Quarkus browser-session logout.

Request body: none.

### cURL

```sh
curl --request GET "$BASE_URL/token/logout" \
  --header "Authorization: Bearer $ACCESS_TOKEN"
```

### Response example — HTTP 200

```text
You are logged out
```

## Shared Scrip request fields

Used by add/update, admin distribution, personal execution and thematic investment. `scrips` is one object for add/update-one; an array for update-list, execution, investment and admin creation. Review uses only token and ltp from this DTO.

For add/update, required trading strings are exchange, token, qty, price, product, transType, priceType, orderType, ret and source; qty must parse as a positive integer. Execution/admin/investment also require tradingSymbol but preserve the source’s presence-only quantity validation. For thematic investment, top-level source replaces nested source. The broker may impose additional value constraints.

| Field | JSON type | Usage |
|---|---|---|
| `id` | integer | Required positive existing scrip ID for updates; not an order number. |
| `sortOrder` | string | Optional ordering metadata. |
| `exchange` | string | Required trading field; e.g. NSE, BSE, NFO, CDS, MCX. Admin allowed exchanges are configured separately. |
| `token` | string | Required contract token; e.g. "2188". |
| `tradingSymbol` | string | Required for execution/admin/investment; add/update derive the saved value from the contract cache. |
| `qty` | string | Required quantity string. Add/update require a positive integer. |
| `price` | string | Required price string; broker validation still applies. |
| `product` | string | Required, e.g. CNC. No local enum enforcement. |
| `transType` | string | Required, convention BUY/SELL. No local enum enforcement. |
| `priceType` | string | Required, e.g. MKT/LMT. No local enum enforcement. |
| `orderType` | string | Required, e.g. Regular. Thematic orders can become AMO. |
| `ret` | string | Required retention, e.g. DAY. |
| `triggerPrice` | string | Optional trading value, forwarded to upstream orders when supplied. |
| `disClosedQty` | string | Optional disclosed quantity; exact casing required. |
| `mktProtection` | string | Optional trading value, forwarded to upstream orders when supplied. |
| `target` | string | Optional trading value, forwarded to upstream orders when supplied. |
| `stopLoss` | string | Optional trading value, forwarded to upstream orders when supplied. |
| `trailingStopLoss` | string | Optional trading value, forwarded to upstream orders when supplied. |
| `createdBy` | string | Accepted legacy field; personal basket writes use authenticated identity. |
| `source` | string | Required; e.g. WEB. Investment uses the top-level source instead. |
| `lotSize` | string | Optional string. Admin derives contract lot size; retrieval refreshes it; update can store the supplied value. Distinct from top-level review lotSize and investment lots. |
| `formattedInsName` | string | Contract display name; personal add/update derive it from cache. |
| `weekTag` | string | Contract week metadata; personal add/update derive it from cache. |
| `expiry` | timestamp | Contract expiry. Add/update replace it using the contract cache. |
| `expiryDate` | string | Optional string validity metadata; distinct from top-level admin expiryDate timestamp. |
| `validityDays` | string | Optional validity string. Update retains existing value when blank. |
| `ltp` | number | Review input price; omitted numeric value defaults to zero. Numeric strings are accepted. |
| `userSegment` | string | Accepted legacy metadata; not used by active order payload mapping. |
| `weightage` | string | Accepted legacy metadata; review uses stored thematic weightage. |
| `version` | integer | Recommendation version used by V1 execution/holdings. |

## Compatibility and verification

The complete response bodies above retain null fields and source spellings. The smoke fixture validates these shapes for its seeded data; it is not proof of identical results for every production input. `java/api.txt` contains eight examples and differs from some current Java DTOs; the Go service follows the current source except for documented corrections. See [PARITY.md](PARITY.md).

There are no additional public routes for scheduler jobs, direct notification sending, vendor provisioning or thematic master creation in the supplied controller inventory. These depend on configured services, stored data or internal repository functions.

Regenerate after changing handlers/examples: run `make smoke`, then `make audit`. The documentation generator fails when any registered route is missing reviewed documentation or a passing response fixture.
