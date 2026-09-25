package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testApp(t *testing.T) *App {
	t.Helper()
	c, e := config.Load("../../config.yaml")
	if e != nil { // Test defaults are explicit; no service credentials are needed.
		c.Database.Driver = "sqlite"
		c.Database.DSN = "file:test?mode=memory&cache=shared"
	}
	c.Auth.Enabled = false
	c.Auth.DevUser = "USER1"
	c.Auth.AdminToken = "admin-test"
	c.Cache.Provider = "file"
	c.Logging.Access = false
	c.Modules.Scheduler = false
	return newTestApp(t, c)
}
func newTestApp(t *testing.T, c config.Config) *App {
	t.Helper()
	db, e := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if e != nil {
		t.Fatal(e)
	}
	sql, _ := db.DB()
	sql.SetMaxOpenConns(1)
	if e = Migrate(db); e != nil {
		t.Fatal(e)
	}
	cache := &MemoryCache{Values: map[string]map[string]json.RawMessage{}}
	a, e := New(context.Background(), c, db, cache)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { a.Close(context.Background()); sql.Close() })
	a.Now = func() time.Time { return time.Date(2026, 9, 10, 10, 0, 0, 0, a.location) }
	contract := model.ContractMasterModel{Exch: ptr("NSE"), Token: ptr("2188"), TradingSymbol: ptr("GENCON-EQ"), FormattedInsName: ptr("GENCON-EQ"), LotSize: ptr("1")}
	cache.Put(context.Background(), c.Cache.Maps["contracts"], "NSE_2188", contract)
	cache.Put(context.Background(), c.Cache.Maps["sessions"], "USER1_REST_SESSION", "session")
	return a
}
func request(t *testing.T, a *App, method, path, body string) (int, any) {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Authorization", "Bearer test")
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	if w.Code == 204 {
		return w.Code, nil
	}
	var out any
	if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil {
		t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
	}
	return w.Code, out
}
func ok(t *testing.T, out any) M {
	t.Helper()
	m, yes := out.(map[string]any)
	if !yes || m["status"] != "Ok" {
		t.Fatalf("expected Ok, got %#v", out)
	}
	return m
}
func create(t *testing.T, a *App, name string) int {
	t.Helper()
	_, out := request(t, a, "POST", "/basketorder/create", `{"basketName":"`+name+`"}`)
	m := ok(t, out)
	return integer(m["result"].([]any)[0].(map[string]any)["basketId"])
}

const scripJSON = `{"exchange":"NSE","token":"2188","tradingSymbol":"GENCON-EQ","qty":"1","price":"43.04","product":"CNC","transType":"BUY","priceType":"MKT","orderType":"Regular","ret":"DAY","triggerPrice":"","target":"0","stopLoss":"0","trailingStopLoss":"","source":"WEB","expiry":"2026-09-15"}`

