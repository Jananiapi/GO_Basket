# API inventory

Generated from active Java JAX-RS annotations. `python3 tools/audit_parity.py` fails if a Java route lacks a Go route or a Go route lacks a Java counterpart. This checks route presence, not behavioral equivalence.

| Method | Path | Module | Java method |
|---|---|---|---|
| POST | `/basketorder/add/scrips` | `basket` | `addBasketScrip` |
| POST | `/basketorder/create` | `basket` | `createBasketName` |
| POST | `/basketorder/delete/scrips` | `basket` | `deleteBasketScrip` |
| DELETE | `/basketorder/delete/{basketId}` | `basket` | `deleteBasket` |
| POST | `/basketorder/execute` | `basket` | `excuteBasketOrder` |
| GET | `/basketorder/get` | `basket` | `getBasketNames` |
| GET | `/basketorder/get/scrips/{basketId}` | `basket` | `retrieveScrips` |
| POST | `/basketorder/nest/spanmargin` | `margin` | `getBasketMargin` |
| POST | `/basketorder/rename` | `basket` | `updateBasketName` |
| GET | `/basketorder/reset/{basketId}` | `basket` | `resetExecutionStatus` |
| POST | `/basketorder/spanmargin` | `margin` | `getSpanMargin` |
| POST | `/basketorder/update/scrips` | `basket` | `updateBasketScrip` |
| POST | `/basketorder/update/scrips/list` | `basket` | `updateBasketScripList` |
| POST | `/basketorderapi/adminCreate` | `admin` | `adminCreateBasketName` |
| POST | `/basketorderapi/adminCreate/temp` | `admin` | `adminCreateBasketNameTemp` |
| DELETE | `/basketorderapi/deleteExpiredBasket` | `admin` | `deleteExpiredBasket` |
| POST | `/cache/delete/expiry` | `cache` | `deleteExpiry` |
| GET | `/research/get/sector/data/{id}` | `research` | `getSectorDetails` |
| POST | `/research/getResearchCall` | `research` | `getResearchCall` |
| POST | `/research/getResearchWithBasket` | `research` | `getResWithBasketDetails` |
| GET | `/research/getUniqStatus` | `research` | `getUniqStatus` |
| POST | `/research/getall/rc/report` | `research` | `getResearchReport` |
| GET | `/research/getall/sector/data` | `research` | `getAllSector` |
| GET | `/thematic/basket/get/category/header` | `thematic` | `getCatSubCatFrUser` |
| GET | `/thematic/basket/get/{id}` | `thematic` | `getDetailsThematicBasket` |
| GET | `/thematic/basket/getall` | `thematic` | `getAllThematicBasket` |
| GET | `/thematic/basket/holdings` | `thematic` | `getBasketHoldings` |
| POST | `/thematic/basket/invest` | `thematic` | `executeThematic` |
| POST | `/thematic/basket/rebalance/details` | `thematic` | `rebalanceThematic` |
| POST | `/thematic/basket/report` | `thematic` | `thematicReport` |
| POST | `/thematic/basket/review` | `thematic` | `reviewThematic` |
| POST | `/thematic/basket/v1/invest` | `thematic` | `executeThematicV1` |
| GET | `/thematic/holdings/get` | `holdings` | `getUserRebalancedBaskets` |
| GET | `/thematic/holdings/get/V1` | `holdings` | `getUserRebalancedBasketsV1` |
| GET | `/token` | `token` | `getTokens` |
| GET | `/token/logout` | `token` | `logout` |
