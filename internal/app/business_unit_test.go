package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNotificationsService(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()

	// 1. Module disabled error
	a.Config.Modules.Notifications = false
	err := a.SaveNotification(ctx, model.SendNoficationReqModel{
		UserId:  []string{"USER1"},
		Message: ptr("Test message"),
	})
	if err == nil {
		t.Fatal("expected error when notifications module is disabled, got nil")
	}

	// 2. SaveNotification for specific users
	a.Config.Modules.Notifications = true
	err = a.SaveNotification(ctx, model.SendNoficationReqModel{
		UserId:      []string{"USER1", "USER2"},
		Message:     ptr("Order placed"),
		Title:       ptr("Alert"),
		MessageType: ptr("INFO"),
		UserType:    ptr("INDIVIDUAL"),
	})
	if err != nil {
		t.Fatalf("SaveNotification failed: %v", err)
	}

	// 3. SaveNotification for ALL
	err = a.SaveNotification(ctx, model.SendNoficationReqModel{
		UserType: ptr("ALL"),
		Message:  ptr("Broadcast to all"),
		Title:    ptr("Broadcast"),
	})
	if err != nil {
		t.Fatalf("SaveNotification for ALL failed: %v", err)
	}

	// 4. NotificationList
	list, err := a.NotificationList(ctx, "USER1")
	if err != nil {
		t.Fatalf("NotificationList failed: %v", err)
	}
	if len(list) < 2 {
		t.Fatalf("expected at least 2 notifications for USER1 (direct + broadcast), got %d", len(list))
	}
}