func TestBasketLifecycle(t *testing.T) {
	a := testApp(t)
	id := create(t, a, "API testing")
	_, out := request(t, a, "POST", "/basketorder/create", `{"basketName":"API testing"}`)
	if out.(map[string]any)["message"] != "Basket name already exist" {
		t.Fatal(out)
	}
	_, out = request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": id, "scrips": json.RawMessage(scripJSON)}))
	m := ok(t, out)
	s := m["result"].([]any)[0].(map[string]any)
	if s["expiry"] != nil || s["lotSize"] != nil || s["formattedInsName"] != "GENCON-EQ" {
		t.Fatal(s)
	}
	sid := integer(s["id"])
	_, out = request(t, a, "GET", "/basketorder/get/scrips/"+str(id), "")
	s = ok(t, out)["result"].([]any)[0].(map[string]any)
	if s["lotSize"] != "1" {
		t.Fatal(s)
	}
	var update M
	json.Unmarshal([]byte(scripJSON), &update)
	update["id"] = sid
	update["qty"] = "2"
	_, out = request(t, a, "POST", "/basketorder/update/scrips", jsonString(M{"basketId": id, "scrips": update}))
	if ok(t, out)["result"] != nil {
		t.Fatal(out)
	}
	_, out = request(t, a, "POST", "/basketorder/rename", jsonString(M{"basketId": id, "basketName": "renamed"}))
	ok(t, out)
	_, out = request(t, a, "POST", "/basketorder/delete/scrips", jsonString(M{"basketId": id, "scripsId": []int{sid}}))
	if len(ok(t, out)["result"].([]any)) != 0 {
		t.Fatal(out)
	}
	_, out = request(t, a, "DELETE", "/basketorder/delete/"+str(id), "")
	ok(t, out)
}
func TestBasketOwnershipAndLimit(t *testing.T) {
	a := testApp(t)
	id := create(t, a, "owned")
	a.Config.Business.MaxScrips = 1
	for i := 0; i < 2; i++ {
		_, out := request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": id, "scrips": json.RawMessage(scripJSON)}))
		if i == 0 {
			ok(t, out)
		} else if out.(map[string]any)["message"] != "Basket reached the maximum limits" {
			t.Fatal(out)
		}
	}
	a.auth.config.DevUser = "USER2"
	_, out := request(t, a, "POST", "/basketorder/delete/scrips", jsonString(M{"basketId": id, "scripsId": []int{1}}))
	if out.(map[string]any)["message"] != "Invalid basket" {
		t.Fatal(out)
	}
}
func TestExecuteAndReset(t *testing.T) {
	a := testApp(t)
	id := create(t, a, "exec")
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer test" {
			t.Error("missing bearer")
		}
		var orders []M
		json.NewDecoder(r.Body).Decode(&orders)
		if len(orders) != 1 || orders[0]["tradingSymbol"] != "GENCON-EQ" {
			t.Error(orders)
		}
		jsonWrite(w, 200, []Response{success(M{"orderNo": "123"})})
	}))
	defer srv.Close()
	a.Config.Upstream.OrderURL = srv.URL
	_, out := request(t, a, "POST", "/basketorder/execute", jsonString(M{"basketId": id, "scrips": []json.RawMessage{json.RawMessage(scripJSON)}}))
	ok(t, out.([]any)[0])
	if calls != 1 {
		t.Fatal(calls)
	}
	_, out = request(t, a, "GET", "/basketorder/get", "")
	b := ok(t, out)["result"].([]any)[0].(map[string]any)
	if b["isExecuted"] != "1" {
		t.Fatal(b)
	}
	_, out = request(t, a, "GET", "/basketorder/reset/"+str(id), "")
	b = ok(t, out)["result"].([]any)[0].(map[string]any)
	if b["isExecuted"] != "0" {
		t.Fatal(b)
	}
}
func TestSpanMargin(t *testing.T) {
	a := testApp(t)
	_, out := request(t, a, "POST", "/basketorder/spanmargin", `[{"exchange":"NSE","token":"2188","qty":"2","price":"43.04","transType":"BUY"}]`)
	m := ok(t, out)["result"].([]any)[0].(map[string]any)
	if m["span"] != "17.22" {
		t.Fatal(m)
	}
}
func TestModuleSwitches(t *testing.T) {
	a := testApp(t)
	a.Config.Modules = config.Modules{}
	h := a.Handler()
	for _, rt := range routes {
		p := strings.ReplaceAll(strings.ReplaceAll(rt.Path, "{id}", "1"), "{basketId}", "1")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(rt.Method, p, nil))
		if w.Code != 404 {
			t.Errorf("disabled %s returned %d", p, w.Code)
		}
	}
}

