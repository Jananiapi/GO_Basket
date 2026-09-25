"""Hand-reviewed operation semantics shared by Markdown and OpenAPI generators."""
SCRIP = {"exchange":"NSE", "token":"2188", "tradingSymbol":"GENCON-EQ", "qty":"1", "price":"43.04", "product":"CNC", "transType":"BUY", "priceType":"MKT", "orderType":"Regular", "ret":"DAY", "source":"WEB"}
UPDATE = {**SCRIP, "id":2, "qty":"2"}
INVEST = {"basketId":1,"lots":2,"basketAction":"BUY","source":"WEB","investmentAmount":43.04,"scrips":[{**SCRIP,"version":1}]}
ADMIN = {"apiKey":"your-vendor-api-key","basketName":"Research","userId":["USER1"],"expiryDate":"2026-09-30","scrips":[SCRIP],"pushNotification":1,"title":"Research basket","message":"New research basket available"}
# Field descriptions explicitly distinguish handler validation from useful input.
DATA = {}
def op(action,title,body,fields,notes,errors=()):
    DATA[action] = dict(title=title,body=body,fields=fields,notes=notes,errors=list(errors))
op('create','Create a personal basket',{'basketName':'Smoke personal'},
   [('basketName','string','Required','Nonblank; unique for the authenticated user.')],
   'Creates the basket and returns the user’s personal basket list, newest ID first.', ['Invalid Parameter','Basket name already exist'])
op('get','List personal baskets',None,[],
   'Returns basketId, basketName, isExecuted (string "0"/"1"), createdOn (epoch milliseconds), and scripCount. Research-call baskets are excluded. An empty list returns status Ok, message No records found, result null.')
op('rename','Rename a personal basket',{'basketId':2,'basketName':'Smoke renamed'},
   [('basketId','integer','Required','Positive ID owned by the caller.'),('basketName','string','Required','Nonblank, unique for the user, including the current name.')],
   'Persists the name and returns the personal basket list.', ['Invalid Parameter','Basket name already exist','Invalid basket'])
for action,title in [('addScrip','Add one scrip'),('updateScrip','Update one scrip'),('updateScripList','Update a list of scrips')]:
    update=action!='addScrip'
    op(action,title,{'basketId':2,'scrips':([UPDATE] if action=='updateScripList' else UPDATE if update else SCRIP)},
       [('basketId','integer','Required','Positive ID owned by the caller.'),('scrips','Scrip[]' if action=='updateScripList' else 'Scrip','Required','See shared Scrip fields below. '+('Each id must identify an existing scrip in this basket.' if update else 'A single object, not an array.'))],
       ('Updates are transactional. Returns status Ok, message Success, result null. This is a full field update, not PATCH; send every required trading field. A missing scrip returns the legacy combination status Ok and message Invalid basket.' if update else 'Adds one scrip; maximum business.max_scrips defaults to 20. Returns all basket scrips. Exchange/token/symbol/expiry/format/week-tag are refreshed from the contract cache; add response lotSize is null.'),
       ['Invalid Parameter','Invalid basket']+([] if update else ['Basket reached the maximum limits']))
op('scrips','Get basket scrips',None,[],
   'Requires basket ownership. Returns scrips ordered by ID and refreshes lotSize from the contract cache. No scrips: status Ok, message No records found, result null.', ['Invalid Parameter','Invalid basket'])
op('deleteScrip','Delete selected basket scrips',{'basketId':2,'scripsId':[2]},
   [('basketId','integer','Required','Positive owned basket ID.'),('scripsId','integer[]','Required','Nonempty list of scrip IDs belonging to this basket.')],
   'Deletes matching scrips and returns the remaining scrip array, which can be empty.', ['Invalid Parameter','Invalid basket','No records found'])
op('delete','Delete a personal basket',None,[],
   'Requires ownership. Deletes the basket and its scrips transactionally; related notifications are deleted when the notification module is enabled. Returns the remaining personal basket list.', ['Invalid Parameter','Invalid basket'])
op('execute','Execute a personal basket',{'basketId':2,'scrips':[SCRIP]},
   [('basketId','integer','Required','Positive owned basket ID.'),('scrips','Scrip[]','Required','Nonempty submitted order list; tradingSymbol and source required per scrip.')],
   'Submits the request’s scrips to the configured bulk-order service, forwarding the bearer token. Does not load the saved basket scrips as the order request. A non-null upstream response marks the basket executed; individual broker rejections are not aggregated. Returns an ARRAY of response envelopes. No automatic order retry.', ['Invalid Parameter','Invalid basket','Failed'])
op('reset','Reset basket execution flag',None,[],
   'Sets isExecuted to "0" for an owned basket and returns the personal basket list. This does not cancel broker orders.', ['Invalid Parameter','Invalid basket'])
