# Conversion decisions and verification limits

All 36 active controller routes are implemented or preserve the source's explicit empty behavior. The Java project contains 138 source files; 68 DTO/entity/cache data classes have Go declarations, including nullability and database column mappings. Framework proxies/interfaces are consolidated into Go functions, not recreated as empty wrappers. See API.md and SOURCE_COVERAGE.md for the inventory.

This is a source-based port with tested behavior, not a proven identical replacement for every possible production input. No canonical database DDL, production cache snapshot, broker sandbox, or executable golden-response suite was supplied. The bundled runner was used for Java serialization compatibility tests, but the full Java application was not run against production dependencies.

## Source vs api.txt

- `api.txt` documents eight examples, not all endpoints. IDs, dates, and market/contract values in examples are data, not constants.
- The execution message is `Basket executed successfully` by default, matching api.txt. Java's constant says `Basket exeuted successfully`; set `business.execution_message` to that string if clients depend on it.
- The current Java thematic-detail DTO differs from the example: it has `risk`, `subcategory`, `analyst`, methodology/rationale fields, etc. The Go detail endpoint uses the current source's fields and its `total_invst_amt` → `minInvstAmt` mapping. It does not invent values for the older example's absent fields. Null scrip properties from the example remain explicit in map-built responses, although Java's `ScripDetailResponse` also carries `NON_NULL` annotations.
- `disClosedQty` is the Java request property. The differently cased `disclosedQty` used in api.txt is ignored on the basket request, leaving `disClosedQty: null`; upstream order payloads correctly use `disclosedQty`. Unknown input properties are ignored. Numeric strings accepted by Jackson (including quoted LTP) are accepted by the Go DTO decoder.
- Scrip exchange/token/symbol/expiry/format/week-tag come from the contract master. Add responses leave lot size null; retrieval refreshes it from the contract master. Update copies the request lot size. Admin distribution multiplies requested quantities by the contract lot size.
- Some sample nullable values do not follow the shown request under the current source (for example market protection). Go follows the current mapping, not hardcoded sample values.
- Basket response dates use epoch milliseconds. Thematic catalog/detail dates use calendar dates, matching the JDBC `getDate` behavior and sample. Thematic/report field spellings such as `attachement` and `speclizationTag` are retained.

## Preserved business behavior

- Duplicate basket-name checks are scoped to the authenticated user. Personal basket lists exclude research-call baskets and sort by descending basket ID. Responses retain the distinct empty-list/null-result cases.
- Add/update validations require the Java fields, positive integer quantity for add/update, and configured scrip limits. Execute retains the source's presence checks. Basket execution marks a basket executed when a non-null upstream response is received; the legacy contract does not aggregate individual broker rejections.
- Admin vendor authorization, allowed exchanges, expiry requirement, all-device-user fallback, dated basket names, research-idea cache marking, per-user basket/scrip persistence, push flag, and `/temp` behavior are retained.
- Research user visibility includes explicit user mappings and `ALL`. The active research API uses analyst/status filtering and open calls plus closed calls created within seven days by default. Older category/tag/expiry filtering exists separately in `ResearchData`, as it did in the unused DAO path.
- Research scrips are ordered by creation time. Category/subcategory grouping and current source response fields are retained. The current Java SQL does not select remarks/sort-order for the active endpoint, so its null/zero output is retained.
- Thematic catalogs require active, SEND_NOW, Open baskets visible to the user. Details and investment use the Java basket-existence check. The source does not apply the catalog's user-mapping filter to detail/invest requests; Go retains that behavior.
- Thematic detail/catalog scrips use the latest version. Review uses the Java all-version query and token mapping; duplicate tokens across versions cause a failed response, rather than silently changing the Java review algorithm. Review quantity is stored quantity × `lotSize`, not `lots`; totals use request LTP, HALF_UP to two places, a zero balance, unchanged recommended weights, and the configured buffer text.
- Trading hours are weekdays 09:15–15:30 in Asia/Kolkata by default, with exact closing-time boundary behavior. There is no holiday calendar in the source. Thematic orders are tagged `TBK:<id>` and become AMO outside those hours.
- V1 execution saves the first bulk response, the execution journal, a replacement user master, execution details, and BUY/SELL inactivation rules; it refreshes the order book asynchronously. The legacy invest path saves its older journal shape. No automatic broker-order retries were introduced.
- The legacy holdings endpoint aggregates request quantities and lots by basket when the stored response contains an order number. It retains the source's assignment of invested amount, rather than replacing it with a new sum formula.
- Holdings V1 uses active EXECUTED/complete records and latest-version recommended quantities; V0 uses the older EXECUTED-record aggregation. Both retain the source's EXECUTED-only initial basket discovery and lot calculation. Consequently a basket with only `complete` records can be absent: this source inconsistency is deliberately documented, not silently changed.
- Rebalance actions remain Add New, Add More, Reduce, Exit, or null. V0 compares current quantities; V1 compares original recommended quantities. Removed tokens get Exit. Rebalanced quantities are current admin quantities × basket lots. V1 excludes sold/inactive masters.
- Expiry cleanup deletes dates strictly before local midnight, not dates equal to today, and removes related notifications for expired/deleted baskets. Cleanup jobs honor module and job switches.
- NEST's source builds multiple legs but ultimately sends only the first. `business.legacy_nest_first_only: true` retains this. Set false only when intentionally adopting a full-leg request.