func TestThematicReviewDetailsAndInvest(t *testing.T) {
	a := testApp(t)
	db := a.DB
	b := model.ThematicMaster{BasketName: ptr("Thematic"), Status: ptr("Open"), ActionType: ptr("SEND_NOW"), TotalInvstAmt: 100, RebalanceAvailable: 1}
	if e := db.Create(&b).Error; e != nil {
		t.Fatal(e)
	}
	s := model.ThematicScrip{BasketId: b.Id, Exchange: ptr("NSE"), Token: ptr("2188"), Qty: ptr("2"), Price: ptr("43.04"), Weightage: ptr("100"), TradingSymbol: ptr("GENCON-EQ"), FormattedInsName: ptr("GENCON-EQ"), OrderType: ptr("Regular"), PriceType: ptr("MKT"), TransType: ptr("BUY"), Version: 1}
	if e := db.Create(&s).Error; e != nil {
		t.Fatal(e)
	}
	db.Create(&model.ReasearchCallUsers{UserId: ptr("ALL"), ThematicBasketId: int(b.Id)})
	_, out := request(t, a, "GET", "/thematic/basket/getall", "")
	if len(ok(t, out)["result"].([]any)) != 1 {
		t.Fatal(out)
	}
	_, out = request(t, a, "GET", "/thematic/basket/get/"+str(b.Id), "")
	ok(t, out)
	_, out = request(t, a, "POST", "/thematic/basket/review", jsonString(M{"basketId": b.Id, "lotSize": 3, "scrips": []M{{"token": "2188", "ltp": "43.045"}}}))
	m := ok(t, out)["result"].([]any)[0].(map[string]any)
	if m["totalInvested"] != "258.27" {
		t.Fatal(m)
	}
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var orders []M
		json.NewDecoder(r.Body).Decode(&orders)
		if orders[0]["remark"] != "TBK:"+str(b.Id) || orders[0]["orderType"] != "Regular" {
			t.Error(orders)
		}
		jsonWrite(w, 200, []Response{success([]M{{"orderNo": "THEMATIC-1", "requestTime": "now"}})})
	}))
	defer srv.Close()
	a.Config.Upstream.OrderURL = srv.URL
	var sr M
	json.Unmarshal([]byte(scripJSON), &sr)
	sr["version"] = 1
	_, out = request(t, a, "POST", "/thematic/basket/v1/invest", jsonString(M{"basketId": b.Id, "lots": 2, "basketAction": "BUY", "source": "WEB", "investmentAmount": 86.08, "scrips": []M{sr}}))
	ok(t, out.([]any)[0])
	if calls != 1 {
		t.Fatal(calls)
	}
	var ds []model.ExecutionDetail
	if e := db.Find(&ds).Error; e != nil || len(ds) != 1 {
		t.Fatal(e, ds)
	}
	if val(ds[0].OrderNo) != "THEMATIC-1" || val(ds[0].LotSize) != 2 {
		t.Fatal(ds)
	}
	_, out = request(t, a, "GET", "/thematic/holdings/get/V1", "")
	h := ok(t, out)["result"].([]any)
	if len(h) != 1 {
		t.Fatal(out)
	}
	holding := h[0].(map[string]any)
	if holding["investedAmount"] != 43.04 || integer(holding["lotSize"]) != 2 {
		t.Fatal(holding)
	}
	_, out = request(t, a, "POST", "/thematic/basket/report", jsonString(M{"basketId": b.Id, "isViewed": 1}))
	ok(t, out)
	_, out = request(t, a, "POST", "/thematic/basket/rebalance/details", jsonString(M{"basketId": b.Id}))
	if out.(map[string]any)["message"] != "No rebalance data available." {
		t.Fatal(out)
	}
}
func TestThematicAMOMarketBoundaries(t *testing.T) {
	a := testApp(t)
	var s model.ScripRequestModel
	json.Unmarshal([]byte(scripJSON), &s)
	for _, tc := range []struct {
		at   string
		want string
	}{{"2026-09-10T09:14:59", "AMO"}, {"2026-09-10T09:15:00", "Regular"}, {"2026-09-10T15:30:00", "Regular"}, {"2026-09-10T15:30:01", "AMO"}, {"2026-09-12T10:00:00", "AMO"}} {
		tm, _ := time.ParseInLocation("2006-01-02T15:04:05", tc.at, a.location)
		a.Now = func() time.Time { return tm }
		order := a.orderRequest([]model.ScripRequestModel{s}, 1)[0]
		if str(order["orderType"]) != tc.want {
			t.Error(tc, order)
		}
	}
}
func TestResearchVisibilityAndReports(t *testing.T) {
	a := testApp(t)
	now := a.Now()
	recent := model.Timestamp(now.AddDate(0, 0, -2))
	old := model.Timestamp(now.AddDate(0, 0, -8))
	for i, created := range []*model.Timestamp{&recent, &old} {
		m := model.ResearchMaster{ResearchcallOrderEntity: model.ResearchcallOrderEntity{Category: ptr("Equity"), SubCategory: ptr("Intraday"), Status: ptr("closed"), CreatedOn: created, ActiveStatus: 1}, AnalystName: ptr("Analyst")}
		if e := a.DB.Create(&m).Error; e != nil {
			t.Fatal(e)
		}
		a.DB.Create(&model.ReasearchCallUsers{ResearchCallId: int(m.Id), UserId: ptr("ALL")})
		s := model.ResearchcallScripEntity{ResearchcallId: int(m.Id), Token: ptr(str(i)), Exchange: ptr("NSE"), Expiry: ptr("2026-09-15"), Retention: ptr("DAY"), StopLossUpperBound: ptr("5")}
		if e := a.DB.Create(&s).Error; e != nil {
			t.Fatal(e)
		}
	}
	_, out := request(t, a, "POST", "/research/getResearchCall", `{}`)
	groups := ok(t, out)["result"].([]any)
	baskets := groups[0].(map[string]any)["Equity"].(map[string]any)["Intraday"].([]any)
	if len(baskets) != 1 {
		t.Fatal(out)
	}
	sc := baskets[0].(map[string]any)["scripdetails"].([]any)[0].(map[string]any)
	if sc["stopLossUpperBound"] != "5" || sc["expiry"] != "2026-09-15" {
		t.Fatal(sc)
	}
	_, out = request(t, a, "POST", "/research/getResearchCall", `{"status":"closed"}`)
	groups = ok(t, out)["result"].([]any)
	if len(groups[0].(map[string]any)["Equity"].(map[string]any)["Intraday"].([]any)) != 2 {
		t.Fatal(out)
	}
	_, out = request(t, a, "POST", "/research/getResearchWithBasket", `{}`)
	ok(t, out)
	a.auth.config.DevUser = "OTHER"
	_, out = request(t, a, "GET", "/research/getUniqStatus", "")
	ok(t, out)
	a.DB.Create(&model.SectorReportsEntity{Name: ptr("report"), Type: ptr("Equity"), ActiveStatus: 1})
	_, out = request(t, a, "GET", "/research/getall/sector/data", "")
	ok(t, out)
	_, out = request(t, a, "GET", "/research/get/sector/data/1", "")
	ok(t, out)
	_, out = request(t, a, "POST", "/research/getall/rc/report", `{"basketType":"Equity"}`)
	ok(t, out)
}
func TestExpiryCleanup(t *testing.T) {
	a := testApp(t)
	yesterday := model.Timestamp(a.midnight().AddDate(0, 0, -1))
	today := model.Timestamp(a.midnight())
	for i, expiry := range []*model.Timestamp{&yesterday, &today} {
		b := model.BasketNameEntity{BasketName: ptr(str(i)), UserId: ptr("USER1"), ExpiryDate: expiry}
		a.DB.Create(&b)
		a.DB.Create(&model.BasketScripEntity{BasketId: b.BasketId, Expiry: expiry})
		a.DB.Create(&model.UserNotification{BasketId: ptr(str(b.BasketId))})
	}
	resp := a.deleteExpiredScrips(context.Background())
	if resp.Message != "1-Record Deleted" {
		t.Fatal(resp)
	}
	resp = a.deleteExpiredBaskets(context.Background())
	if resp.Status != "Ok" {
		t.Fatal(resp)
	}
	var n int64
	a.DB.Model(&model.BasketNameEntity{}).Count(&n)
	if n != 1 {
		t.Fatal(n)
	}
	a.DB.Model(&model.UserNotification{}).Count(&n)
	if n != 1 {
		t.Fatal(n)
	}
}
func TestAdminTempPersistsWithoutPush(t *testing.T) {
	a := testApp(t)
	a.DB.Create(&model.VendorAppEntity{ApiKey: ptr("vendor"), TppAuthorization: 1})
	a.DB.Create(&model.DeviceMappingEntity{UserId: ptr("USER1"), DeviceId: ptr("device")})
	pushes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { pushes++; jsonWrite(w, 200, M{"success": 1}) }))
	defer srv.Close()
	a.Config.Notifications.Provider = "legacy"
	a.Config.Notifications.URL = srv.URL
	req := M{"apiKey": "vendor", "basketName": "Research", "userId": []string{"USER1"}, "expiryDate": "2026-09-30", "scrips": []json.RawMessage{json.RawMessage(scripJSON)}, "pushNotification": 1, "title": "Title", "message": "Message"}
	r := httptest.NewRequest("POST", "/basketorderapi/adminCreate/temp", strings.NewReader(jsonString(req)))
	r.Header.Set("Authorization", "admin-test")
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	var out M
	json.Unmarshal(w.Body.Bytes(), &out)
	ok(t, out)
	deadline := time.Now().Add(3 * time.Second)
	for {
		var n int64
		a.DB.Model(&model.UserNotification{}).Count(&n)
		if n == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("notification was not saved")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if pushes != 0 {
		t.Fatal(pushes)
	}
}

