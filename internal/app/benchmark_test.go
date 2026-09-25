package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupBenchmarkApp(b *testing.B) (*App, int64, int64, int64) {
	b.Helper()
	c, err := config.Load("../../testdata/config.local.yaml")
	if err != nil {
		b.Fatal(err)
	}
	c.Auth.Enabled = false
	c.Auth.DevUser = "BENCH_USER"
	c.Auth.AdminToken = "bench-admin"
	c.Logging.Access = false
	c.Logging.Database = false
	c.Modules.Scheduler = false
	c.Modules.Notifications = false

	db, err := gorm.Open(sqlite.Open("file:"+strings.ReplaceAll(b.Name(), "/", "_")+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		b.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err = Migrate(db); err != nil {
		b.Fatal(err)
	}

	cache := &MemoryCache{Values: map[string]map[string]json.RawMessage{}}
	a, err := New(context.Background(), c, db, cache)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		a.Close(context.Background())
		sqlDB.Close()
	})

	now := time.Date(2026, 9, 10, 10, 0, 0, 0, a.location)
	a.Now = func() time.Time { return now }

	// Seed contracts
	for i := 1; i <= 20; i++ {
		tokenStr := fmt.Sprintf("%d", 2000+i)
		contract := model.ContractMasterModel{
			Exch:             ptr("NSE"),
			Token:            ptr(tokenStr),
			TradingSymbol:    ptr(fmt.Sprintf("STOCK%d-EQ", i)),
			FormattedInsName: ptr(fmt.Sprintf("STOCK%d-EQ", i)),
			LotSize:          ptr("1"),
			Pdc:              ptr("100.00"),
		}
		cache.Put(context.Background(), c.Cache.Maps["contracts"], "NSE_"+tokenStr, contract)
	}
	cache.Put(context.Background(), c.Cache.Maps["sessions"], "BENCH_USER_REST_SESSION", "session")

	// Seed Basket Order data
	var basketID int64
	for i := 1; i <= 10; i++ {
		bn := model.BasketNameEntity{
			UserId:       ptr("BENCH_USER"),
			BasketName:   ptr(fmt.Sprintf("BenchBasket_%d", i)),
			IsExecuted:   ptr("0"),
			ActiveStatus: 1,
			CreatedBy:    ptr("BENCH_USER"),
		}
		db.Create(&bn)
		if i == 1 {
			basketID = bn.BasketId
		}
		for j := 1; j <= 5; j++ {
			tokenStr := fmt.Sprintf("%d", 2000+j)
			scrip := model.BasketScripEntity{
				BasketId:         bn.BasketId,
				Exchange:         ptr("NSE"),
				Token:            ptr(tokenStr),
				Qty:              ptr("10"),
				Price:            ptr("100.00"),
				Product:          ptr("CNC"),
				TransType:        ptr("BUY"),
				PriceType:        ptr("MKT"),
				OrderType:        ptr("Regular"),
				Ret:              ptr("DAY"),
				TradingSymbol:    ptr(fmt.Sprintf("STOCK%d-EQ", j)),
				FormattedInsName: ptr(fmt.Sprintf("STOCK%d-EQ", j)),
				ActiveStatus:     1,
			}
			db.Create(&scrip)
		}
	}

	// Seed Thematic Baskets
	var thematicID int64
	for i := 1; i <= 5; i++ {
		tb := model.ThematicMaster{
			Category:           ptr("Equity"),
			SubCategory:        ptr("Growth"),
			BasketName:         ptr(fmt.Sprintf("Thematic_%d", i)),
			ShortDescription:   ptr("Short desc"),
			LongDescription:    ptr("Long desc"),
			Status:             ptr("Open"),
			ActionType:         ptr("SEND_NOW"),
			TotalInvstAmt:      1000.0,
			RebalanceAvailable: 1,
			ActiveStatus:       1,
		}
		db.Create(&tb)
		if i == 1 {
			thematicID = tb.Id
		}
		db.Create(&model.ReasearchCallUsers{UserId: ptr("BENCH_USER"), ThematicBasketId: int(tb.Id)})
		for j := 1; j <= 5; j++ {
			tokenStr := fmt.Sprintf("%d", 2000+j)
			ts := model.ThematicScrip{
				BasketId:         tb.Id,
				Exchange:         ptr("NSE"),
				Token:            ptr(tokenStr),
				Qty:              ptr("5"),
				Price:            ptr("100.00"),
				Weightage:        ptr("20.00"),
				TradingSymbol:    ptr(fmt.Sprintf("STOCK%d-EQ", j)),
				FormattedInsName: ptr(fmt.Sprintf("STOCK%d-EQ", j)),
				OrderType:        ptr("Regular"),
				PriceType:        ptr("MKT"),
				TransType:        ptr("BUY"),
				Version:          1,
			}
			db.Create(&ts)
		}
		// Seed executions for holdings
		nowTS := model.Timestamp(now)
		for j := 1; j <= 5; j++ {
			tokenStr := fmt.Sprintf("%d", 2000+j)
			d := model.ExecutionDetail{
				ActiveStatus: 1,
				ThematicExecDetails: model.ThematicExecDetails{
					ExecutionId:   ptr(fmt.Sprintf("%d", i)),
					UserId:        ptr("BENCH_USER"),
					ResearchId:    ptr(int(tb.Id)),
					ResearchType:  ptr(2),
					BasketAction:  ptr("BUY"),
					TransType:     ptr("BUY"),
					OrderStatus:   ptr("EXECUTED"),
					ExecutedQty:   ptr(10),
					RecommQty:     ptr(5),
					ExecutedPrice: ptr("100.00"),
					Exch:          ptr("NSE"),
					Token:         ptr(tokenStr),
					TradingSymbol: ptr(fmt.Sprintf("STOCK%d-EQ", j)),
					Version:       ptr(1),
					LotSize:       ptr(2),
					ExecutedOn:    &nowTS,
				},
			}
			db.Create(&d)
		}
		db.Create(&model.ThematicExeMasterEntity{
			UserId:       ptr("BENCH_USER"),
			BasketId:     tb.Id,
			BasketAction: ptr("BUY"),
			ActiveStatus: 1,
		})
	}

	// Seed Research Calls
	var researchID int64
	for i := 1; i <= 5; i++ {
		rc := model.ResearchMaster{
			ResearchcallOrderEntity: model.ResearchcallOrderEntity{
				Category:     ptr("Equity"),
				SubCategory:  ptr("Positional"),
				Status:       ptr("open"),
				CreatedOn:    ptr(model.Timestamp(now)),
				ActiveStatus: 1,
			},
			AnalystName: ptr(fmt.Sprintf("Analyst_%d", i)),
		}
		db.Create(&rc)
		if i == 1 {
			researchID = rc.Id
		}
		db.Create(&model.ReasearchCallUsers{UserId: ptr("BENCH_USER"), ResearchCallId: int(rc.Id)})
		for j := 1; j <= 4; j++ {
			tokenStr := fmt.Sprintf("%d", 2000+j)
			rs := model.ResearchcallScripEntity{
				ResearchcallId:   int(rc.Id),
				Token:            ptr(tokenStr),
				Exchange:         ptr("NSE"),
				Qty:              ptr("10"),
				Expiry:           ptr("2026-09-30"),
				Retention:        ptr("DAY"),
				TradingSymbol:    ptr(fmt.Sprintf("STOCK%d-EQ", j)),
				FormattedInsName: ptr(fmt.Sprintf("STOCK%d-EQ", j)),
			}
			db.Create(&rs)
		}
	}

	return a, basketID, thematicID, researchID
}