## Corrections and platform changes

These changes are explicit rather than hidden behind a claim of byte-for-byte identity:

1. Scrip deletion/update require both basket ownership and membership of the scrip in that basket. Java omitted one or both checks in some paths.
2. Basket/scrip/notification writes use transactions; list updates roll back on failure. Java could partially update a list or leave partially created data. Parent-row locking enforces add limits under concurrent calls. SQL Server requires table hints rather than GORM's generic FOR UPDATE clause.
3. Missing contract entries, malformed/null requests, invalid numeric inputs, missing admin headers, and unsupported cache formats fail with errors rather than null-pointer crashes or empty scrip inserts.
4. The malformed research-with-basket SQL (missing AND) and SQL dialect assumptions are corrected. Dynamic filter values remain bound parameters. No database/schema name is hardcoded in queries.
5. V1 broker result shortages are recorded as failed details rather than indexing beyond the returned results. A submitted order is never retried automatically following a database error. A persistence failure after submission is logged distinctly and still needs operational reconciliation.
6. Order-book matching normalizes trailing colon/semicolon variants consistently for both matching and updating. Contract lookup in order-book binding is functional; Java allocated an empty contract instead of looking it up.
7. Holdings determine the user's version before computing rebalance actions so output does not depend on Java HashMap iteration order. Money calculations use decimal arithmetic; pathological floating-point boundary cases may differ from Java double/BigDecimal(double) or database implicit casts.
8. Background work uses a bounded campaign queue and graceful shutdown. Java repeatedly created ad-hoc thread pools, some of which could stop before all notification batches were submitted. Campaign processing concurrency is configurable; timing/interleaving is not identical to Java's 250-thread per-user fanout.
9. HTTP timeouts, request/response limits, and checked TLS are explicit. Java's global trust-all HTTPS helper is replaced by per-client TLS configuration. The optional compatibility override is off by default.
10. SQL/ClickHouse logs have meaningful elapsed times and bounded, redacted bodies; raw bearer tokens and customer keys are not retained. Audit logging uses bounded synchronous writes rather than shared mutable lists. SQL audit tables must be provisioned separately, as in the supplied Java deployment.
11. The hardcoded OTP and production-device notification demonstration in `utility/Test.java` are not exposed as authentication or production endpoints. Legacy hashing/decryption and other utility logic are available in `internal/utility`.

## Empty and framework-owned behavior

`GET /thematic/basket/get/category/header` calls a Java service that literally returns null. Go returns HTTP 204; there was no category-header algorithm to convert.

`GET /token` returns escaped HTML for the verified username/scope and `refresh_token: false`. Go is a bearer-token service and does not own Quarkus browser sessions/refresh tokens. `/token/logout` returns the source text `You are logged out`; clients must discard their bearer token. It does not revoke the identity provider's token or implement Quarkus's interactive browser-session lifecycle. An interactive OIDC session flow requires a separate product decision; it is not simulated here.

Unused JDBC helpers are consolidated into repository/service functions and GORM operations. The dead malformed `insertScripLogs` path and old experimental query variants are not independently exposed as new HTTP APIs. They are not exercised by any supplied controller. No new production notification test endpoint is added.

## Cache interoperability

The native Hazelcast provider reads default Java Serializable streams (type ID -100) as data, without loading classes or executing Java methods. It supports the supplied contract and customer DTOs, nested login settings, primitive/array/reference values, enum/key representations, Date, ArrayList, and HashMap data. It emits Java-compatible ArrayList<DeviceMappingEntity> values for device mapping writes. Strings/bools use Hazelcast's native serializers. JSON strings/Hazelcast JSON values are also accepted.

Verified against the supplied runner's Hazelcast 5.0.2 and actual compiled DTO classes:

- Java-written contract: expiry, lot size, symbol, and modified-UTF-8 data read by Go.
- Java-written customer: nested settings and key-representation object read by Go.
- Go-written device list: deserialized and checked by the original Java classes on a local Hazelcast member.

Custom Externalizable types, arbitrary custom writeObject protocols, gzip-enabled Java serialization, and different publisher schemas are not generally supported. The source configuration uses the default Java serialization settings. Unknown formats fail explicitly. Parser limits bound stream size/nesting/reference counts.

The optional HTTP cache provider expects:

- `GET <cache.url>/maps/<map>/<key>` → a JSON value, or 404 for a missing entry.
- `PUT <cache.url>/maps/<map>/<key>` with a JSON value → any 2xx.
- Optional `Authorization: Bearer <cache.token>`.

No external cache bridge is required for the tested source format; the HTTP provider is an extension point, not a bundled server.

## Verification performed

- Go build, vet, and race-enabled tests.
- SQLite API tests for the major success/failure/business branches.
- Live local MySQL 8.4, PostgreSQL 17, and SQL Server 2022 migrations and integration scenarios: basket create/add/get/execute/delete, thematic detail/review/V1 invest, holdings, and research retrieval.
- Local Hazelcast/Java round-trip fixture checks described above.
- Annotation-based inventory: 36 Java routes and 36 Go routes with no missing or extra paths; all 138 Java files accounted for.

Production OIDC, live broker execution/margins/order book, production FCM delivery, production ClickHouse schemas, existing database data compatibility, and unusual Java cache object variants still require deployment-specific verification. The tests use mock HTTP servers or disposable local databases and do not place real orders or send real notifications.