func TestHoldingsV1UsesOriginalRecommendation(t *testing.T) {
	a := testApp(t)
	b := model.ThematicMaster{BasketName: ptr("rebalance"), Status: ptr("Open")}
	if e := a.DB.Create(&b).Error; e != nil {
		t.Fatal(e)
	}
	a.DB.Create(&model.ThematicScrip{BasketId: b.Id, Exchange: ptr("NSE"), Token: ptr("2188"), Qty: ptr("15"), Version: 2})
	now := model.Now()
	d := model.ExecutionDetail{ActiveStatus: 1, ThematicExecDetails: model.ThematicExecDetails{ExecutionId: ptr("1"), UserId: ptr("USER1"), ResearchId: ptr(int(b.Id)), ResearchType: ptr(2), BasketAction: ptr("BUY"), TransType: ptr("BUY"), OrderStatus: ptr("EXECUTED"), ExecutedQty: ptr(20), RecommQty: ptr(10), ExecutedPrice: ptr("5"), Exch: ptr("NSE"), Token: ptr("2188"), Version: ptr(1), LotSize: ptr(2), ExecutedOn: now}}
	if e := a.DB.Create(&d).Error; e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct{ path, action string }{{"/thematic/holdings/get", "Reduce"}, {"/thematic/holdings/get/V1", "Add More"}} {
		_, out := request(t, a, "GET", tc.path, "")
		h := ok(t, out)["result"].([]any)[0].(map[string]any)
		s := h["scripList"].([]any)[0].(map[string]any)
		if s["rebalancedAction"] != tc.action || integer(s["rebalancedQty"]) != 30 || h["investedAmount"] != float64(100) {
			t.Fatal(h)
		}
	}
	a.DB.Create(&model.ThematicExeMasterEntity{UserId: ptr("USER1"), BasketId: b.Id, BasketAction: ptr("SELL"), ActiveStatus: 1})
	_, out := request(t, a, "GET", "/thematic/holdings/get/V1", "")
	if len(ok(t, out)["result"].([]any)) != 0 {
		t.Fatal(out)
	}
}
func TestNestMarginAndOrderBook(t *testing.T) {
	a := testApp(t)
	a.Cache.Put(context.Background(), a.Config.Cache.Maps["customers"], "USER1", M{"stringPkey4": "key", "tomcatcount": "2"})
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("jKey") != "key" || r.URL.Query().Get("jsessionid") != ".2" {
			t.Error("customer query mapping")
		}
		if r.URL.Path == "/margin" {
			var legs []M
			json.Unmarshal([]byte(r.URL.Query().Get("jData")), &legs)
			if len(legs) != 1 || legs[0]["netQty"] != "2" {
				t.Error(legs)
			}
			jsonWrite(w, 200, M{"stat": "Ok", "spanRequirement": "123.45"})
			return
		}
		jsonWrite(w, 200, []M{{"stat": "Ok", "Nstordno": "order:", "Exchange": "NSE", "token": "2188", "Qty": 2, "Avgprc": "42.00", "Prc": "0", "Status": "complete", "OrderedTime": "10/09/2026 10:00:00", "Pcode": "CNC", "Trsym": "GENCON-EQ", "Trantype": "B"}})
	}))
	defer srv.Close()
	a.Config.Upstream.BasketMarginURL = srv.URL + "/margin"
	_, out := request(t, a, "POST", "/basketorder/nest/spanmargin", `[{"exchange":"NSE","qty":"2","symbol":"ABC"},{"exchange":"NSE","qty":"3","symbol":"DEF"}]`)
	if ok(t, out)["result"].([]any)[0].(map[string]any)["span"] != "123.45" {
		t.Fatal(out)
	}
	a.Config.Upstream.OrderBookURL = srv.URL + "/orders"
	d := model.ExecutionDetail{ThematicExecDetails: model.ThematicExecDetails{ExecutionId: ptr("99"), OrderNo: ptr("order"), OrderStatus: ptr("EXECUTED")}}
	if e := a.DB.Create(&d).Error; e != nil {
		t.Fatal(e)
	}
	if e := a.refreshOrderBook(context.Background(), 99, "USER1", []M{{"orderNo": "order"}}); e != nil {
		t.Fatal(e)
	}
	a.DB.First(&d, "id = ?", d.Id)
	if val(d.OrderStatus) != "complete" || val(d.ExecutedQty) != 2 || val(d.ExecutedPrice) != "42.00" {
		t.Fatal(d)
	}
	if calls != 2 {
		t.Fatal(calls)
	}
}