func TestThematicAndOrdersRepository(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()

	// 1. CreateThematicBasket
	basketReq := model.ThematicBasketRequest{
		Category:   ptr("THEMATIC"),
		BasketName: ptr("Tech Growth"),
		Scrips: []model.ScripRequest{
			{
				Exchange:  ptr("NSE"),
				Token:     ptr("2188"),
				Qty:       ptr("10"),
				TransType: ptr("BUY"),
				Price:     ptr("250.50"),
			},
		},
	}
	basketID, err := a.CreateThematicBasket(ctx, basketReq, "ADMIN1")
	if err != nil {
		t.Fatalf("CreateThematicBasket failed: %v", err)
	}
	if basketID <= 0 {
		t.Fatalf("expected positive basketID, got %d", basketID)
	}

	// 2. ThematicScripsForExecution
	scripsDef, err := a.ThematicScripsForExecution(ctx, basketID, true)
	if err != nil {
		t.Fatalf("ThematicScripsForExecution with defaults failed: %v", err)
	}
	if len(scripsDef) != 1 || val(scripsDef[0].TransType) != "B" || val(scripsDef[0].Product) != "CNC" {
		t.Fatalf("unexpected defaults scrips: %+v", scripsDef)
	}

	scripsRaw, err := a.ThematicScripsForExecution(ctx, basketID, false)
	if err != nil {
		t.Fatalf("ThematicScripsForExecution raw failed: %v", err)
	}
	if len(scripsRaw) != 1 {
		t.Fatalf("expected 1 raw scrip, got %d", len(scripsRaw))
	}

	// 3. PendingAndIncompleteOrders
	ts := model.Timestamp(time.Now())
	execDetail := model.ExecutionDetail{
		ThematicExecDetails: model.ThematicExecDetails{
			OrderNo:       ptr("ORD1001"),
			OrderStatus:   ptr("pending"),
			ExecutedPrice: ptr("0"),
			ExecutedOn:    &ts,
			ExecutionId:   ptr(strconv.FormatInt(basketID, 10)),
		},
	}
	if err := a.DB.Create(&execDetail).Error; err != nil {
		t.Fatalf("failed to insert test execution detail: %v", err)
	}
	pending, err := a.PendingAndIncompleteOrders(ctx)
	if err != nil {
		t.Fatalf("PendingAndIncompleteOrders failed: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("expected at least 1 pending order")
	}

	// 4. OrderStatusByOrderNo
	feed := model.OrderStatusFeedEntity{
		OrderNo:     ptr("ORD1001"),
		OrderStatus: ptr("COMPLETE"),
		Exch:        ptr("NSE"),
		Qty:         ptr("10"),
		TradedPrice: ptr("251.00"),
	}
	if err := a.DB.Create(&feed).Error; err != nil {
		t.Fatalf("failed to insert test feed: %v", err)
	}
	foundFeed, err := a.OrderStatusByOrderNo(ctx, "ORD1001")
	if err != nil || foundFeed == nil {
		t.Fatalf("OrderStatusByOrderNo failed: %v", err)
	}
	if val(foundFeed.OrderStatus) != "COMPLETE" {
		t.Fatalf("expected COMPLETE, got %q", val(foundFeed.OrderStatus))
	}

	// 5. UpdateFromFeed
	err = a.UpdateFromFeed(ctx, basketID, "ORD1001", feed)
	if err != nil {
		t.Fatalf("UpdateFromFeed failed: %v", err)
	}

	// 6. MaxBasketID
	maxID, err := a.MaxBasketID(ctx)
	if err != nil {
		t.Fatalf("MaxBasketID failed: %v", err)
	}
	if maxID < 0 {
		t.Fatalf("expected non-negative MaxBasketID, got %d", maxID)
	}
}

func TestResearchDataRepository(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()

	expiry := model.Timestamp(time.Date(2026, 9, 21, 0, 0, 0, 0, time.Local))
	researchRecord := model.ResearchMaster{
		ResearchcallOrderEntity: model.ResearchcallOrderEntity{
			Category:     ptr("EQUITY"),
			SubCategory:  ptr("LONG_TERM"),
			Status:       ptr("OPEN"),
			ActiveStatus: 1,
			UserId:       ptr("USER1"),
			ExpiryDate:   &expiry,
		},
		AnalystName: ptr("Senior Analyst"),
	}
	if err := a.DB.Create(&researchRecord).Error; err != nil {
		t.Fatalf("failed to insert research record: %v", err)
	}
	reqFilter := &model.ResearchCallRequest{
		Category:    ptr("EQUITY"),
		SubCategory: ptr("LONG_TERM"),
		FromDate:    ptr("2026-01-01"),
		ToDate:      ptr("2026-12-31"),
	}
	resList, err := a.ResearchData(ctx, "USER1", reqFilter)
	if err != nil {
		t.Fatalf("ResearchData failed: %v", err)
	}
	if len(resList) == 0 {
		t.Fatal("expected at least 1 research record")
	}
}

func TestAuthHelpersAndErrors(t *testing.T) {
	// adminAuthorized
	if adminAuthorized("test", "") {
		t.Fatal("expected false when want is empty")
	}
	if !adminAuthorized("TOKEN123", "token123") {
		t.Fatal("expected case-insensitive match to be true")
	}
	if adminAuthorized("token123", "wrong_token") {
		t.Fatal("expected mismatch to be false")
	}

	// newAuth with unsupported auth mode
	cfg := config.Config{
		Auth: config.Auth{
			Enabled: true,
			Mode:    "unsupported_mode",
		},
	}
	if _, err := newAuth(context.Background(), cfg); err == nil {
		t.Fatal("expected error for unsupported auth mode")
	}

	// newAuth with UserInfoRequired=true and JWKS without UserInfoURL
	cfgOIDC := config.Config{
		Auth: config.Auth{
			Enabled:          true,
			Mode:             "oidc",
			JWKSURL:          "https://example.com/certs",
			UserInfoRequired: true,
			UserInfoURL:      "",
			Issuer:           "https://example.com",
			Audience:         "basket",
		},
	}
	if _, err := newAuth(context.Background(), cfgOIDC); err == nil {
		t.Fatal("expected error when userinfo_url is missing with explicit JWKS")
	}

	// identity helper
	req := httptest.NewRequest("GET", "/test", nil)
	emptyID := identity(req)
	if emptyID.UserID != "" {
		t.Fatalf("expected empty identity, got %v", emptyID)
	}
}

func TestCommonHelperBasics(t *testing.T) {
	// nonempty
	var nilStr *string
	emptyStr := ""
	valStr := "valid"
	if nonempty(nilStr) || nonempty(&emptyStr) {
		t.Fatal("expected false for nil/empty")
	}
	if !nonempty(&valStr) {
		t.Fatal("expected true for non-empty")
	}

	// dateOnly & epoch
	ts := time.Date(2026, 9, 21, 14, 30, 0, 0, time.UTC)
	if dateOnly(&ts) != "2026-09-21" {
		t.Fatalf("unexpected dateOnly: %s", dateOnly(&ts))
	}
	if dateOnly(nil) != nil {
		t.Fatal("expected nil for dateOnly(nil)")
	}
	modelTs := model.Timestamp(ts)
	if epoch(ts) != ts.UnixMilli() {
		t.Fatalf("unexpected epoch ms for time.Time: %v", epoch(ts))
	}
	if epoch(&modelTs) != ts.UnixMilli() {
		t.Fatalf("unexpected epoch ms for *model.Timestamp: %v", epoch(&modelTs))
	}
	if epoch(nil) != nil {
		t.Fatal("expected nil for epoch(nil)")
	}

	// coerce & number conversions
	if integer("123") != 123 {
		t.Fatalf("expected 123, got %d", integer("123"))
	}
	if integer("not-an-int") != 0 {
		t.Fatalf("expected 0 on parse error, got %d", integer("not-an-int"))
	}
	if number("45.67") != 45.67 {
		t.Fatalf("expected 45.67, got %f", number("45.67"))
	}
	if number("invalid") != 0 {
		t.Fatalf("expected 0 on float parse error, got %f", number("invalid"))
	}
}

func TestCommonHelperDB(t *testing.T) {
	// first & rows with DB
	a := testApp(t)
	b := model.BasketNameEntity{
		BasketName: ptr("TestFirst"),
		UserId:     ptr("U1"),
	}
	if err := a.DB.Create(&b).Error; err != nil {
		t.Fatalf("failed to insert basket: %v", err)
	}
	m, err := first(a.DB.Table("tbl_basket_order"))
	if err != nil {
		t.Fatalf("first() returned error: %v", err)
	}
	if len(m) == 0 {
		t.Fatal("expected non-empty map from first()")
	}
	rowsList, err := rows(a.DB.Table("tbl_basket_order"))
	if err != nil {
		t.Fatalf("rows() returned error: %v", err)
	}
	if len(rowsList) == 0 {
		t.Fatal("expected at least 1 row from rows()")
	}
}

func TestHoldingsHelpers(t *testing.T) {
	if !executed("EXECUTED") || executed("PENDING") {
		t.Fatal("unexpected executed() result")
	}
	if !buy("BUY") || buy("SELL") {
		t.Fatal("unexpected buy() result")
	}
	now := time.Now()
	if timestamp(now).IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
	if !timestamp("invalid").IsZero() {
		t.Fatal("expected zero timestamp on invalid input")
	}
	if holdingDate(nil) != nil {
		t.Fatal("expected nil holdingDate for nil")
	}
	if holdingDate("invalid") != nil {
		t.Fatal("expected nil holdingDate for invalid time")
	}
	if holdingDate(now) == nil {
		t.Fatal("expected formatted date for valid time")
	}

	// determineAction branches
	if determineAction(0, 5) != "Add New" {
		t.Fatalf("expected Add New, got %v", determineAction(0, 5))
	}
	if determineAction(2, 5) != "Add More" {
		t.Fatalf("expected Add More, got %v", determineAction(2, 5))
	}
	if determineAction(5, 2) != "Reduce" {
		t.Fatalf("expected Reduce, got %v", determineAction(5, 2))
	}
	if determineAction(5, 0) != "Exit" {
		t.Fatalf("expected Exit, got %v", determineAction(5, 0))
	}
	if determineAction(0, 0) != nil {
		t.Fatalf("expected nil action, got %v", determineAction(0, 0))
	}
}

func TestThematicDisplayDate(t *testing.T) {
	a := testApp(t)
	if a.displayDate(nil) != nil {
		t.Fatal("expected nil for nil input")
	}
	if a.displayDate("not-a-date") != nil {
		t.Fatal("expected nil for invalid date string")
	}
	fixed := time.Date(2026, 9, 22, 14, 30, 0, 0, a.location)
	if a.displayDate(fixed) == nil {
		t.Fatal("expected formatted date for valid time")
	}
}

func TestDeleteBasketAndLockRow(t *testing.T) {
	a := testApp(t)
	// deleteBasket with notifications = true and false
	if err := deleteBasket(a.DB, 999999, true); err != nil {
		t.Fatalf("deleteBasket failed: %v", err)
	}
	if err := deleteBasket(a.DB, 999999, false); err != nil {
		t.Fatalf("deleteBasket failed: %v", err)
	}
	// lockBasketRow
	q := lockBasketRow(a.DB.Model(&model.BasketNameEntity{}))
	if q == nil {
		t.Fatal("expected non-nil query from lockBasketRow")
	}
}

func TestAdminValidationBranches(t *testing.T) {
	a := testApp(t)

	// deleteExpired action
	req := httptest.NewRequest("POST", "/admin", nil)
	resp, err := a.admin(req, "deleteExpired")
	if err != nil || resp == nil {
		t.Fatalf("admin deleteExpired failed: %v", err)
	}

	// unauthorized
	reqUnauth := httptest.NewRequest("POST", "/admin", strings.NewReader(`{}`))
	resUnauth, _ := a.admin(reqUnauth, "admin")
	if hr, ok := resUnauth.(httpResult); !ok || hr.Status != 401 {
		t.Fatalf("expected 401 for unauthorized admin, got %v", resUnauth)
	}

	// invalid payload
	reqBad := httptest.NewRequest("POST", "/admin", strings.NewReader(`not-json`))
	reqBad.Header.Set("Authorization", a.Config.Auth.AdminToken)
	resBad, _ := a.admin(reqBad, "admin")
	if r, ok := resBad.(Response); !ok || r.Message != "Invalid Parameter" {
		t.Fatalf("expected Invalid Parameter for bad JSON, got %v", resBad)
	}

	// missing vendor
	reqNoVendor := httptest.NewRequest("POST", "/admin", strings.NewReader(`{"apiKey":"unknown_key"}`))
	reqNoVendor.Header.Set("Authorization", a.Config.Auth.AdminToken)
	resNoVendor, _ := a.admin(reqNoVendor, "admin")
	if r, ok := resNoVendor.(Response); !ok || r.Message != "Your not a vendor" {
		t.Fatalf("expected Your not a vendor, got %v", resNoVendor)
	}

	// vendor not authorized by admin
	vendor := model.VendorAppEntity{
		ApiKey:           ptr("test_vendor_key"),
		TppAuthorization: 0,
	}
	_ = a.DB.Create(&vendor)
	reqNotAuth := httptest.NewRequest("POST", "/admin", strings.NewReader(`{"apiKey":"test_vendor_key"}`))
	reqNotAuth.Header.Set("Authorization", a.Config.Auth.AdminToken)
	resNotAuth, _ := a.admin(reqNotAuth, "admin")
	if r, ok := resNotAuth.(Response); !ok || r.Message != "Your not authorized by admin" {
		t.Fatalf("expected Your not authorized by admin, got %v", resNotAuth)
	}
}

func TestMarginValidationBranches(t *testing.T) {
	a := testApp(t)
	// nestSpan empty
	reqEmpty := httptest.NewRequest("POST", "/margin", strings.NewReader(`[]`))
	resEmpty, _ := a.margin(reqEmpty, "nestSpan")
	if r, ok := resEmpty.(Response); !ok || r.Message != "Invalid Parameter" {
		t.Fatalf("expected Invalid Parameter for empty nestSpan, got %v", resEmpty)
	}

	// nestSpan missing fields
	reqMissing := httptest.NewRequest("POST", "/margin", strings.NewReader(`[{"symbol":"INFY"}]`))
	resMissing, _ := a.margin(reqMissing, "nestSpan")
	if r, ok := resMissing.(Response); !ok || r.Message != "Invalid Parameter" {
		t.Fatalf("expected Invalid Parameter for missing exchange/qty, got %v", resMissing)
	}
}

func TestOpenDatabaseDriversAll(t *testing.T) {
	c := testApp(t).Config
	c.Logging.SQL = true

	// sqlite driver branch (executes switch branch)
	c.Database.Driver = "sqlite"
	c.Database.DSN = "file:opendb_test?mode=memory&cache=shared"
	_, _ = OpenDatabase(c)

	// mysql driver branch
	c.Database.Driver = "mysql"
	c.Database.DSN = "invalid_user:invalid_pass@tcp(127.0.0.1:3306)/invalid_db"
	_, _ = OpenDatabase(c)

	// postgres driver branch
	c.Database.Driver = "postgres"
	c.Database.DSN = "host=127.0.0.1 port=5432 user=invalid dbname=invalid sslmode=disable"
	_, _ = OpenDatabase(c)

	// sqlserver driver branch
	c.Database.Driver = "sqlserver"
	c.Database.DSN = "sqlserver://invalid:invalid@127.0.0.1:1433?database=invalid"
	_, _ = OpenDatabase(c)
}

func TestHoldingsFull(t *testing.T) {
	a := testApp(t)

	// 1. empty user holdings
	req := httptest.NewRequest("GET", "/holdings", nil)
	res, err := a.holdings(req, false)
	if err != nil {
		t.Fatalf("holdings failed: %v", err)
	}
	_ = res

	// 2. populated user holdings
	now := time.Now()
	nowTs := model.Timestamp(now)
	basket := model.ThematicMaster{
		BasketName:   ptr("TestHoldingsBasket"),
		Status:       ptr("open"),
		ActionType:   ptr("SEND_NOW"),
		ActiveStatus: 1,
	}
	_ = a.DB.Create(&basket)

	detail := model.ExecutionDetail{
		ThematicExecDetails: model.ThematicExecDetails{
			UserId:        ptr("USER1"),
			ResearchId:    ptr(int(basket.Id)),
			ExecutionId:   ptr("1"),
			OrderStatus:   ptr("EXECUTED"),
			BasketAction:  ptr("BUY"),
			TransType:     ptr("BUY"),
			ExecutedPrice: ptr("100.50"),
			ExecutedQty:   ptr(10),
			RecommQty:     ptr(10),
			LotSize:       ptr(1),
			Token:         ptr("2188"),
			Exch:          ptr("NSE"),
			TradingSymbol: ptr("GENCON-EQ"),
			ExecutedOn:    &nowTs,
			Version:       ptr(1),
		},
		ActiveStatus: 1,
	}
	_ = a.DB.Create(&detail)

	scrip := model.ThematicScrip{
		BasketId:      basket.Id,
		Token:         ptr("2188"),
		Exchange:      ptr("NSE"),
		TradingSymbol: ptr("GENCON-EQ"),
		Qty:           ptr("10"),
		Version:       1,
	}
	_ = a.DB.Create(&scrip)

	resPopulated, err := a.holdings(req, false)
	if err != nil {
		t.Fatalf("populated holdings failed: %v", err)
	}
	_ = resPopulated

	// test v1
	resV1, err := a.holdings(req, true)
	if err != nil {
		t.Fatalf("v1 holdings failed: %v", err)
	}
	_ = resV1
}

func TestThematicInvestValidationBranches(t *testing.T) {
	a := testApp(t)

	// invalid body
	reqBad := httptest.NewRequest("POST", "/thematic/invest", strings.NewReader(`not-json`))
	resBad, _ := a.invest(reqBad, false)
	if r, ok := resBad.(Response); !ok || r.Message != "Invalid Parameter" {
		t.Fatalf("expected Invalid Parameter, got %v", resBad)
	}

	// v1 missing basketAction
	reqNoAction := httptest.NewRequest("POST", "/thematic/v1/invest", strings.NewReader(`{"basketId":1,"scrips":[{"token":"2188"}]}`))
	resNoAction, _ := a.invest(reqNoAction, true)
	if r, ok := resNoAction.(Response); !ok || r.Message != "Invalid Parameter basketAction" {
		t.Fatalf("expected Invalid Parameter basketAction, got %v", resNoAction)
	}

	// v1 lots <= 0
	reqZeroLots := httptest.NewRequest("POST", "/thematic/v1/invest", strings.NewReader(`{"basketId":1,"basketAction":"BUY","lots":0,"scrips":[{"token":"2188"}]}`))
	resZeroLots, _ := a.invest(reqZeroLots, true)
	if r, ok := resZeroLots.(Response); !ok || r.Message != "Invalid Parameter lots" {
		t.Fatalf("expected Invalid Parameter lots, got %v", resZeroLots)
	}

	// unauthenticated (no token)
	validScripJSON := `{"basketId":999999,"source":"WEB","scrips":[{"exchange":"NSE","token":"2188","qty":"1","price":"100","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}]}`
	reqNoToken := httptest.NewRequest("POST", "/thematic/invest", strings.NewReader(validScripJSON))
	resNoToken, _ := a.invest(reqNoToken, false)
	if hr, ok := resNoToken.(httpResult); !ok || hr.Status != 401 {
		t.Fatalf("expected 401, got %v", resNoToken)
	}

	// basket not found
	reqNotFound := httptest.NewRequest("POST", "/thematic/invest", strings.NewReader(validScripJSON))
	reqNotFound = reqNotFound.WithContext(context.WithValue(reqNotFound.Context(), identityKey{}, Identity{UserID: "USER1", Token: "token123"}))
	resNotFound, _ := a.invest(reqNotFound, false)
	if r, ok := resNotFound.(Response); !ok || r.Message != "Invalid basket" {
		t.Fatalf("expected Invalid basket, got %v", resNotFound)
	}
}

func TestExecuteBasketValidationBranches(t *testing.T) {
	a := testApp(t)

	// invalid parameter
	reqBad := httptest.NewRequest("POST", "/basket/execute", strings.NewReader(`not-json`))
	resBad, _ := a.executeBasket(reqBad)
	if r, ok := resBad.(Response); !ok || r.Message != "Invalid Parameter" {
		t.Fatalf("expected Invalid Parameter, got %v", resBad)
	}

	// unauthorized
	validScripJSON := `{"basketId":999999,"source":"WEB","scrips":[{"exchange":"NSE","token":"2188","qty":"1","price":"100","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}]}`
	reqUnauth := httptest.NewRequest("POST", "/basket/execute", strings.NewReader(validScripJSON))
	resUnauth, _ := a.executeBasket(reqUnauth)
	if hr, ok := resUnauth.(httpResult); !ok || hr.Status != 401 {
		t.Fatalf("expected 401, got %v", resUnauth)
	}

	// invalid basket (not owned)
	reqNotOwned := httptest.NewRequest("POST", "/basket/execute", strings.NewReader(validScripJSON))
	reqNotOwned = reqNotOwned.WithContext(context.WithValue(reqNotOwned.Context(), identityKey{}, Identity{UserID: "USER1", Token: "token123"}))
	resNotOwned, _ := a.executeBasket(reqNotOwned)
	if r, ok := resNotOwned.(Response); !ok || r.Message != "Invalid basket" {
		t.Fatalf("expected Invalid basket, got %v", resNotOwned)
	}
}

func TestCommonHelpersFullCoverage(t *testing.T) {
	// epoch
	if epoch(nil) != nil {
		t.Fatal("expected nil for nil epoch")
	}
	now := time.Now()
	if epoch(now) != now.UnixMilli() {
		t.Fatal("expected matching unix milli for time.Time")
	}
	var nilTs *model.Timestamp
	if epoch(nilTs) != nil {
		t.Fatal("expected nil for nil timestamp")
	}
	ts := model.Timestamp(now)
	if epoch(&ts) != now.UnixMilli() {
		t.Fatal("expected matching unix milli for *model.Timestamp")
	}
	// string date format
	dateStr := now.Format("2006-01-02 15:04:05")
	if epoch(dateStr) == nil {
		t.Fatal("expected parsed unix milli for date string")
	}
	if epoch("not-a-date") != "not-a-date" {
		t.Fatal("expected original value returned for invalid date string")
	}

	// dateOnly
	if dateOnly(nil) != nil {
		t.Fatal("expected nil for nil dateOnly")
	}
	if dateOnly("2026-09-22 15:00:00") != "2026-09-22" {
		t.Fatal("expected 10 char substring for dateOnly")
	}
	if dateOnly("short") != "short" {
		t.Fatal("expected original string if shorter than 10")
	}

	// str
	if str(nil) != "" {
		t.Fatal("expected empty string for nil")
	}
	sVal := "hello"
	if str(&sVal) != "hello" {
		t.Fatal("expected dereferenced string")
	}
	if str([]byte("bytes")) != "bytes" {
		t.Fatal("expected string from bytes")
	}
	if str(now) != now.Format("2006-01-02 15:04:05") {
		t.Fatal("expected formatted time string")
	}
	if str(&ts) != now.Format("2006-01-02 15:04:05") {
		t.Fatal("expected formatted timestamp string")
	}
	if str(nilTs) != "" {
		t.Fatal("expected empty string for nil timestamp")
	}
	if str(12345) != "12345" {
		t.Fatal("expected string for int")
	}

	// coerce
	if coerce(nil, reflect.TypeOf("")) != nil {
		t.Fatal("expected nil")
	}
	tsType := reflect.TypeOf(model.Timestamp{})
	if coerce("", tsType) != nil {
		t.Fatal("expected nil for empty timestamp string")
	}
	if coerce("2026-01-01", tsType) != "2026-01-01" {
		t.Fatal("expected string value for non-empty timestamp")
	}
	// slice coercion
	sliceType := reflect.TypeOf([]string{})
	coercedSlice := coerce([]any{"a", "b"}, sliceType)
	if xs, ok := coercedSlice.([]any); !ok || len(xs) != 2 {
		t.Fatal("expected slice coercion")
	}
	// string coercion from json.Number and bool
	strType := reflect.TypeOf("")
	if coerce(json.Number("42"), strType) != "42" {
		t.Fatal("expected string from json.Number")
	}
	if coerce(true, strType) != "true" {
		t.Fatal("expected string from bool")
	}
	// int/float coercion from string
	intType := reflect.TypeOf(int(0))
	if coerce("", intType) != nil {
		t.Fatal("expected nil for empty string in int")
	}
	if coerce("123", intType) != json.Number("123") {
		t.Fatal("expected json.Number for valid number string")
	}
	if coerce("not-num", intType) != "not-num" {
		t.Fatal("expected original string for invalid number")
	}
}

func TestThematicAdditionalBranches(t *testing.T) {
	a := testApp(t)

	// thematicAll with no mapping
	reqAll := httptest.NewRequest("GET", "/thematic/basket/getall", nil)
	reqAll = reqAll.WithContext(context.WithValue(reqAll.Context(), identityKey{}, Identity{UserID: "NOMAPUSER"}))
	resAll, _ := a.thematic(reqAll, "thematicAll")
	if r, ok := resAll.(Response); !ok || r.Message != "No data found for this user" {
		t.Fatalf("expected 'No data found for this user', got %v", resAll)
	}

	// thematicDetails not found
	reqDetails := httptest.NewRequest("GET", "/thematic/basket/get/99999", nil)
	reqDetails.SetPathValue("id", "99999")
	resDetails, _ := a.thematic(reqDetails, "thematicDetails")
	if r, ok := resDetails.(Response); !ok || r.Message != "Basket not found for ID: 99999" {
		t.Fatalf("expected 'Basket not found for ID: 99999', got %v", resDetails)
	}

	// rebalance count == 0
	reqRebal := httptest.NewRequest("POST", "/thematic/basket/rebalance/details", strings.NewReader(`{"basketId":99999}`))
	reqRebal = reqRebal.WithContext(context.WithValue(reqRebal.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resRebal, _ := a.thematic(reqRebal, "rebalance")
	if r, ok := resRebal.(Response); !ok || r.Message != "Rebalance not available for this user or basket." {
		t.Fatalf("expected 'Rebalance not available for this user or basket.', got %v", resRebal)
	}
}

func TestAdminCleanupsAndLogging(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()

	// 1. deleteExpiredBaskets when empty
	resBasketsEmpty := a.deleteExpiredBaskets(ctx)
	if resBasketsEmpty.Message != "No records found" {
		t.Fatalf("expected 'No records found', got %v", resBasketsEmpty)
	}

	// Insert expired basket
	past := a.Now().Add(-48 * time.Hour)
	pastTs := model.Timestamp(past)
	basket := model.BasketNameEntity{
		BasketName: ptr("ExpiredBasket"),
		UserId:     ptr("USER1"),
		ExpiryDate: &pastTs,
	}
	_ = a.DB.Create(&basket)

	resBaskets := a.deleteExpiredBaskets(ctx)
	if resBaskets.Message != "Success" {
		t.Fatalf("expected 'Success', got %v", resBaskets)
	}

	// 2. deleteExpiredScrips
	scrip := model.BasketScripEntity{
		BasketId: basket.BasketId,
		Token:    ptr("2188"),
		Expiry:   &pastTs,
	}
	_ = a.DB.Create(&scrip)

	resScrips := a.deleteExpiredScrips(ctx)
	if !strings.Contains(str(resScrips.Message), "Record Deleted") {
		t.Fatalf("expected record deleted, got %v", resScrips)
	}

	// 3. a.admin with deleteExpired action
	reqAdmin := httptest.NewRequest("DELETE", "/basketorderapi/deleteExpiredBasket", nil)
	resAdmin, _ := a.admin(reqAdmin, "deleteExpired")
	if _, ok := resAdmin.(Response); !ok {
		t.Fatalf("expected Response, got %v", resAdmin)
	}

	// 4. ClickHouse logging path
	chServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer chServer.Close()

	a.Config.Logging.ClickHouseURL = chServer.URL
	a.Config.Logging.ClickHouseDatabase = "default"
	a.Config.Logging.ClickHouseRestTable = "tbl_rest_log"
	a.Config.Logging.MaxBodyBytes = 100

	record := M{
		"req_body": "test-body",
		"res_body": "test-response",
		"in_time":  time.Now(),
	}
	a.writeLog(ctx, "tbl_rest_log", "tbl_rest_log", record, 200)

	// invalid ClickHouse identifier
	a.Config.Logging.ClickHouseDatabase = "invalid;identifier"
	a.writeLog(ctx, "tbl_rest_log", "tbl_rest_log", record, 200)
}

func TestBusinessUnitExtendedCoverage(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	now := a.Now()
	nowTs := model.Timestamp(now)

	// DetermineAction branches
	if act := determineAction(0, 5); act != "Add New" {
		t.Fatalf("expected Add New, got %v", act)
	}
	if act := determineAction(5, 10); act != "Add More" {
		t.Fatalf("expected Add More, got %v", act)
	}
	if act := determineAction(10, 5); act != "Reduce" {
		t.Fatalf("expected Reduce, got %v", act)
	}
	if act := determineAction(5, 0); act != "Exit" {
		t.Fatalf("expected Exit, got %v", act)
	}
	if act := determineAction(5, 5); act != nil {
		t.Fatalf("expected nil, got %v", act)
	}

	// Executed and timestamp and holdingDate
	if !executed("EXECUTED") || executed("PENDING") {
		t.Fatal("executed check failed")
	}
	if holdingDate(nil) != nil || holdingDate("not-a-date") != nil {
		t.Fatal("holdingDate check failed")
	}
	if holdingDate(now) == nil {
		t.Fatal("expected formatted date from holdingDate")
	}

	// Setup data for buildHolding
	master := model.ThematicMaster{
		Id:         10,
		Status:     ptr("open"),
		BasketName: ptr("Test Basket 10"),
	}
	_ = a.DB.Create(&master)

	scrip1 := model.ThematicScrip{
		BasketId:      10,
		Version:       2,
		Token:         ptr("2188"),
		Exchange:      ptr("NSE"),
		TradingSymbol: ptr("GENCON-EQ"),
		Weightage:     ptr("50"),
		Qty:           ptr("10"),
	}
	scripExit := model.ThematicScrip{
		BasketId:      10,
		Version:       1,
		Token:         ptr("9999"),
		Exchange:      ptr("NSE"),
		TradingSymbol: ptr("EXIT-EQ"),
		Weightage:     ptr("50"),
		Qty:           ptr("5"),
	}
	_ = a.DB.Create(&scrip1)
	_ = a.DB.Create(&scripExit)

	iOne := 1
	iTwo := 2
	iFive := 5
	iTen := 10
	iTwenty := 20
	execDetail1 := model.ExecutionDetail{
		ThematicExecDetails: model.ThematicExecDetails{
			UserId:        ptr("USER1"),
			ResearchId:    &iTen,
			OrderStatus:   ptr("EXECUTED"),
			BasketAction:  ptr("BUY"),
			TransType:     ptr("BUY"),
			LotSize:       &iTwo,
			ExecutionId:   ptr("EXEC-10"),
			Token:         ptr("2188"),
			Exch:          ptr("NSE"),
			TradingSymbol: ptr("GENCON-EQ"),
			ExecutedPrice: ptr("100.5"),
			ExecutedQty:   &iTwenty,
			ExecutedOn:    &nowTs,
			Version:       &iOne,
		},
		ActiveStatus: 1,
	}
	execDetailExit := model.ExecutionDetail{
		ThematicExecDetails: model.ThematicExecDetails{
			UserId:        ptr("USER1"),
			ResearchId:    &iTen,
			OrderStatus:   ptr("EXECUTED"),
			BasketAction:  ptr("BUY"),
			TransType:     ptr("BUY"),
			LotSize:       &iTwo,
			ExecutionId:   ptr("EXEC-10"),
			Token:         ptr("9999"),
			Exch:          ptr("NSE"),
			TradingSymbol: ptr("EXIT-EQ"),
			ExecutedPrice: ptr("50.0"),
			ExecutedQty:   &iFive,
			ExecutedOn:    &nowTs,
			Version:       &iOne,
		},
		ActiveStatus: 1,
	}
	_ = a.DB.Create(&execDetail1)
	_ = a.DB.Create(&execDetailExit)

	userMaster := model.ThematicExeMasterEntity{
		UserId:       ptr("USER1"),
		BasketId:     10,
		BasketAction: ptr("BUY"),
		ActiveStatus: 1,
	}
	_ = a.DB.Create(&userMaster)

	// Call holdings with v1=true and v1=false
	reqH1 := httptest.NewRequest("GET", "/thematic/holdings/get/V1", nil)
	reqH1 = reqH1.WithContext(context.WithValue(reqH1.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resH1, err := a.holdings(reqH1, true)
	if err != nil {
		t.Fatalf("holdings v1 failed: %v", err)
	}
	if okH, ok := resH1.(Response); !ok || okH.Status != "Ok" {
		t.Fatalf("expected Ok from holdings v1, got %v", resH1)
	}

	reqH0 := httptest.NewRequest("GET", "/thematic/holdings/get", nil)
	reqH0 = reqH0.WithContext(context.WithValue(reqH0.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resH0, err := a.holdings(reqH0, false)
	if err != nil {
		t.Fatalf("holdings v0 failed: %v", err)
	}
	if okH, ok := resH0.(Response); !ok || okH.Status != "Ok" {
		t.Fatalf("expected Ok from holdings v0, got %v", resH0)
	}

	// Holdings user info not found
	reqHNone := httptest.NewRequest("GET", "/thematic/holdings/get", nil)
	reqHNone = reqHNone.WithContext(context.WithValue(reqHNone.Context(), identityKey{}, Identity{UserID: "NOUSER"}))
	resHNone, _ := a.holdings(reqHNone, false)
	if m, ok := resHNone.(M); !ok || m["message"] != "User info not found" {
		t.Fatalf("expected User info not found, got %v", resHNone)
	}

	// LegacyHoldings
	userExec := model.ThematicExecution{
		UserId:         "USER1",
		BasketId:       10,
		IsExecuted:     1,
		InvestedAmount: 500.5,
		OrderResponse:  `[{"status":"SUCCESS"}]`,
		OrderRequest:   `[{"exchange":"NSE","token":"2188","qty":"10","price":"100.5","tradingSymbol":"GENCON-EQ"}]`,
		Lots:           2,
		CreatedOn:      &nowTs,
	}
	_ = a.DB.Create(&userExec)

	reqLegacy := httptest.NewRequest("GET", "/thematic/basket/holdings", nil)
	reqLegacy = reqLegacy.WithContext(context.WithValue(reqLegacy.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resLegacy, err := a.thematic(reqLegacy, "legacyHoldings")
	if err != nil {
		t.Fatalf("legacyHoldings failed: %v", err)
	}
	if okL, ok := resLegacy.(Response); !ok || okL.Status != "Ok" {
		t.Fatalf("expected Ok from legacyHoldings, got %v", resLegacy)
	}
	if a.displayDate(nil) != nil || a.displayDate("not-a-date") != nil {
		t.Fatal("displayDate check failed")
	}
	if a.displayDate(now) == nil {
		t.Fatal("expected date string from displayDate")
	}

	// Orderbook UpdateFromFeed & normalizeOrder & updateExecutionDetailRow
	err = a.UpdateFromFeed(ctx, 10, "ORD-FEED-1", model.OrderStatusFeedEntity{
		OrderStatus: ptr("COMPLETE"),
		Exch:        ptr("NSE"),
		Qty:         ptr("10"),
		TradedPrice: ptr("100.5"),
		Reason:      ptr(""),
	})
	if err != nil {
		t.Fatalf("UpdateFromFeed failed: %v", err)
	}

	if norm := normalizeOrder("  ORD-FEED-1;:  "); norm != "ORD-FEED-1" {
		t.Fatalf("expected ORD-FEED-1, got %q", norm)
	}

	detailRow := model.ExecutionDetail{
		ThematicExecDetails: model.ThematicExecDetails{
			ExecutionId: ptr("10"),
			UserId:      ptr("USER1"),
			OrderNo:     ptr("ORD-FEED-2"),
			LotSize:     &iOne,
		},
		ActiveStatus: 1,
	}
	_ = a.DB.Create(&detailRow)

	err = a.updateExecutionDetailRow(ctx, 10, "ORD-FEED-2", M{
		"orderStatus":   "COMPLETE",
		"exchange":      "NSE",
		"tradingSymbol": "GENCON-EQ",
		"token":         "2188",
		"transType":     "BUY",
		"qty":           "10",
	}, "100.5")
	if err != nil {
		t.Fatalf("updateExecutionDetailRow failed: %v", err)
	}

	// Research groupResearchOrders
	rcMaster := model.ResearchMaster{
		ResearchcallOrderEntity: model.ResearchcallOrderEntity{
			Id:           201,
			Category:     ptr("SectorCall"),
			SubCategory:  ptr("Auto"),
			Status:       ptr("open"),
			ActiveStatus: 1,
			CreatedOn:    &nowTs,
		},
		AnalystName: ptr("LeadAnalyst"),
	}
	_ = a.DB.Create(&rcMaster)
	_ = a.DB.Create(&model.ReasearchCallUsers{ResearchCallId: 201, UserId: ptr("USER1")})
	_ = a.DB.Create(&model.ResearchcallScripEntity{
		ResearchcallId: 201,
		Token:          ptr("2188"),
		Exchange:       ptr("NSE"),
		ActiveStatus:   1,
	})

	reqRc := httptest.NewRequest("POST", "/research/getResearchCall", strings.NewReader(`{"status":"open"}`))
	reqRc = reqRc.WithContext(context.WithValue(reqRc.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resRc, err := a.research(reqRc, "research")
	if err != nil {
		t.Fatalf("research call failed: %v", err)
	}
	if okR, ok := resRc.(Response); !ok || okR.Status != "Ok" {
		t.Fatalf("expected Ok from research call, got %v", resRc)
	}

	reqRcB := httptest.NewRequest("POST", "/research/getResearchWithBasket", strings.NewReader(`{}`))
	reqRcB = reqRcB.WithContext(context.WithValue(reqRcB.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resRcB, err := a.research(reqRcB, "researchBasket")
	if err != nil {
		t.Fatalf("researchBasket call failed: %v", err)
	}
	if okR, ok := resRcB.(Response); !ok || okR.Status != "Ok" {
		t.Fatalf("expected Ok from researchBasket call, got %v", resRcB)
	}
	if !notFound(a.DB.First(&model.BasketNameEntity{}, -1).Error) || notFound(nil) {
		t.Fatal("notFound check failed")
	}

	// Basket lockBasketRow & executeBasket
	_ = lockBasketRow(a.DB)

	orderUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Response{{Status: "SUCCESS", Message: "Order Placed"}})
	}))
	defer orderUpstream.Close()
	a.Config.Upstream.OrderURL = orderUpstream.URL

	ownedBasket := model.BasketNameEntity{
		BasketId:   60,
		BasketName: ptr("ExecutionTestBasket"),
		UserId:     ptr("USER1"),
	}
	_ = a.DB.Create(&ownedBasket)

	reqExecute := httptest.NewRequest("POST", "/basketorder/execute", strings.NewReader(`{"basketId":60,"scrips":[{"exchange":"NSE","token":"2188","qty":"1","price":"100","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}]}`))
	reqExecute = reqExecute.WithContext(context.WithValue(reqExecute.Context(), identityKey{}, Identity{UserID: "USER1", Token: "order-token-123"}))
	resExecute, err := a.executeBasket(reqExecute)
	if err != nil {
		t.Fatalf("executeBasket failed: %v", err)
	}
	if r, ok := resExecute.(Response); !ok || r.Status != "Ok" {
		t.Fatalf("expected Ok from executeBasket, got %v", resExecute)
	}

	// DispatchModule direct calls
	dummyReq := httptest.NewRequest("GET", "/dummy", nil)
	dummyReq = dummyReq.WithContext(context.WithValue(dummyReq.Context(), identityKey{}, Identity{UserID: "USER1"}))
	for _, mod := range []string{"basket", "admin", "margin", "research", "thematic", "holdings", "cache", "unknown"} {
		_, _ = a.dispatchModule(dummyReq, route{Module: mod, Action: "invalid"})
	}

	// Special actions
	recToken := httptest.NewRecorder()
	reqToken := httptest.NewRequest("GET", "/token", nil)
	reqToken = reqToken.WithContext(context.WithValue(reqToken.Context(), identityKey{}, Identity{UserID: "USER1", Scope: "all"}))
	if !handleSpecialActions(recToken, reqToken, route{Action: "token"}) || recToken.Code != 200 {
		t.Fatal("expected token special action handled")
	}

	recLogout := httptest.NewRecorder()
	reqLogout := httptest.NewRequest("GET", "/token/logout", nil)
	if !handleSpecialActions(recLogout, reqLogout, route{Action: "logout"}) || recLogout.Code != 200 {
		t.Fatal("expected logout special action handled")
	}

	recCat := httptest.NewRecorder()
	reqCat := httptest.NewRequest("GET", "/category", nil)
	if !handleSpecialActions(recCat, reqCat, route{Action: "category"}) || recCat.Code != 204 {
		t.Fatal("expected category special action handled")
	}

	if handleSpecialActions(recCat, reqCat, route{Action: "other"}) {
		t.Fatal("expected false for other action")
	}

	// Cache MemoryCache operations
	memCache := &MemoryCache{Values: map[string]map[string]json.RawMessage{}}
	_ = memCache.Put(ctx, "testmap", "k1", "v1")
	_ = memCache.Delete(ctx, "testmap", "k1")
	_ = memCache.Close(ctx)
}

func TestBasketAndThematicAdditionalCoverage(t *testing.T) {
	a := testApp(t)

	// 1. Basket Create, Rename, AddScrip, UpdateScrip, Scrips, Reset, Delete
	// Create Basket
	createJSON := `{"basketName":"AlphaBasket"}`
	reqCreate := httptest.NewRequest("POST", "/basketorder/create", strings.NewReader(createJSON))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resCreate, err := a.basket(reqCreate, "create")
	if err != nil {
		t.Fatalf("basket create failed: %v", err)
	}
	if _, ok := resCreate.(Response); !ok {
		t.Fatalf("expected Response, got %v", resCreate)
	}

	// Create duplicate basket name
	reqDup := httptest.NewRequest("POST", "/basketorder/create", strings.NewReader(createJSON))
	reqDup = reqDup.WithContext(context.WithValue(reqDup.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resDup, _ := a.basket(reqDup, "create")
	if r, ok := resDup.(Response); !ok || r.Message != "Basket name already exist" {
		t.Fatalf("expected 'Basket name already exist', got %v", resDup)
	}

	// Fetch created basket ID
	var bEntity model.BasketNameEntity
	a.DB.First(&bEntity, "basket_name = ? AND user_id = ?", "AlphaBasket", "USER1")
	bID := bEntity.BasketId

	// Rename Basket
	renameJSON := fmt.Sprintf(`{"basketId":%d,"basketName":"BetaBasket"}`, bID)
	reqRename := httptest.NewRequest("POST", "/basketorder/rename", strings.NewReader(renameJSON))
	reqRename = reqRename.WithContext(context.WithValue(reqRename.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resRename, err := a.basket(reqRename, "rename")
	if err != nil {
		t.Fatalf("basket rename failed: %v", err)
	}
	if _, ok := resRename.(Response); !ok {
		t.Fatalf("expected Response, got %v", resRename)
	}

	// AddScrip
	addScripJSON := fmt.Sprintf(`{"basketId":%d,"scrips":{"exchange":"NSE","token":"2188","qty":"5","price":"100","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}}`, bID)
	reqAdd := httptest.NewRequest("POST", "/basketorder/add/scrips", strings.NewReader(addScripJSON))
	reqAdd = reqAdd.WithContext(context.WithValue(reqAdd.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resAdd, err := a.basket(reqAdd, "addScrip")
	if err != nil {
		t.Fatalf("basket addScrip failed: %v", err)
	}
	if _, ok := resAdd.(Response); !ok {
		t.Fatalf("expected Response, got %v", resAdd)
	}

	// Fetch added scrip ID
	var sEntity model.BasketScripEntity
	a.DB.First(&sEntity, "basket_id = ?", bID)

	// UpdateScrip
	updateScripJSON := fmt.Sprintf(`{"basketId":%d,"scrips":{"id":%d,"exchange":"NSE","token":"2188","qty":"10","price":"105","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}}`, bID, sEntity.Id)
	reqUpdate := httptest.NewRequest("POST", "/basketorder/update/scrips", strings.NewReader(updateScripJSON))
	reqUpdate = reqUpdate.WithContext(context.WithValue(reqUpdate.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resUpdate, err := a.basket(reqUpdate, "updateScrip")
	if err != nil {
		t.Fatalf("basket updateScrip failed: %v", err)
	}
	if _, ok := resUpdate.(Response); !ok {
		t.Fatalf("expected Response, got %v", resUpdate)
	}

	// UpdateScripList
	updateListJSON := fmt.Sprintf(`{"basketId":%d,"scrips":[{"id":%d,"exchange":"NSE","token":"2188","qty":"15","price":"110","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}]}`, bID, sEntity.Id)
	reqUpdateList := httptest.NewRequest("POST", "/basketorder/update/scrips/list", strings.NewReader(updateListJSON))
	reqUpdateList = reqUpdateList.WithContext(context.WithValue(reqUpdateList.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resUpdateList, err := a.basket(reqUpdateList, "updateScripList")
	if err != nil {
		t.Fatalf("basket updateScripList failed: %v", err)
	}
	if _, ok := resUpdateList.(Response); !ok {
		t.Fatalf("expected Response, got %v", resUpdateList)
	}

	// Scrips (live and non-live)
	reqScrips := httptest.NewRequest("GET", fmt.Sprintf("/basketorder/get/scrips/%d", bID), nil)
	reqScrips.SetPathValue("basketId", strconv.Itoa(int(bID)))
	reqScrips = reqScrips.WithContext(context.WithValue(reqScrips.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resScrips, err := a.basket(reqScrips, "scrips")
	if err != nil {
		t.Fatalf("basket scrips failed: %v", err)
	}
	if _, ok := resScrips.(Response); !ok {
		t.Fatalf("expected Response, got %v", resScrips)
	}

	// Reset
	reqReset := httptest.NewRequest("GET", fmt.Sprintf("/basketorder/reset/%d", bID), nil)
	reqReset.SetPathValue("basketId", strconv.Itoa(int(bID)))
	reqReset = reqReset.WithContext(context.WithValue(reqReset.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resReset, err := a.basket(reqReset, "reset")
	if err != nil {
		t.Fatalf("basket reset failed: %v", err)
	}
	if _, ok := resReset.(Response); !ok {
		t.Fatalf("expected Response, got %v", resReset)
	}

	// DeleteScrip
	delScripJSON := fmt.Sprintf(`{"basketId":%d,"scripsId":[%d]}`, bID, sEntity.Id)
	reqDelScrip := httptest.NewRequest("POST", "/basketorder/delete/scrips", strings.NewReader(delScripJSON))
	reqDelScrip = reqDelScrip.WithContext(context.WithValue(reqDelScrip.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resDelScrip, err := a.basket(reqDelScrip, "deleteScrip")
	if err != nil {
		t.Fatalf("basket deleteScrip failed: %v", err)
	}
	if _, ok := resDelScrip.(Response); !ok {
		t.Fatalf("expected Response, got %v", resDelScrip)
	}

	// Delete basket
	reqDel := httptest.NewRequest("DELETE", fmt.Sprintf("/basketorder/delete/%d", bID), nil)
	reqDel.SetPathValue("basketId", strconv.Itoa(int(bID)))
	reqDel = reqDel.WithContext(context.WithValue(reqDel.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resDel, err := a.basket(reqDel, "delete")
	if err != nil {
		t.Fatalf("basket delete failed: %v", err)
	}
	if _, ok := resDel.(Response); !ok {
		t.Fatalf("expected Response, got %v", resDel)
	}

	// 2. ThematicDetails with SectorReports and Rebalance
	tm := model.ThematicMaster{
		Id:                 30,
		Status:             ptr("Open"),
		BasketName:         ptr("DetailsBasket"),
		RebalanceAvailable: 1,
	}
	_ = a.DB.Create(&tm)

	reportDoc := model.SectorReportsEntity{
		ReferanceId:  ptr(int64(30)),
		Name:         ptr("ReportDoc"),
		ActiveStatus: 1,
	}
	_ = a.DB.Create(&reportDoc)

	reqDetails := httptest.NewRequest("GET", "/thematic/basket/get/30", nil)
	reqDetails.SetPathValue("id", "30")
	reqDetails = reqDetails.WithContext(context.WithValue(reqDetails.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resDetails, err := a.thematic(reqDetails, "thematicDetails")
	if err != nil {
		t.Fatalf("thematicDetails failed: %v", err)
	}
	if okD, ok := resDetails.(Response); !ok || okD.Status != "Ok" {
		t.Fatalf("expected Ok from thematicDetails, got %v", resDetails)
	}

	// Rebalance scrips
	rebalScrip := model.RebalanceScrip{
		BasketId:      30,
		Token:         ptr("2188"),
		Exchange:      ptr("NSE"),
		TradingSymbol: ptr("GENCON-EQ"),
	}
	_ = a.DB.Create(&rebalScrip)

	userExec := model.ThematicExecution{
		UserId:     "USER1",
		BasketId:   30,
		IsExecuted: 1,
	}
	_ = a.DB.Create(&userExec)

	reqRebal := httptest.NewRequest("POST", "/thematic/basket/rebalance/details", strings.NewReader(`{"basketId":30}`))
	reqRebal = reqRebal.WithContext(context.WithValue(reqRebal.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resRebal, err := a.thematic(reqRebal, "rebalance")
	if err != nil {
		t.Fatalf("thematic rebalance failed: %v", err)
	}
	if okR, ok := resRebal.(Response); !ok || okR.Status != "Ok" {
		t.Fatalf("expected Ok from thematic rebalance, got %v", resRebal)
	}

	// View
	reqView := httptest.NewRequest("POST", "/thematic/basket/report", strings.NewReader(`{"basketId":30,"isViewed":1}`))
	reqView = reqView.WithContext(context.WithValue(reqView.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resView, err := a.thematic(reqView, "view")
	if err != nil {
		t.Fatalf("thematic view failed: %v", err)
	}
	if okV, ok := resView.(Response); !ok || okV.Status != "Ok" {
		t.Fatalf("expected Ok from thematic view, got %v", resView)
	}
}

func TestAdminAndUpstreamCoverage(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	now := a.Now()
	nowTs := model.Timestamp(now)

	// Upstream margin with NFO and MCX
	spanSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(M{"stat": "Ok", "span_trade": "500.5", "expo_trade": "200.2"})
	}))
	defer spanSrv.Close()
	a.Config.Upstream.SpanURL = spanSrv.URL

	contract := model.ContractMasterModel{
		Exch:        ptr("NFO"),
		Token:       ptr("2188"),
		InsType:     ptr("OPTIDX"),
		OptionType:  ptr("CE"),
		StrikePrice: ptr("100"),
		LotSize:     ptr("50"),
		Expiry:      &nowTs,
		Symbol:      ptr("NIFTY"),
	}
	_ = a.Cache.Put(ctx, a.Config.Cache.Maps["contracts"], "NFO_2188", contract)
	_ = a.Cache.Put(ctx, a.Config.Cache.Maps["sessions"], "USER1_REST_SESSION", "active-session")

	reqMargin := httptest.NewRequest("POST", "/basketorder/spanmargin", strings.NewReader(`[{"exchange":"NFO","token":"2188","qty":"50","price":"10","transType":"BUY"}]`))
	reqMargin = reqMargin.WithContext(context.WithValue(reqMargin.Context(), identityKey{}, Identity{UserID: "USER1"}))
	resMargin, err := a.margin(reqMargin, "span")
	if err != nil {
		t.Fatalf("margin call failed: %v", err)
	}
	if okM, ok := resMargin.(Response); !ok || okM.Status != "Ok" {
		t.Fatalf("expected Ok from margin call, got %v", resMargin)
	}

	// Admin create and temp create
	_ = a.DB.Create(&model.VendorAppEntity{ApiKey: ptr("valid-api-key"), TppAuthorization: 1})
	_ = a.DB.Create(&model.DeviceMappingEntity{UserId: ptr("USER1"), DeviceId: ptr("dev-id-1")})

	adminPayload := fmt.Sprintf(`{"apiKey":"valid-api-key","basketName":"AdminTestBasket","expiryDate":%d,"userId":["USER1"],"pushNotification":0,"scrips":[{"exchange":"NSE","token":"2188","qty":"1","price":"100","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}]}`, now.Add(24*time.Hour).UnixMilli())

	reqAdmin := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(adminPayload))
	reqAdmin.Header.Set("Authorization", "admin-test")
	resAdmin, err := a.admin(reqAdmin, "adminCreate")
	if err != nil {
		t.Fatalf("adminCreate failed: %v", err)
	}
	if okA, ok := resAdmin.(Response); !ok || okA.Status != "Ok" {
		t.Fatalf("expected Ok from adminCreate, got %v", resAdmin)
	}

	reqAdminTemp := httptest.NewRequest("POST", "/basketorderapi/adminCreate/temp", strings.NewReader(adminPayload))
	reqAdminTemp.Header.Set("Authorization", "admin-test")
	resAdminTemp, err := a.admin(reqAdminTemp, "adminTemp")
	if err != nil {
		t.Fatalf("adminTemp failed: %v", err)
	}
	if okA, ok := resAdminTemp.(Response); !ok || okA.Status != "Ok" {
		t.Fatalf("expected Ok from adminTemp, got %v", resAdminTemp)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestRemainingBranchesForHighCoverage(t *testing.T) {
	a := testApp(t)
	ctx := context.Background()
	now := a.Now()

	// 1. Notifications: legacy provider & errors
	legacyOkSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(M{"failure": 0})
	}))
	defer legacyOkSrv.Close()

	a.Config.Modules.Notifications = true
	a.Config.Notifications.Provider = "legacy"
	a.Config.Notifications.URL = legacyOkSrv.URL
	a.Config.Notifications.BatchSize = 2
	a.Config.Notifications.APIKey = "legacy-key"

	// Success batching
	if err := a.notify(ctx, []string{"dev1", "dev2", "dev3"}, "Alert", "Body", M{"k": "v"}); err != nil {
		t.Fatalf("notify legacy failed: %v", err)
	}

	// Legacy failure response
	legacyFailSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(M{"failure": 1})
	}))
	defer legacyFailSrv.Close()
	a.Config.Notifications.URL = legacyFailSrv.URL
	if err := a.notify(ctx, []string{"dev1"}, "Alert", "Body", M{}); err == nil {
		t.Fatal("expected notify failure when result has failures > 0")
	}

	// Unsupported provider
	a.Config.Notifications.Provider = "custom_unsupported"
	if err := a.notify(ctx, []string{"dev1"}, "Alert", "Body", M{}); err == nil {
		t.Fatal("expected error for unsupported notification provider")
	}

	// No devices / disabled
	a.Config.Modules.Notifications = false
	if err := a.notify(ctx, []string{"dev1"}, "Alert", "Body", M{}); err != nil {
		t.Fatalf("expected nil when disabled, got %v", err)
	}
	a.Config.Modules.Notifications = true
	if err := a.notify(ctx, nil, "Alert", "Body", M{}); err != nil {
		t.Fatalf("expected nil when devices empty, got %v", err)
	}

	// 2. Admin: deleteExpiredScrips & deleteExpiredBaskets & admin actions
	expTs := model.Timestamp(now.Add(-48 * time.Hour))
	_ = a.DB.Create(&model.BasketScripEntity{BasketId: 991, Expiry: &expTs})
	_ = a.deleteExpiredScrips(ctx)

	_ = a.DB.Create(&model.BasketNameEntity{BasketId: 992, UserId: ptr("USER1"), ExpiryDate: &expTs})
	_ = a.DB.Create(&model.UserNotification{BasketId: ptr("992")})
	reqDelExp := httptest.NewRequest("GET", "/basketorderapi/deleteExpired", nil)
	if res, err := a.admin(reqDelExp, "deleteExpired"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok from admin deleteExpired, got res=%v, err=%v", res, err)
	}

	// admin unauthorized
	reqAdminUnauth := httptest.NewRequest("POST", "/basketorderapi/adminCreate", nil)
	reqAdminUnauth.Header.Set("Authorization", "wrong-token")
	if res, _ := a.admin(reqAdminUnauth, "adminCreate"); res.(httpResult).Status != 401 {
		t.Fatalf("expected 401 for unauthorized admin, got %v", res)
	}

	// admin invalid json
	reqAdminBadJSON := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader("bad-json"))
	reqAdminBadJSON.Header.Set("Authorization", "admin-test")
	if res, _ := a.admin(reqAdminBadJSON, "adminCreate"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for bad JSON, got %v", res)
	}

	// admin unknown vendor key
	reqAdminUnknownVendor := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(`{"apiKey":"unknown-api-key"}`))
	reqAdminUnknownVendor.Header.Set("Authorization", "admin-test")
	if res, _ := a.admin(reqAdminUnknownVendor, "adminCreate"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for unknown vendor, got %v", res)
	}

	// 3. Basket: scrips with refreshLot, updateScrips error, deleteBasket with notifications, executeBasket
	// scrips refreshLot on empty
	if res, err := a.scrips(ctx, 999991, true); err != nil || res.Status != "Ok" {
		t.Fatalf("expected Ok with no records found, got res=%v, err=%v", res, err)
	}

	// scrips refreshLot on non-empty
	_ = a.Cache.Put(ctx, a.Config.Cache.Maps["contracts"], "NSE_2188", model.ContractMasterModel{LotSize: ptr("25")})
	_ = a.DB.Create(&model.BasketScripEntity{BasketId: 881, Exchange: ptr("NSE"), Token: ptr("2188")})
	if res, err := a.scrips(ctx, 881, true); err != nil || res.Status != "Ok" {
		t.Fatalf("expected Ok for scrips with refreshLot, got res=%v, err=%v", res, err)
	}

	// updateScrips with missing scrip ID
	_ = a.DB.Create(&model.BasketNameEntity{BasketId: 882, UserId: ptr("USER1")})
	if res, err := a.updateScrips(ctx, 882, "USER1", []model.ScripRequestModel{{Id: 999992, Exchange: ptr("NSE"), Token: ptr("2188")}}); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected invalid basket response, got res=%v, err=%v", res, err)
	}

	// deleteBasket with notifications
	_ = a.DB.Create(&model.BasketNameEntity{BasketId: 883, UserId: ptr("USER1")})
	_ = a.DB.Create(&model.BasketScripEntity{BasketId: 883, Exchange: ptr("NSE"), Token: ptr("2188")})
	_ = a.DB.Create(&model.UserNotification{BasketId: ptr("883")})
	if err := deleteBasket(a.DB, 883, true); err != nil {
		t.Fatalf("deleteBasket with notifications failed: %v", err)
	}

	// lockBasketRow
	tx := a.DB.Begin()
	_ = lockBasketRow(tx)
	_ = tx.Rollback()

	// executeBasket: invalid param
	reqExecBad := httptest.NewRequest("POST", "/basketorder/executeBasketOrder", strings.NewReader(`{"basketId":0}`))
	if res, _ := a.executeBasket(reqExecBad); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for invalid executeBasket param, got %v", res)
	}

	// executeBasket: invalid scrip
	reqExecBadScrip := httptest.NewRequest("POST", "/basketorder/executeBasketOrder", strings.NewReader(`{"basketId":882,"scrips":[{"exchange":""}]}`))
	if res, _ := a.executeBasket(reqExecBadScrip); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for invalid scrip in executeBasket, got %v", res)
	}

	// executeBasket: missing token in identity
	fullScrip := `{"exchange":"NSE","token":"2188","qty":"1","price":"100","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}`
	reqExecNoToken := httptest.NewRequest("POST", "/basketorder/executeBasketOrder", strings.NewReader(fmt.Sprintf(`{"basketId":882,"scrips":[%s]}`, fullScrip)))
	reqExecNoToken = reqExecNoToken.WithContext(context.WithValue(reqExecNoToken.Context(), identityKey{}, Identity{UserID: "USER1", Token: ""}))
	if res, _ := a.executeBasket(reqExecNoToken); res.(httpResult).Status != 401 {
		t.Fatalf("expected 401 when token is empty, got %v", res)
	}

	// executeBasket: unowned basket
	reqExecUnowned := httptest.NewRequest("POST", "/basketorder/executeBasketOrder", strings.NewReader(fmt.Sprintf(`{"basketId":999993,"scrips":[%s]}`, fullScrip)))
	reqExecUnowned = reqExecUnowned.WithContext(context.WithValue(reqExecUnowned.Context(), identityKey{}, Identity{UserID: "USER1", Token: "valid-tok"}))
	if res, _ := a.executeBasket(reqExecUnowned); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for unowned basket, got %v", res)
	}

	// executeBasket: upstream 401
	order401Srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer order401Srv.Close()
	a.Config.Upstream.OrderURL = order401Srv.URL

	reqExecUpstream401 := httptest.NewRequest("POST", "/basketorder/executeBasketOrder", strings.NewReader(fmt.Sprintf(`{"basketId":882,"scrips":[%s]}`, fullScrip)))
	reqExecUpstream401 = reqExecUpstream401.WithContext(context.WithValue(reqExecUpstream401.Context(), identityKey{}, Identity{UserID: "USER1", Token: "valid-tok"}))
	if res, _ := a.executeBasket(reqExecUpstream401); res.(httpResult).Status != 401 {
		t.Fatalf("expected 401 from upstream 401, got %v", res)
	}

	// executeBasket: upstream nil / null response
	orderNullSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("null"))
	}))
	defer orderNullSrv.Close()
	a.Config.Upstream.OrderURL = orderNullSrv.URL
	if res, _ := a.executeBasket(reqExecUpstream401); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed from upstream null response, got %v", res)
	}

	// 4. Thematic: thematicAll, thematicDetails not found, aggregateOrderRequestScrips, isOrderPlaced
	_ = a.DB.Table("tbl_researchcall_usermapping").Create(M{"user_id": "USER_THEM", "thematic_basket_id": 501})
	_ = a.DB.Table(tblThematicBasketMaster).Create(M{
		"id":              501,
		"basket_name":     "Growth Thematic",
		"active_status":   1,
		"action_type":     "SEND_NOW",
		"status":          "OPEN",
		"total_invst_amt": 10000,
		"created_on":      now,
		"expiry_date":     now.Add(24 * time.Hour),
	})
	_ = a.DB.Table(tblThematicBasketScrips).Create(M{
		"basket_id": 501,
		"exchange":  "NSE",
		"token":     "2188",
		"price":     100,
		"qty":       10,
	})
	reqThemAll := httptest.NewRequest("GET", "/thematic/thematicAll", nil)
	reqThemAll = reqThemAll.WithContext(context.WithValue(reqThemAll.Context(), identityKey{}, Identity{UserID: "USER_THEM"}))
	if res, err := a.thematic(reqThemAll, "thematicAll"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for thematicAll with records, got res=%v, err=%v", res, err)
	}

	reqThemDet404 := httptest.NewRequest("GET", "/thematic/thematicDetails/999994", nil)
	reqThemDet404.SetPathValue("id", "999994")
	if res, err := a.thematic(reqThemDet404, "thematicDetails"); err != nil || res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for missing thematic details, got res=%v, err=%v", res, err)
	}

	// aggregateOrderRequestScrips
	baseScrips := []M{{"token": "2188", "exchange": "NSE", "tradingSymbol": "GENCON-EQ", "qty": "10"}}
	merged := aggregateOrderRequestScrips(baseScrips, `[{"token":"2188","qty":5},{"token":"3000","exchange":"BSE","tradingSymbol":"XYZ","qty":2}]`)
	if len(merged) != 2 || merged[0]["qty"] != "15" {
		t.Fatalf("unexpected aggregateOrderRequestScrips result: %v", merged)
	}

	// isOrderPlaced
	if !isOrderPlaced([]Response{{Result: []M{{"orderNo": "ORD001"}}}}) {
		t.Fatal("expected isOrderPlaced to be true")
	}
	if isOrderPlaced([]Response{{Result: []M{{"orderNo": ""}}}}) {
		t.Fatal("expected isOrderPlaced to be false")
	}

	// 5. OrderBook: bindOrderBook branches & refreshOrderBook
	obMCX := a.bindOrderBook(ctx, M{
		"Exchange":            "MCX",
		"token":               "2188",
		"Qty":                 "2",
		"Status":              "TRIGGER PENDING",
		"OrderedTime":         "23/09/2026 14:30:00",
		"ordergenerationtype": "AMO",
		"Pcode":               "BO",
		"Prc":                 "0.0",
		"Avgprc":              "150.0",
		"Trantype":            "B",
	})
	if obMCX["orderStatus"] != "pending" || obMCX["orderType"] != "AMO" || obMCX["transType"] != "BUY" {
		t.Fatalf("unexpected bindOrderBook MCX result: %v", obMCX)
	}

	obCancelled := a.bindOrderBook(ctx, M{
		"Exchange":            "NSE",
		"token":               "2188",
		"Qty":                 "5",
		"Status":              "CANCELLED AFTER MARKET ORDER",
		"ordergenerationtype": "REGULAR",
		"Pcode":               "CO",
		"Prc":                 "200.0",
		"Trantype":            "S",
	})
	if obCancelled["orderStatus"] != "Cancelled AMO" || obCancelled["orderType"] != "Cover" || obCancelled["transType"] != "SELL" {
		t.Fatalf("unexpected bindOrderBook Cancelled result: %v", obCancelled)
	}

	// refreshOrderBook empty wanted
	if err := a.refreshOrderBook(ctx, 123, "USER1", []M{{"orderNo": ""}}); err != nil {
		t.Fatalf("expected nil for empty wanted orders, got %v", err)
	}

	// refreshOrderBook missing customer cache
	a.Config.Upstream.OrderBookURL = "http://example.com/orderbook"
	if err := a.refreshOrderBook(ctx, 123, "USER1", []M{{"orderNo": "12345"}}); err == nil {
		t.Fatal("expected error when customer cache missing")
	}

	// 6. Research: status, sector, reports with basketType, research with analystName
	_ = a.DB.Table("tbl_researchcall_master").Create(M{"id": 701, "status": "OPEN", "active_status": 1, "analyst_name": "ANALYST_TOP"})
	_ = a.DB.Table("tbl_research_scrip").Create(M{"researchcall_id": 701, "exchange": "NSE", "token": "2188", "qty": 1})
	reqResStatus := httptest.NewRequest("GET", "/research/status", nil)
	if res, err := a.research(reqResStatus, "status"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for research status, got res=%v, err=%v", res, err)
	}

	_ = a.DB.Table("tbl_sector_reports").Create(M{
		"id":            801,
		"name":          "Pharma Sector",
		"type":          "SECTOR",
		"active_status": 1,
		"created_by":    1,
		"updated_by":    1,
		"created_on":    now,
		"updated_on":    now,
	})
	reqResSector := httptest.NewRequest("GET", "/research/sector/801", nil)
	reqResSector.SetPathValue("id", "801")
	if res, err := a.research(reqResSector, "sector"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for research sector, got res=%v, err=%v", res, err)
	}

	reqResReports := httptest.NewRequest("POST", "/research/reports", strings.NewReader(`{"basketType":"SECTOR"}`))
	if res, err := a.research(reqResReports, "reports"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for research reports with basketType, got res=%v, err=%v", res, err)
	}

	_ = a.DB.Table("tbl_researchcall_usermapping").Create(M{"user_id": "USER_RES", "researchcall_id": 701})
	reqResCall := httptest.NewRequest("POST", "/research/research", strings.NewReader(`{"analystName":"ANALYST_TOP","status":"OPEN"}`))
	reqResCall = reqResCall.WithContext(context.WithValue(reqResCall.Context(), identityKey{}, Identity{UserID: "USER_RES"}))
	if res, err := a.research(reqResCall, "research"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for research with analystName, got res=%v, err=%v", res, err)
	}

	// 7. Upstream: nestSpanMargin & call
	reqNestBad := httptest.NewRequest("POST", "/basketorder/spanmargin", strings.NewReader("bad-json"))
	if res, _ := a.nestSpanMargin(reqNestBad, Identity{UserID: "USER1"}); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for bad JSON in nestSpanMargin, got %v", res)
	}

	reqNestMissingField := httptest.NewRequest("POST", "/basketorder/spanmargin", strings.NewReader(`[{"exchange":""}]`))
	if res, _ := a.nestSpanMargin(reqNestMissingField, Identity{UserID: "USER1"}); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Failed for missing field in nestSpanMargin, got %v", res)
	}

	// call with empty URL
	if _, err := a.call(ctx, "", "", "application/json", "", nil); err == nil {
		t.Fatal("expected error for empty upstream URL")
	}

	// call response exceeds limit
	limitSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("response-longer-than-limit"))
	}))
	defer limitSrv.Close()
	a.Config.Upstream.MaxResponseBytes = 5
	if _, err := a.call(ctx, limitSrv.URL, "", "application/json", "", nil); err == nil {
		t.Fatal("expected error when response exceeds MaxResponseBytes")
	}
	a.Config.Upstream.MaxResponseBytes = 1048576

	// call 500 error
	err500Srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer err500Srv.Close()
	if _, err := a.call(ctx, err500Srv.URL, "", "application/json", "", nil); err == nil {
		t.Fatal("expected error when upstream returns HTTP 500")
	}

	// 8. Admin additional branches
	_ = a.DB.Create(&model.VendorAppEntity{ApiKey: ptr("valid-api-key"), TppAuthorization: 1})
	_ = a.DB.Create(&model.DeviceMappingEntity{UserId: ptr("USER1"), DeviceId: ptr("dev-id-1")})
	_ = a.DB.Create(&model.VendorAppEntity{ApiKey: ptr("unauth-vendor-key"), TppAuthorization: 0})
	reqAdminUnauthVendor := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(`{"apiKey":"unauth-vendor-key"}`))
	reqAdminUnauthVendor.Header.Set("Authorization", "admin-test")
	if res, _ := a.admin(reqAdminUnauthVendor, "adminCreate"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for unauth vendor, got %v", res)
	}

	reqAdminMissingName := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(`{"apiKey":"valid-api-key","basketName":"","expiryDate":12345678,"scrips":[]}`))
	reqAdminMissingName.Header.Set("Authorization", "admin-test")
	if res, _ := a.admin(reqAdminMissingName, "adminCreate"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for missing name, got %v", res)
	}

	a.Config.Business.MaxScrips = 1
	reqAdminTooManyScrips := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(fmt.Sprintf(`{"apiKey":"valid-api-key","basketName":"Test","expiryDate":12345678,"scrips":[%s,%s]}`, fullScrip, fullScrip)))
	reqAdminTooManyScrips.Header.Set("Authorization", "admin-test")
	if res, _ := a.admin(reqAdminTooManyScrips, "adminCreate"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for too many scrips, got %v", res)
	}
	a.Config.Business.MaxScrips = 50

	badExchScrip := `{"exchange":"INVALID","token":"2188","qty":"1","price":"100","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}`
	reqAdminBadExch := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(fmt.Sprintf(`{"apiKey":"valid-api-key","basketName":"Test","expiryDate":12345678,"scrips":[%s]}`, badExchScrip)))
	reqAdminBadExch.Header.Set("Authorization", "admin-test")
	if res, _ := a.admin(reqAdminBadExch, "adminCreate"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for bad exchange, got %v", res)
	}

	reqAdminPush := httptest.NewRequest("POST", "/basketorderapi/adminCreate", strings.NewReader(fmt.Sprintf(`{"apiKey":"valid-api-key","basketName":"PushBasket","expiryDate":%d,"userId":[],"pushNotification":1,"title":"Title","message":"Msg","scrips":[%s]}`, now.Add(24*time.Hour).UnixMilli(), fullScrip)))
	reqAdminPush.Header.Set("Authorization", "admin-test")
	if res, err := a.admin(reqAdminPush, "adminCreate"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for admin create with push, got res=%v, err=%v", res, err)
	}

	// 9. Basket additional branches
	reqBasketGetEmpty := httptest.NewRequest("GET", "/basketorder/getBasketOrders", nil)
	reqBasketGetEmpty = reqBasketGetEmpty.WithContext(context.WithValue(reqBasketGetEmpty.Context(), identityKey{}, Identity{UserID: "USER_NO_BASKETS"}))
	if res, err := a.basket(reqBasketGetEmpty, "get"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok with no records found, got res=%v, err=%v", res, err)
	}

	reqBasketIDZero := httptest.NewRequest("GET", "/basketorder/getBasketOrderScrips/0", nil)
	reqBasketIDZero = reqBasketIDZero.WithContext(context.WithValue(reqBasketIDZero.Context(), identityKey{}, Identity{UserID: "USER1"}))
	reqBasketIDZero.SetPathValue("basketId", "0")
	if res, _ := a.basket(reqBasketIDZero, "scrips"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for basket id 0, got %v", res)
	}

	reqBasketRenameBad := httptest.NewRequest("POST", "/basketorder/renameBasket", strings.NewReader(`{"basketId":0,"basketName":"NewName"}`))
	reqBasketRenameBad = reqBasketRenameBad.WithContext(context.WithValue(reqBasketRenameBad.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqBasketRenameBad, "rename"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for rename id 0, got %v", res)
	}

	_ = a.DB.Create(&model.BasketNameEntity{BasketId: 884, UserId: ptr("USER1"), BasketName: ptr("ExistingName")})
	reqBasketRenameDup := httptest.NewRequest("POST", "/basketorder/renameBasket", strings.NewReader(`{"basketId":884,"basketName":"ExistingName"}`))
	reqBasketRenameDup = reqBasketRenameDup.WithContext(context.WithValue(reqBasketRenameDup.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqBasketRenameDup, "rename"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for duplicate basket name, got %v", res)
	}

	reqAddBad := httptest.NewRequest("POST", "/basketorder/addScripToBasket", strings.NewReader(`{"basketId":0,"scrips":{}}`))
	reqAddBad = reqAddBad.WithContext(context.WithValue(reqAddBad.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqAddBad, "addScrip"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for addScrip id 0, got %v", res)
	}

	reqUpdBad := httptest.NewRequest("POST", "/basketorder/updateBasketOrderScrip", strings.NewReader(`{"basketId":884,"scrips":{"id":0}}`))
	reqUpdBad = reqUpdBad.WithContext(context.WithValue(reqUpdBad.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqUpdBad, "updateScrip"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for updateScrip id 0, got %v", res)
	}

	reqUpdListBad := httptest.NewRequest("POST", "/basketorder/updateBasketOrderScripList", strings.NewReader(`bad-json`))
	reqUpdListBad = reqUpdListBad.WithContext(context.WithValue(reqUpdListBad.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqUpdListBad, "updateScripList"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for updateScripList bad json, got %v", res)
	}

	reqDelScripBad := httptest.NewRequest("POST", "/basketorder/deleteBasketOrderScrips", strings.NewReader(`{"basketId":0,"scripsId":[]}`))
	reqDelScripBad = reqDelScripBad.WithContext(context.WithValue(reqDelScripBad.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqDelScripBad, "deleteScrip"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for deleteScrip id 0, got %v", res)
	}

	reqDelScripUnowned := httptest.NewRequest("POST", "/basketorder/deleteBasketOrderScrips", strings.NewReader(`{"basketId":999995,"scripsId":[1]}`))
	reqDelScripUnowned = reqDelScripUnowned.WithContext(context.WithValue(reqDelScripUnowned.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqDelScripUnowned, "deleteScrip"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for deleteScrip unowned basket, got %v", res)
	}

	reqDelScripNoRows := httptest.NewRequest("POST", "/basketorder/deleteBasketOrderScrips", strings.NewReader(`{"basketId":884,"scripsId":[999996]}`))
	reqDelScripNoRows = reqDelScripNoRows.WithContext(context.WithValue(reqDelScripNoRows.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.basket(reqDelScripNoRows, "deleteScrip"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for deleteScrip no rows affected, got %v", res)
	}

	// 10. Thematic additional branches
	_ = a.DB.Table(tblThematicBasketMaster).Create(M{
		"id":              601,
		"basket_name":     "Details Thematic",
		"active_status":   1,
		"created_on":      now,
		"expiry_date":     now.Add(24 * time.Hour),
		"total_invst_amt": 5000,
	})
	_ = a.DB.Table("tbl_sector_reports").Create(M{
		"id":            802,
		"referance_id":  601,
		"name":          "Ref Report",
		"active_status": 1,
	})
	reqThemDetWithDocs := httptest.NewRequest("GET", "/thematic/thematicDetails/601", nil)
	reqThemDetWithDocs.SetPathValue("id", "601")
	if res, err := a.thematic(reqThemDetWithDocs, "thematicDetails"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for thematicDetails with docs, got res=%v, err=%v", res, err)
	}

	reqThemRev404 := httptest.NewRequest("POST", "/thematic/review", strings.NewReader(`{"basketId":999997}`))
	reqThemRev404 = reqThemRev404.WithContext(context.WithValue(reqThemRev404.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.thematic(reqThemRev404, "review"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for review with missing basket, got %v", res)
	}

	reqThemRevEmptyScrips := httptest.NewRequest("POST", "/thematic/review", strings.NewReader(`{"basketId":601,"scrips":[]}`))
	reqThemRevEmptyScrips = reqThemRevEmptyScrips.WithContext(context.WithValue(reqThemRevEmptyScrips.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.thematic(reqThemRevEmptyScrips, "review"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for review with empty scrips, got %v", res)
	}

	reqThemRevNoDBScrips := httptest.NewRequest("POST", "/thematic/review", strings.NewReader(`{"basketId":601,"scrips":[{"token":"2188","ltp":100}]}`))
	reqThemRevNoDBScrips = reqThemRevNoDBScrips.WithContext(context.WithValue(reqThemRevNoDBScrips.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.thematic(reqThemRevNoDBScrips, "review"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for review with no DB scrips, got %v", res)
	}

	_ = a.DB.Table(tblThematicBasketScrips).Create(M{"basket_id": 601, "token": "2188", "qty": 10, "weightage": 50})
	reqThemRevTokenMismatch := httptest.NewRequest("POST", "/thematic/review", strings.NewReader(`{"basketId":601,"scrips":[{"token":"999998","ltp":100}]}`))
	reqThemRevTokenMismatch = reqThemRevTokenMismatch.WithContext(context.WithValue(reqThemRevTokenMismatch.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.thematic(reqThemRevTokenMismatch, "review"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for review with mismatched token, got %v", res)
	}

	reqThemRevOk := httptest.NewRequest("POST", "/thematic/review", strings.NewReader(`{"basketId":601,"lotSize":1,"scrips":[{"token":"2188","ltp":100.5}]}`))
	reqThemRevOk = reqThemRevOk.WithContext(context.WithValue(reqThemRevOk.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, err := a.thematic(reqThemRevOk, "review"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for valid thematic review, got res=%v, err=%v", res, err)
	}

	reqRebalNoCount := httptest.NewRequest("POST", "/thematic/rebalance", strings.NewReader(`{"basketId":601}`))
	reqRebalNoCount = reqRebalNoCount.WithContext(context.WithValue(reqRebalNoCount.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.thematic(reqRebalNoCount, "rebalance"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for rebalance count 0, got %v", res)
	}

	// 11. Repository: ResearchData & OrderStatusByOrderNo
	_ = a.DB.Create(&model.OrderStatusFeedEntity{OrderNo: ptr("FEED-001"), OrderStatus: ptr("COMPLETE"), TradedPrice: ptr("150.0")})
	if feed, err := a.OrderStatusByOrderNo(ctx, "FEED-001"); err != nil || val(feed.OrderNo) != "FEED-001" {
		t.Fatalf("expected OrderStatusByOrderNo to succeed, got %v, err=%v", feed, err)
	}

	rReq := &model.ResearchCallRequest{
		Category:    ptr("EQUITY"),
		SubCategory: ptr("LARGE_CAP"),
		Tags:        []string{"GROWTH"},
		FromDate:    ptr("2026-01-01"),
		ToDate:      ptr("2026-12-31"),
	}
	_, _ = a.ResearchData(ctx, "USER1", rReq)

	rReqFromOnly := &model.ResearchCallRequest{
		FromDate: ptr("2026-01-01"),
	}
	_, _ = a.ResearchData(ctx, "USER1", rReqFromOnly)

	rReqToOnly := &model.ResearchCallRequest{
		ToDate: ptr("2026-12-31"),
	}
	_, _ = a.ResearchData(ctx, "USER1", rReqToOnly)

	_, _ = a.ResearchData(ctx, "USER1", nil)

	// 12. Direct admin basket creation with push notification
	admScrips := []model.BasketScripEntity{{Exchange: ptr("NSE"), Token: ptr("2188"), Qty: ptr("10")}}
	_ = a.createAdminBasket(ctx, "USER1", model.AdminBasketOrderReq{
		BasketName:       ptr("DirectAdmin"),
		ExpiryDate:       &expTs,
		PushNotification: 1,
		Title:            ptr("DirectTitle"),
		Message:          ptr("DirectMsg"),
	}, admScrips, false)

	// 13. Rebalance with data
	_ = a.DB.Table(tblThematicBasketMaster).Create(M{
		"id":                  602,
		"rebalance_available": 1,
		"active_status":       1,
	})
	_ = a.DB.Table("tbl_user_thematic_exec").Create(M{
		"basket_id":   602,
		"user_id":     "USER1",
		"is_executed": 1,
	})
	reqRebalWithCount := httptest.NewRequest("POST", "/thematic/rebalance", strings.NewReader(`{"basketId":602}`))
	reqRebalWithCount = reqRebalWithCount.WithContext(context.WithValue(reqRebalWithCount.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, _ := a.thematic(reqRebalWithCount, "rebalance"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for rebalance with no scrips, got %v", res)
	}

	_ = a.DB.Table("tbl_thematic_basket_rebalance_scrips").Create(M{
		"basket_id": 602,
		"token":     "2188",
		"exchange":  "NSE",
		"qty":       5,
		"version":   1,
	})
	reqRebalWithScrips := httptest.NewRequest("POST", "/thematic/rebalance", strings.NewReader(`{"basketId":602}`))
	reqRebalWithScrips = reqRebalWithScrips.WithContext(context.WithValue(reqRebalWithScrips.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, err := a.thematic(reqRebalWithScrips, "rebalance"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for rebalance with scrips, got res=%v, err=%v", res, err)
	}

	// 14. Thematic invest validation branches
	reqInvestNoAction := httptest.NewRequest("POST", "/thematic/investV1", strings.NewReader(fmt.Sprintf(`{"basketId":601,"lots":1,"scrips":[%s]}`, fullScrip)))
	reqInvestNoAction = reqInvestNoAction.WithContext(context.WithValue(reqInvestNoAction.Context(), identityKey{}, Identity{UserID: "USER1", Token: "tok"}))
	if res, _ := a.thematic(reqInvestNoAction, "investV1"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for investV1 missing action, got %v", res)
	}

	reqInvestBadLots := httptest.NewRequest("POST", "/thematic/investV1", strings.NewReader(fmt.Sprintf(`{"basketId":601,"basketAction":"BUY","lots":0,"scrips":[%s]}`, fullScrip)))
	reqInvestBadLots = reqInvestBadLots.WithContext(context.WithValue(reqInvestBadLots.Context(), identityKey{}, Identity{UserID: "USER1", Token: "tok"}))
	if res, _ := a.thematic(reqInvestBadLots, "investV1"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for investV1 lots 0, got %v", res)
	}

	reqInvestNoToken := httptest.NewRequest("POST", "/thematic/invest", strings.NewReader(fmt.Sprintf(`{"basketId":601,"source":"WEB","scrips":[%s]}`, fullScrip)))
	reqInvestNoToken = reqInvestNoToken.WithContext(context.WithValue(reqInvestNoToken.Context(), identityKey{}, Identity{UserID: "USER1", Token: ""}))
	if res, _ := a.thematic(reqInvestNoToken, "invest"); res.(httpResult).Status != 401 {
		t.Fatalf("expected 401 for invest missing token, got %v", res)
	}

	reqInvestBasket404 := httptest.NewRequest("POST", "/thematic/invest", strings.NewReader(fmt.Sprintf(`{"basketId":999999,"source":"WEB","scrips":[%s]}`, fullScrip)))
	reqInvestBasket404 = reqInvestBasket404.WithContext(context.WithValue(reqInvestBasket404.Context(), identityKey{}, Identity{UserID: "USER1", Token: "tok"}))
	if res, _ := a.thematic(reqInvestBasket404, "invest"); res.(Response).Status != "Not ok" {
		t.Fatalf("expected Not ok for invest unowned basket, got %v", res)
	}

	// 15. Basket rename, addScrip, updateScrip, deleteScrip success paths
	reqRenameOk := httptest.NewRequest("POST", "/basketorder/renameBasket", strings.NewReader(`{"basketId":884,"basketName":"RenamedBasket"}`))
	reqRenameOk = reqRenameOk.WithContext(context.WithValue(reqRenameOk.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, err := a.basket(reqRenameOk, "rename"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for valid rename, got res=%v, err=%v", res, err)
	}

	reqAddScripOk := httptest.NewRequest("POST", "/basketorder/addScripToBasket", strings.NewReader(fmt.Sprintf(`{"basketId":884,"scrips":%s}`, fullScrip)))
	reqAddScripOk = reqAddScripOk.WithContext(context.WithValue(reqAddScripOk.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, err := a.basket(reqAddScripOk, "addScrip"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for valid addScrip, got res=%v, err=%v", res, err)
	}

	var addedScrip model.BasketScripEntity
	_ = a.DB.Where("basket_id = ?", 884).First(&addedScrip)
	updScripJSON := fmt.Sprintf(`{"basketId":884,"scrips":{"id":%d,"exchange":"NSE","token":"2188","qty":"2","price":"105","product":"CNC","transType":"BUY","priceType":"LIMIT","orderType":"REGULAR","ret":"DAY","source":"WEB","tradingSymbol":"GENCON-EQ"}}`, addedScrip.Id)
	reqUpdScripOk := httptest.NewRequest("POST", "/basketorder/updateBasketOrderScrip", strings.NewReader(updScripJSON))
	reqUpdScripOk = reqUpdScripOk.WithContext(context.WithValue(reqUpdScripOk.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, err := a.basket(reqUpdScripOk, "updateScrip"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for valid updateScrip, got res=%v, err=%v", res, err)
	}

	reqDelScripOk := httptest.NewRequest("POST", "/basketorder/deleteBasketOrderScrips", strings.NewReader(fmt.Sprintf(`{"basketId":884,"scripsId":[%d]}`, addedScrip.Id)))
	reqDelScripOk = reqDelScripOk.WithContext(context.WithValue(reqDelScripOk.Context(), identityKey{}, Identity{UserID: "USER1"}))
	if res, err := a.basket(reqDelScripOk, "deleteScrip"); err != nil || res.(Response).Status != "Ok" {
		t.Fatalf("expected Ok for valid deleteScrip, got res=%v, err=%v", res, err)
	}

	// 16. ClickHouse logging in logging.go
	chSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer chSrv.Close()
	a.Config.Logging.ClickHouseURL = chSrv.URL
	a.Config.Logging.ClickHouseDatabase = "default"
	a.Config.Logging.ClickHouseAccessTable = "access_table"
	a.Config.Logging.ClickHouseRestTable = "rest_table"
	a.Config.Logging.ClickHouseUser = "default"
	a.Config.Logging.ClickHousePassword = "pwd"

	a.writeLog(ctx, "tbl_access_log", "access_table", M{"user_id": "USER1", "req_body": "test-req", "in_time": now}, 200)
	a.restLog(ctx, "http://example.com/api?q=1", "body", "resp", now)

	// 17. db.go OpenDatabase and Migrate
	_ = Migrate(a.DB)
	_, _ = OpenDatabase(config.Config{
		Database: config.Database{
			Driver: "unsupported",
		},
	})
	dbMem, err := OpenDatabase(config.Config{
		Database: config.Database{
			Driver:      "sqlite",
			DSN:         ":memory:",
			AutoMigrate: true,
		},
		Logging: config.Logging{
			SQL: true,
		},
	})
	if err == nil && dbMem != nil {
		sqlDb, _ := dbMem.DB()
		if sqlDb != nil {
			_ = sqlDb.Close()
		}
	}

	// 18. auth.go HMAC JWT verify
	authHmac := &Authenticator{
		config: config.Auth{
			Enabled:    true,
			Mode:       "jwt",
			HMACSecret: "secret123",
			Issuer:     "test-issuer",
			Audience:   "test-audience",
			UserClaim:  "sub",
		},
		http: a.HTTP,
	}
	tokenHmac := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":   "test-issuer",
		"aud":   "test-audience",
		"sub":   "USER_HMAC",
		"ucc":   "UCC123",
		"name":  "User Name",
		"scope": "read",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	})
	signedToken, _ := tokenHmac.SignedString([]byte("secret123"))

	reqHmac := httptest.NewRequest("GET", "/test", nil)
	reqHmac.Header.Set("Authorization", "Bearer "+signedToken)
	idHmac, err := authHmac.verify(reqHmac)
	if err != nil || idHmac.UserID != "USER_HMAC" {
		t.Fatalf("expected valid HMAC identity, got %v, err=%v", idHmac, err)
	}

	reqNoBearer := httptest.NewRequest("GET", "/test", nil)
	_, _ = authHmac.verify(reqNoBearer)
}