func recordLatency(b *testing.B, name string, latencies []time.Duration) {
	if len(latencies) == 0 {
		return
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	avg := total / time.Duration(len(latencies))
	p50 := latencies[int(float64(len(latencies))*0.50)]
	p95 := latencies[int(float64(len(latencies))*0.95)]
	p99 := latencies[int(float64(len(latencies))*0.99)]
	b.Logf("[%s] count=%d avg=%v p50=%v p95=%v p99=%v", name, len(latencies), avg, p50, p95, p99)
}

func BenchmarkGetBasketReports(b *testing.B) {
	a, _, _, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/basketorder/get", nil)
		req.Header.Set("Authorization", "Bearer test")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "GET /basketorder/get", latencies)
}

func BenchmarkGetBasketScrips(b *testing.B) {
	a, basketID, _, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	url := fmt.Sprintf("/basketorder/get/scrips/%d", basketID)
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer test")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "GET /basketorder/get/scrips/{id}", latencies)
}

func BenchmarkSpanMargin(b *testing.B) {
	a, _, _, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	body := `[{"exchange":"NSE","token":"2001","qty":"2","price":"100.00","transType":"BUY"},{"exchange":"NSE","token":"2002","qty":"5","price":"50.00","transType":"BUY"}]`
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/basketorder/spanmargin", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer test")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "POST /basketorder/spanmargin", latencies)
}

func BenchmarkThematicGetAll(b *testing.B) {
	a, _, _, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/thematic/basket/getall", nil)
		req.Header.Set("Authorization", "Bearer test")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "GET /thematic/basket/getall", latencies)
}

func BenchmarkThematicDetails(b *testing.B) {
	a, _, thematicID, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	url := fmt.Sprintf("/thematic/basket/get/%d", thematicID)
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer test")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "GET /thematic/basket/get/{id}", latencies)
}

func BenchmarkThematicReview(b *testing.B) {
	a, _, thematicID, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	body := fmt.Sprintf(`{"basketId":%d,"lotSize":2,"scrips":[{"token":"2001","ltp":100.50},{"token":"2002","ltp":50.25}]}`, thematicID)
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/thematic/basket/review", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer test")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "POST /thematic/basket/review", latencies)
}

func BenchmarkThematicHoldingsV1(b *testing.B) {
	a, _, _, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/thematic/holdings/get/V1", nil)
		req.Header.Set("Authorization", "Bearer test")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "GET /thematic/holdings/get/V1", latencies)
}

func BenchmarkResearchGetResearchCall(b *testing.B) {
	a, _, _, _ := setupBenchmarkApp(b)
	handler := a.Handler()
	body := `{}`
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/research/getResearchCall", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer test")
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		t0 := time.Now()
		handler.ServeHTTP(w, req)
		latencies = append(latencies, time.Since(t0))
		if w.Code != 200 {
			b.Fatalf("status %d", w.Code)
		}
	}
	recordLatency(b, "POST /research/getResearchCall", latencies)
}