func TestBasketReportsOptimization(t *testing.T) {
	a := testApp(t)

	// Verify empty list when no baskets exist
	_, out := request(t, a, "GET", "/basketorder/get", "")
	if out.(map[string]any)["message"] != "No records found" {
		t.Fatalf("expected No records found, got %#v", out)
	}

	// 1. Create basket 1 (will have 0 scrips)
	id1 := create(t, a, "Basket_Zero_Scrips")

	// 2. Create basket 2 (will have 1 scrip)
	id2 := create(t, a, "Basket_One_Scrip")
	_, _ = request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": id2, "scrips": json.RawMessage(scripJSON)}))

	// 3. Create basket 3 (will have 2 scrips)
	contract2 := model.ContractMasterModel{Exch: ptr("NSE"), Token: ptr("2189"), TradingSymbol: ptr("GENCON2-EQ"), FormattedInsName: ptr("GENCON2-EQ"), LotSize: ptr("1")}
	a.Cache.Put(context.Background(), a.Config.Cache.Maps["contracts"], "NSE_2189", contract2)

	id3 := create(t, a, "Basket_Two_Scrips")
	_, _ = request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": id3, "scrips": json.RawMessage(scripJSON)}))
	var scrip2 M
	json.Unmarshal([]byte(scripJSON), &scrip2)
	scrip2["token"] = "2189"
	scrip2["tradingSymbol"] = "GENCON2-EQ"
	_, _ = request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": id3, "scrips": scrip2}))

	// 4. Create a research call basket (research_call = 1) for the same user, which must be EXCLUDED
	rcBasket := model.BasketNameEntity{
		UserId:       ptr("USER1"),
		BasketName:   ptr("Research_Basket"),
		ResearchCall: 1,
		IsExecuted:   ptr("0"),
	}
	if err := a.DB.Create(&rcBasket).Error; err != nil {
		t.Fatal(err)
	}

	// Fetch baskets via GET /basketorder/get
	_, getOut := request(t, a, "GET", "/basketorder/get", "")
	res := ok(t, getOut)["result"].([]any)

	// Verify only 3 baskets returned (research_call = 1 excluded)
	if len(res) != 3 {
		t.Fatalf("expected 3 baskets, got %d", len(res))
	}

	// Verify ordering: basket_id DESC (id3 > id2 > id1)
	b0 := res[0].(map[string]any)
	b1 := res[1].(map[string]any)
	b2 := res[2].(map[string]any)

	if integer(b0["basketId"]) != id3 || integer(b1["basketId"]) != id2 || integer(b2["basketId"]) != id1 {
		t.Fatalf("ordering mismatch: got IDs [%v, %v, %v], expected [%d, %d, %d]",
			b0["basketId"], b1["basketId"], b2["basketId"], id3, id2, id1)
	}

	// Verify Scrip counts
	if integer(b0["scripCount"]) != 2 {
		t.Fatalf("expected basket 3 scripCount=2, got %v", b0["scripCount"])
	}
	if integer(b1["scripCount"]) != 1 {
		t.Fatalf("expected basket 2 scripCount=1, got %v", b1["scripCount"])
	}
	if integer(b2["scripCount"]) != 0 {
		t.Fatalf("expected basket 1 scripCount=0, got %v", b2["scripCount"])
	}

	// Verify exact response structure fields
	for _, b := range []map[string]any{b0, b1, b2} {
		if _, ok := b["basketId"]; !ok {
			t.Fatal("missing basketId")
		}
		if _, ok := b["basketName"]; !ok {
			t.Fatal("missing basketName")
		}
		if _, ok := b["isExecuted"]; !ok {
			t.Fatal("missing isExecuted")
		}
		if _, ok := b["createdOn"]; !ok {
			t.Fatal("missing createdOn")
		}
		if _, ok := b["scripCount"]; !ok {
			t.Fatal("missing scripCount")
		}
	}
}

