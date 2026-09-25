package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type smokeTransport func(*http.Request) (*http.Response, error)

func (f smokeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestAllAPIsSmoke uses real loopback HTTP, a disposable SQLite database, a
// fixture cache, signed JWTs, and an outbound allowlist containing one mock only.
// Route coverage is checked against the production registration table.
func TestAllAPIsSmoke(t *testing.T) {
	c, err := config.Load("../../testdata/config.local.yaml")
	if err != nil {
		t.Fatal(err)
	}
	c.Auth.Enabled = false
	c.Auth.DevUser = "USER1"
	c.Auth.AdminToken = "smoke-admin-token"
	c.Logging.Access = false
	c.Modules.Scheduler = false
	c.Modules.Notifications = true
	// A single worker makes the final queue barrier cover every preceding job.
	c.Business.AdminWorkers = 1
	a := newTestApp(t, c)
	a.auth.config.Enabled = true
	a.auth.config.Mode = "hmac"
	a.auth.config.HMACSecret = "smoke-test-only-secret"
	a.auth.config.Issuer = "smoke-issuer"
	a.auth.config.Audience = "basket"
	a.auth.config.UserClaim = "preferred_username"
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"iss": "smoke-issuer", "aud": "basket", "exp": time.Now().Add(time.Hour).Unix(), "preferred_username": "user1", "scope": "basket", "name": "Smoke User"}).SignedString([]byte(a.auth.config.HMACSecret))
	if err != nil {
		t.Fatal(err)
	}
	var orderCalls, spanCalls, nestCalls, bookCalls, pushCalls atomic.Int32
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/orders":
			n := orderCalls.Add(1)
			if r.Header.Get("Authorization") != "Bearer "+signed {
				t.Error("upstream bearer forwarding failed")
			}
			var orders []M
			if err := json.NewDecoder(r.Body).Decode(&orders); err != nil || len(orders) != 1 {
				t.Errorf("unexpected orders: %v %v", orders, err)
			}
			if len(orders) == 1 && n > 1 && !strings.HasPrefix(str(orders[0]["remark"]), "TBK:") {
				t.Error("missing thematic tag")
			}
			jsonWrite(w, 200, []Response{success([]M{{"orderNo": fmt.Sprintf("DRY-%d", n), "requestTime": "2026-09-10 10:00:00"}})})
		case "/span":
			spanCalls.Add(1)
			b, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(b), `"buyqty":"2"`) {
				t.Errorf("invalid span payload: %s", b)
			}
			jsonWrite(w, 200, M{"stat": "Ok", "span_trade": "100", "expo_trade": "25"})
		case "/nest":
			nestCalls.Add(1)
			if r.URL.Query().Get("jKey") != "fixture-key" {
				t.Error("missing NEST customer key")
			}
			jsonWrite(w, 200, M{"stat": "Ok", "spanRequirement": "123.45"})
		case "/book":
			bookCalls.Add(1)
			// Retain EXECUTED so both holdings APIs can discover this fixture.
			jsonWrite(w, 200, []M{{"stat": "Ok", "Nstordno": "DRY-3", "Exchange": "NSE", "token": "2188", "Qty": 1, "Avgprc": "43.04", "Status": "EXECUTED", "OrderedTime": "10/09/2026 10:00:00", "Pcode": "CNC", "Trsym": "GENCON-EQ", "Trantype": "B"}})
		case "/push":
			var payload M
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			if !strings.Contains(jsonString(payload), "smoke-device") {
				t.Error("missing notification device")
			}
			pushCalls.Add(1)
			jsonWrite(w, 200, M{"success": 1, "failure": 0})
		default:
			t.Errorf("unexpected mock path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(mock.Close)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	t.Cleanup(transport.CloseIdleConnections)
	a.HTTP = &http.Client{Timeout: 3 * time.Second, Transport: smokeTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme+"://"+r.URL.Host != mock.URL {
			return nil, fmt.Errorf("dry-run blocked external URL %s", r.URL.Redacted())
		}
		return transport.RoundTrip(r)
	})}
	a.Config.Upstream.OrderURL = mock.URL + "/orders"
	a.Config.Upstream.SpanURL = mock.URL + "/span"
	a.Config.Upstream.BasketMarginURL = mock.URL + "/nest"
	a.Config.Upstream.OrderBookURL = mock.URL + "/book"
	a.Config.Notifications.Provider = "legacy"
	a.Config.Notifications.URL = mock.URL + "/push"
	mustCreate := func(v any) {
		t.Helper()
		if e := a.DB.Create(v).Error; e != nil {
			t.Fatal(e)
		}
	}
	put := func(m, k string, v any) {
		t.Helper()
		if e := a.Cache.Put(context.Background(), a.Config.Cache.Maps[m], k, v); e != nil {
			t.Fatal(e)
		}
	}
	put("customers", "USER1", M{"stringPkey4": "fixture-key", "tomcatcount": "2"})
	put("contracts", "NFO_999", model.ContractMasterModel{Exch: ptr("NFO"), Token: ptr("999"), LotSize: ptr("25"), InsType: ptr("FUTIDX"), Symbol: ptr("NIFTY")})
	mustCreate(&model.VendorAppEntity{ApiKey: ptr("smoke-vendor"), TppAuthorization: 1})
	mustCreate(&model.DeviceMappingEntity{UserId: ptr("USER1"), DeviceId: ptr("smoke-device")})
	thematic := model.ThematicMaster{BasketName: ptr("Smoke thematic"), Status: ptr("Open"), ActionType: ptr("SEND_NOW"), TotalInvstAmt: 100, RebalanceAvailable: 1}
	mustCreate(&thematic)
	ts := model.ThematicScrip{BasketId: thematic.Id, Exchange: ptr("NSE"), Token: ptr("2188"), Qty: ptr("2"), Price: ptr("43.04"), Weightage: ptr("100"), TradingSymbol: ptr("GENCON-EQ"), FormattedInsName: ptr("GENCON-EQ"), OrderType: ptr("Regular"), PriceType: ptr("MKT"), TransType: ptr("BUY"), Version: 1}
	mustCreate(&ts)
	rb := model.RebalanceScrip(ts)
	rb.Id = 0
	mustCreate(&rb)
	mustCreate(&model.ReasearchCallUsers{UserId: ptr("ALL"), ThematicBasketId: int(thematic.Id)})
	now := model.Timestamp(a.Now())
	research := model.ResearchMaster{ResearchcallOrderEntity: model.ResearchcallOrderEntity{Category: ptr("Equity"), SubCategory: ptr("Intraday"), Status: ptr("open"), CreatedOn: &now, ActiveStatus: 1}, AnalystName: ptr("Smoke Analyst")}
	mustCreate(&research)
	mustCreate(&model.ReasearchCallUsers{UserId: ptr("ALL"), ResearchCallId: int(research.Id)})
	mustCreate(&model.ResearchcallScripEntity{ResearchcallId: int(research.Id), Token: ptr("2188"), Exchange: ptr("NSE"), Qty: ptr("2"), Expiry: ptr("2026-09-30")})
	sector := model.SectorReportsEntity{Name: ptr("Smoke report"), Type: ptr("Equity"), ActiveStatus: 1}
	mustCreate(&sector)
	yesterday := model.Timestamp(a.midnight().AddDate(0, 0, -1))
	expired := model.BasketNameEntity{BasketName: ptr("Expired fixture"), UserId: ptr("USER1"), ExpiryDate: &yesterday}
	mustCreate(&expired)
	mustCreate(&model.BasketScripEntity{BasketId: expired.BasketId, Expiry: &yesterday})
	mustCreate(&model.UserNotification{BasketId: ptr(str(expired.BasketId))})
	server := httptest.NewServer(a.Handler())
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = 5 * time.Second
	call := func(t *testing.T, rt route, path, body, authorization string) (int, string) {
		t.Helper()
		r, e := http.NewRequest(rt.Method, server.URL+path, strings.NewReader(body))
		if e != nil {
			t.Fatal(e)
		}
		r.Header.Set("Content-Type", "application/json")
		if authorization != "" {
			r.Header.Set("Authorization", authorization)
		}
		res, e := client.Do(r)
		if e != nil {
			t.Fatal(e)
		}
		defer res.Body.Close()
		b, e := io.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}
		return res.StatusCode, string(b)
	}
	byAction := map[string]route{}
	for _, rt := range routes {
		byAction[rt.Action] = rt
	}
	covered := map[string]bool{}
	var basketID, scripID int
	var scrip M
	if e := json.Unmarshal([]byte(scripJSON), &scrip); e != nil {
		t.Fatal(e)
	}
	firstResult := func(t *testing.T, out any) M {
		t.Helper()
		m := ok(t, out)
		rs, yes := m["result"].([]any)
		if !yes || len(rs) == 0 {
			t.Fatalf("missing populated result: %v", m)
		}
		return rs[0].(map[string]any)
	}
	count := func(t *testing.T, table, where string, args ...any) int64 {
		t.Helper()
		var n int64
		if e := a.DB.Table(table).Where(where, args...).Count(&n).Error; e != nil {
			t.Fatal(e)
		}
		return n
	}
	await := func(t *testing.T, predicate func() bool) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for !predicate() {
			if time.Now().After(deadline) {
				t.Fatal("background side effect did not complete")
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	run := func(action string, body any, check func(*testing.T, any)) {
		t.Helper()
		rt, exists := byAction[action]
		if !exists {
			t.Fatalf("unknown smoke action %s", action)
		}
		covered[action] = true
		if !t.Run(action, func(t *testing.T) {
			path := strings.ReplaceAll(rt.Path, "{basketId}", str(basketID))
			id := thematic.Id
			if action == "sector" {
				id = val(sector.Id)
			}
			path = strings.ReplaceAll(path, "{id}", str(id))
			payload := ""
			if body != nil {
				payload = jsonString(body)
			}
			auth := "Bearer " + signed
			if action == "adminCreate" || action == "adminTemp" {
				auth = a.Config.Auth.AdminToken
			}
			status, raw := call(t, rt, path, payload, auth)
			expected := 200
			if action == "category" {
				expected = 204
			}
			if status != expected {
				t.Fatalf("HTTP %d, want %d: %s", status, expected, raw)
			}
			var out any
			switch action {
			case "category":
				if raw != "" {
					t.Fatal("expected empty category response")
				}
			case "token":
				if !strings.Contains(raw, "username: USER1") || !strings.Contains(raw, "scopes: basket") {
					t.Fatal(raw)
				}
			case "logout":
				if raw != "You are logged out" {
					t.Fatal(raw)
				}
			default:
				if e := json.Unmarshal([]byte(raw), &out); e != nil {
					t.Fatal(e, raw)
				}
				if rt.List {
					list, yes := out.([]any)
					if !yes || len(list) == 0 {
						t.Fatal(out)
					}
					for _, v := range list {
						ok(t, v)
					}
				} else {
					ok(t, out)
				}
			}
			if check != nil {
				check(t, out)
			}
			t.Logf("SMOKE_RESULT %s", jsonString(M{"method": rt.Method, "path": rt.Path, "http_status": status, "response": raw}))
		}) {
			t.FailNow()
		}
	}
	run("create", M{"basketName": "Smoke personal"}, func(t *testing.T, o any) { basketID = integer(firstResult(t, o)["basketId"]) })
	run("get", nil, func(t *testing.T, o any) {
		if !strings.Contains(jsonString(o), "Smoke personal") {
			t.Fatal(o)
		}
	})
	run("rename", M{"basketId": basketID, "basketName": "Smoke renamed"}, func(t *testing.T, o any) {
		if count(t, "tbl_basket_order", "basket_id = ? AND basket_name = ?", basketID, "Smoke renamed") != 1 {
			t.Fatal("rename not persisted")
		}
	})
	run("addScrip", M{"basketId": basketID, "scrips": scrip}, func(t *testing.T, o any) { scripID = integer(firstResult(t, o)["id"]) })
	run("scrips", nil, func(t *testing.T, o any) {
		if firstResult(t, o)["lotSize"] != "1" {
			t.Fatal(o)
		}
	})
	scrip["id"] = scripID
	scrip["qty"] = "2"
	run("updateScrip", M{"basketId": basketID, "scrips": scrip}, func(t *testing.T, o any) {
		var s model.BasketScripEntity
		if e := a.DB.First(&s, "id = ?", scripID).Error; e != nil || val(s.Qty) != "2" {
			t.Fatal("scrip update", e, s)
		}
	})
	scrip["qty"] = "3"
	run("updateScripList", M{"basketId": basketID, "scrips": []M{scrip}}, func(t *testing.T, o any) {
		var s model.BasketScripEntity
		if e := a.DB.First(&s, "id = ?", scripID).Error; e != nil || val(s.Qty) != "3" {
			t.Fatal("scrip list update", e, s)
		}
	})
	run("execute", M{"basketId": basketID, "scrips": []M{scrip}}, func(t *testing.T, o any) {
		if count(t, "tbl_basket_order", "basket_id = ? AND is_executed = ?", basketID, "1") != 1 {
			t.Fatal("execution flag not persisted")
		}
	})
	run("reset", nil, func(t *testing.T, o any) {
		if count(t, "tbl_basket_order", "basket_id = ? AND is_executed = ?", basketID, "0") != 1 {
			t.Fatal("execution flag not reset")
		}
	})
	run("span", []M{{"exchange": "NSE", "token": "2188", "qty": "2", "price": "43.04", "transType": "BUY"}, {"exchange": "NFO", "token": "999", "qty": "2", "price": "10", "transType": "BUY"}}, func(t *testing.T, o any) {
		if firstResult(t, o)["span"] != "142.22" {
			t.Fatal(o)
		}
	})
	run("nestSpan", []M{{"exchange": "NSE", "qty": "2", "symbol": "GENCON-EQ"}}, func(t *testing.T, o any) {
		if firstResult(t, o)["span"] != "123.45" {
			t.Fatal(o)
		}
	})
	run("deleteScrip", M{"basketId": basketID, "scripsId": []int{scripID}}, func(t *testing.T, o any) {
		if len(ok(t, o)["result"].([]any)) != 0 {
			t.Fatal(o)
		}
	})
	run("delete", nil, func(t *testing.T, o any) {
		if count(t, "tbl_basket_order", "basket_id = ?", basketID) != 0 {
			t.Fatal("basket not deleted")
		}
	})
	adminBody := M{"apiKey": "smoke-vendor", "basketName": "Smoke admin", "userId": []string{"USER1"}, "expiryDate": "2026-09-30", "scrips": []json.RawMessage{json.RawMessage(scripJSON)}, "pushNotification": 1, "title": "Smoke", "message": "Dry run"}
	run("adminCreate", adminBody, func(t *testing.T, o any) {
		await(t, func() bool { return pushCalls.Load() == 1 })
		if count(t, "tbl_basket_order", "research_call = ?", 1) != 1 {
			t.Fatal("admin basket missing")
		}
	})
	adminBody["basketName"] = "Smoke temp"
	run("adminTemp", adminBody, func(t *testing.T, o any) {
		await(t, func() bool { return count(t, "tbl_basket_order", "research_call = ?", 1) == 2 })
	})
	run("research", M{}, func(t *testing.T, o any) {
		if !strings.Contains(jsonString(o), "Smoke Analyst") {
			t.Fatal(o)
		}
	})
	run("researchBasket", M{}, func(t *testing.T, o any) {
		if !strings.Contains(jsonString(o), "2188") {
			t.Fatal(o)
		}
	})
	run("status", nil, func(t *testing.T, o any) {
		if !strings.Contains(jsonString(o), "open") {
			t.Fatal(o)
		}
	})
	run("sectors", nil, func(t *testing.T, o any) {
		if firstResult(t, o)["title"] != "Smoke report" {
			t.Fatal(o)
		}
	})
	run("sector", nil, func(t *testing.T, o any) {
		if firstResult(t, o)["name"] != "Smoke report" {
			t.Fatal(o)
		}
	})
	run("reports", M{"basketType": "Equity"}, func(t *testing.T, o any) {
		if firstResult(t, o)["title"] != "Smoke report" {
			t.Fatal(o)
		}
	})
	run("thematicAll", nil, func(t *testing.T, o any) {
		if firstResult(t, o)["basketName"] != "Smoke thematic" {
			t.Fatal(o)
		}
	})
	run("thematicDetails", nil, func(t *testing.T, o any) {
		if firstResult(t, o)["minInvstAmt"] != float64(100) {
			t.Fatal(o)
		}
	})
	run("review", M{"basketId": thematic.Id, "lotSize": 3, "scrips": []M{{"token": "2188", "ltp": "43.045"}}}, func(t *testing.T, o any) {
		if firstResult(t, o)["totalInvested"] != "258.27" {
			t.Fatal(o)
		}
	})
	scrip["qty"] = "1"
	scrip["version"] = 1
	investment := M{"basketId": thematic.Id, "lots": 2, "basketAction": "BUY", "source": "WEB", "investmentAmount": 43.04, "scrips": []M{scrip}}
	run("invest", investment, func(t *testing.T, o any) {
		if count(t, "tbl_user_thematic_exec", "basket_id = ? AND is_executed = ?", thematic.Id, 1) != 1 {
			t.Fatal("legacy journal missing")
		}
	})
	run("investV1", investment, func(t *testing.T, o any) {
		await(t, func() bool { return bookCalls.Load() == 1 })
		if count(t, "tbl_user_thematic_exec_details", "order_no = ?", "DRY-3") != 1 {
			t.Fatal("execution detail missing")
		}
	})
	run("rebalance", M{"basketId": thematic.Id}, func(t *testing.T, o any) {
		if firstResult(t, o)["token"] != "2188" {
			t.Fatal(o)
		}
	})
	run("view", M{"basketId": thematic.Id, "isViewed": 1}, func(t *testing.T, o any) {
		if count(t, "tbl_user_thematic_exec", "basket_id = ? AND is_viewed = ?", thematic.Id, 1) != 1 {
			t.Fatal("view not saved")
		}
	})
	run("category", nil, nil)
	run("legacyHoldings", nil, func(t *testing.T, o any) {
		if firstResult(t, o)["basketName"] != "Smoke thematic" {
			t.Fatal(o)
		}
	})
	for _, action := range []string{"holdings", "holdingsV1"} {
		run(action, nil, func(t *testing.T, o any) {
			if firstResult(t, o)["investedAmount"] != 43.04 {
				t.Fatal(o)
			}
		})
	}
	run("expiry", nil, func(t *testing.T, o any) {
		if ok(t, o)["message"] != "1-Record Deleted" {
			t.Fatal(o)
		}
	})
	run("deleteExpired", nil, func(t *testing.T, o any) {
		if count(t, "tbl_basket_order", "basket_id = ?", expired.BasketId) != 0 {
			t.Fatal("expired basket not deleted")
		}
	})
	run("token", nil, nil)
	run("logout", nil, nil)
	for _, rt := range routes {
		if !covered[rt.Action] {
			t.Errorf("missing route smoke: %s %s", rt.Method, rt.Path)
		}
	}
	if len(covered) != 36 {
		t.Errorf("expected 36 routes, got %d", len(covered))
	}
	// Drain queued work before checking counters; keep all mocks available.
	done := make(chan struct{})
	a.jobs <- func() { close(done) }
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("queue drain timeout")
	}
	if orderCalls.Load() != 3 || spanCalls.Load() != 1 || nestCalls.Load() != 1 || bookCalls.Load() != 1 || pushCalls.Load() != 1 {
		t.Fatalf("unexpected mock calls orders=%d span=%d nest=%d book=%d push=%d", orderCalls.Load(), spanCalls.Load(), nestCalls.Load(), bookCalls.Load(), pushCalls.Load())
	}
	// Every protected route must reject missing credentials before business work.
	for _, rt := range routes {
		public := false
		for _, p := range a.Config.Auth.PublicPaths {
			public = public || p == rt.Path
		}
		if public {
			continue
		}
		t.Run("unauthorized_"+rt.Action, func(t *testing.T) {
			path := strings.ReplaceAll(strings.ReplaceAll(rt.Path, "{id}", "1"), "{basketId}", "1")
			status, body := call(t, rt, path, "{}", "")
			if status != 401 {
				t.Fatalf("HTTP %d: %s", status, body)
			}
		})
	}
	a.Config.Modules = config.Modules{}
	disabled := httptest.NewServer(a.Handler())
	t.Cleanup(disabled.Close)
	oldServer := server
	server = disabled
	for _, rt := range routes {
		t.Run("disabled_"+rt.Action, func(t *testing.T) {
			path := strings.ReplaceAll(strings.ReplaceAll(rt.Path, "{id}", "1"), "{basketId}", "1")
			status, body := call(t, rt, path, "{}", "Bearer "+signed)
			if status != 404 {
				t.Fatalf("HTTP %d: %s", status, body)
			}
		})
	}
	server = oldServer
}