op('span','Calculate equity and derivative span margin',[
    {'exchange':'NSE','token':'2188','qty':'2','price':'43.04','transType':'BUY'},
    {'exchange':'NFO','token':'999','qty':'2','price':'10','transType':'BUY'}],
   [(x,'string','Required per item',d) for x,d in [('exchange','NSE/BSE equity is calculated locally; other exchanges use derivative span.'),('token','Contract token; derivatives require a contract-cache entry.'),('qty','Numeric quantity string.'),('price','Numeric price string.'),('transType','BUY/B or SELL/S.')]],
   'The body is an array. Requires the user REST session in cache. Equity margin is qty × price / business.equity_margin_divisor (default 5). Adds derivative span_trade and expo_trade; MCX multiplies quantity by contract lot size. Returns a two-decimal span string. Derivative upstream failure preserves the equity subtotal. Missing session returns HTTP 401.', ['Invalid Parameter'])
op('nestSpan','Calculate NEST basket margin',[{'exchange':'NSE','symbol':'GENCON-EQ','qty':'2'}],
   [('exchange','string','Required per item','Exchange.'),('qty','string','Required per item','Quantity sent as netQty.'),('symbol','string','Supply for upstream','Trading symbol. Token, price and transType are accepted DTO fields but not forwarded by this route.')],
   'Nonempty array body. Uses cached customer credentials for the configured NEST endpoint. business.legacy_nest_first_only defaults to true: only the first leg is sent. Set false to send all legs. Returns upstream spanRequirement as span. Upstream HTTP 401 propagates.', ['Invalid Parameter','Upstream Emsg'])
for action,title in [('adminCreate','Distribute an admin research basket'),('adminTemp','Create temporary admin research baskets')]:
    op(action,title,ADMIN,
       [('apiKey','string','Required for authorization','Must identify a vendor with tpp_authorization = 1.'),('basketName','string','Required','Nonblank; a local date/time suffix is appended.'),('expiryDate','timestamp','Required','Example YYYY-MM-DD; see date conventions.'),('scrips','Scrip[]','Required','Non-null array, maximum 20 by default; tradingSymbol required. Source permits an empty array.'),('userId','string[]','Optional','If omitted/empty, distribute to all distinct device-mapped users.'),('description','string','Optional','Basket description.'),('pushNotification','integer','Optional','1 enables notification processing if the module is enabled.'),('title','string','Optional','Push title.'),('message','string','Optional','Push/notification message.')],
       'Requires raw admin Authorization header plus vendor apiKey. Exchanges allowed by default: NSE, BSE, NFO, CDS. Quantities multiply by contract lot size. Marks researchIdeas cache and queues per-user basket creation. HTTP success means the campaign was queued, not that all background work succeeded. '+('Persists notification records for eligible users but sends no push.' if action=='adminTemp' else 'Persists notification records and sends push for eligible device-mapped users.'),
       ['Your not a vendor','Your not authorized by admin','Invalid Parameter','Scrip is more than maximum size','Invalid Exchange','Notification queue is full'])
op('deleteExpired','Delete expired baskets',None,[],
   'Public by default through auth.public_paths; removing the configured exception requires JWT authentication. Deletes baskets whose expiry_date is strictly before local midnight, associated scrips and (if enabled) notifications. Not restricted to the calling user. No body.', ['No records found','Failed'])
op('expiry','Delete expired basket scrips',None,[],
   'Authenticated maintenance operation across users. Deletes scrips with expiry strictly before local midnight; does not delete their baskets. Returns "<count>-Record Deleted"; zero is a successful result. No body.', ['Failed to deleted'])
op('research','Get research calls',{'analystName':'Smoke Analyst','status':'open'},
   [('analystName','string','Optional','Trimmed analyst filter.'),('status','string','Optional','Case-insensitive status filter; omitted means open calls plus recently closed calls.')],
   'Send {} for default filters. Active calls must map to the user or ALL. Default closed-call window is business.closed_research_days (7). Explicit status removes that default age filter. Groups by category then subcategory; each call includes scripdetails. Other fields in the legacy request DTO do not add filters here.', ['No data found for this user','No data found','Invalid Parameter'])
op('researchBasket','Get research scrips grouped as baskets',{},[],
   'A JSON object body is required; {} is sufficient. Uses active research calls visible to the user or ALL. Groups flattened scrip entries by category/subcategory, adding basketId to each scrip. Calls without scrips are skipped. Request filter fields are not applied by this handler.', ['No data found','Invalid Parameter'])
op('status','List distinct research statuses',None,[],
   'Returns distinct statuses from the research master table. This list is not restricted by user mappings or active status.', ['No data found'])
op('sectors','List active sector reports',None,[],
   'Returns active reports with id, title, attachment, description, type, url and createdOn.', ['No data found'])
op('sector','Get an active sector report',None,[],
   'Returns the selected active report in a one-element result array, with name, audit fields and activeStatus.', ['No data found'])
op('reports','Filter active research/sector reports',{'basketType':'Equity'},
   [('basketType','string','Optional','Matches the report type column; {} returns all active reports.')],
   'Returns report summaries in result; an empty result array is successful.', ['Invalid Parameter'])