func TestBasketAPIPerformance(t *testing.T) {
	a := testApp(t)
	a.Config.Business.MaxScrips = 1000

	// Pre-populate 10 baskets with varying number of scrips to simulate realistic workload
	for i := 1; i <= 10; i++ {
		bid := create(t, a, fmt.Sprintf("PreBasket_%d", i))
		for j := 0; j < (i % 4); j++ {
			_, _ = request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": bid, "scrips": json.RawMessage(scripJSON)}))
		}
	}

	const iterations = 50

	// 1. Measure Get Basket Order
	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, out := request(t, a, "GET", "/basketorder/get", "")
		ok(t, out)
	}
	getDuration := time.Since(start) / iterations

	// 2. Measure Create Basket
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, out := request(t, a, "POST", "/basketorder/create", fmt.Sprintf(`{"basketName":"PerfCreate_%d"}`, i))
		ok(t, out)
	}
	createDuration := time.Since(start) / iterations

	// 3. Measure Rename Basket
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, out := request(t, a, "POST", "/basketorder/rename", fmt.Sprintf(`{"basketId":1,"basketName":"Renamed_%d"}`, i))
		ok(t, out)
	}
	renameDuration := time.Since(start) / iterations

	// 4. Measure Add Script
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, out := request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": 1, "scrips": json.RawMessage(scripJSON)}))
		ok(t, out)
	}
	addScriptDuration := time.Since(start) / iterations

	t.Logf("PERF_RESULTS: Get Basket Order avg: %v", getDuration)
	t.Logf("PERF_RESULTS: Create Basket avg: %v", createDuration)
	t.Logf("PERF_RESULTS: Rename Basket avg: %v", renameDuration)
	t.Logf("PERF_RESULTS: Add Script avg: %v", addScriptDuration)
}