op('thematicAll','List available thematic baskets',None,[],
   'Requires active status, action_type SEND_NOW, status Open, and a user/ALL mapping. Scrips use the latest basket version. basketId is a string here. Contract cache supplies pdc when available.', ['No data found for this user'])
op('thematicDetails','Get thematic basket details',None,[],
   'Returns current-source metadata, latest-version scrips and associated active documents. minInvstAmt maps to total_invst_amt. The source checks basket existence here, not catalog visibility. Response field names include subcategory, analyst, risk and minInvstAmt; see example.', ['Basket not found for ID: <id>'])
for action,title in [('invest','Invest in a thematic basket (legacy)'),('investV1','Invest in a thematic basket (V1)')]:
    op(action,title,INVEST,
       [('basketId','integer','Required','Positive existing thematic basket ID.'),('scrips','Scrip[]','Required','Nonempty; tradingSymbol required. Use version for V1 recommendation tracking.'),('source','string','Required','Copies to every submitted scrip, overriding nested source.'),('lots','integer','Required in V1','Positive in V1. Legacy stores 1 regardless of submitted lots.'),('basketAction','string','Required in V1','BUY or SELL drives inactivation logic; handler checks nonblank, not an enum.'),('investmentAmount','number','Optional','Recorded as submitted; not recomputed from prices.')],
       'Forwards bearer token and orders tagged TBK:<basketId>. Outside configured market days/hours (default weekdays 09:15–15:30 Asia/Kolkata), orderType becomes AMO. Returns an ARRAY of response envelopes. '+('Persists execution journal, user master, per-scrip details, BUY/SELL inactivation and queues order-book refresh.' if action=='investV1' else 'Persists the legacy journal used by the legacy holdings endpoint.')+' No automatic retry; a database error after broker submission requires reconciliation.',
       ['Invalid Parameter','Invalid basket','Failed']+(['Invalid Parameter basketAction','Invalid Parameter lots'] if action=='investV1' else []))
op('review','Preview thematic investment',{'basketId':1,'lotSize':3,'scrips':[{'token':'2188','ltp':43.045}]},
   [('basketId','integer','Supply','Existing thematic basket.'),('lotSize','integer','Supply','Quantity multiplier for review; this is not the lots field. Handler does not enforce positivity.'),('scrips','object[]','Required','Nonempty list of token and ltp. Tokens must exist in basket; omitted ltp defaults to zero.')],
   'Does not place orders. Quantity = stored recommended quantity × lotSize. Prices, total and weights use two-decimal rounding. Returns scrips, totalInvested, investmentAmount, balance "0.00" and configured bufferPercentage. All versions are queried; duplicate tokens across versions cause Failed.',
   ['Basket not found.','Scrip details (LTP) must be provided in request.','No scrips found for this basket.','Scrip not found in basket: <token>','Failed'])
op('rebalance','Get thematic rebalance scrips',{'basketId':1},
   [('basketId','integer','Supply','Existing basket with rebalance_available = 1 and a prior execution by this user.')],
   'Returns the latest version from the rebalance scrip table. weightage, marketCap, version and pdc are null in this response. No order placement.', ['Rebalance not available for this user or basket.','No rebalance data available.'])
op('view','Record a thematic basket view',{'basketId':1,'isViewed':1},
   [('basketId','integer','Supply','Thematic basket ID; no existence/positivity validation in this handler.'),('isViewed','integer','Supply','Use 1 to mark viewed; handler does not constrain values.')],
   'Inserts a user execution/view record. Repeated requests insert repeated records. Returns message Basket updated successfully.', ['Invalid Parameter'])
op('category','Get category header (empty source behavior)',None,[],
   'Returns HTTP 204 with no body because the supplied Java service has no category-header implementation.')
op('legacyHoldings','Get legacy thematic holdings',None,[],
   'Aggregates legacy execution journals containing broker order numbers. Returns basketName, userId, investedAmount, createdDate, executedDate, string lotSize and scripList. V1 journals have a different shape and are skipped by this parser. Empty result is successful.')
for action,title in [('holdings','Get thematic holdings and rebalance actions (V0)'),('holdingsV1','Get thematic holdings and rebalance actions (V1)')]:
    op(action,title,None,[],
       'Returns user basket holdings with scripList and rebalancedAction values Add New, Add More, Reduce, Exit or null. '+('V1 compares original recommended quantities to latest recommendations and excludes sold/inactive masters. Aggregation includes EXECUTED/complete records.' if action=='holdingsV1' else 'V0 compares current quantities to latest recommended quantities using EXECUTED records.')+' Both preserve EXECUTED-only initial basket discovery/lot calculation; baskets with only complete records can be absent. Empty result is successful.')
op('token','Show token identity',None,[],
   'Returns escaped HTML with verified username, scopes and refresh_token: false. Does not expose the raw token or create an interactive login session.')
op('logout','Return logout response',None,[],
   'Returns plain text You are logged out. The client must discard its bearer token. This does not revoke provider tokens or implement Quarkus browser-session logout.')
