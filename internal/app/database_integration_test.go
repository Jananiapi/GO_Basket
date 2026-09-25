package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// These DSNs must point to disposable test databases. No test uses config.yaml's DSN.
func TestDatabaseIntegration(t *testing.T) {
	for _, driver := range []string{"mysql", "postgres", "mssql"} {
		t.Run(driver, func(t *testing.T) {
			env := map[string]string{"mysql": "BASKET_TEST_MYSQL_DSN", "postgres": "BASKET_TEST_POSTGRES_DSN", "mssql": "BASKET_TEST_MSSQL_DSN"}[driver]
			dsn := os.Getenv(env)
			if dsn == "" {
				t.Skip("set " + env + " to a disposable database")
			}
			c, _ := config.Load("../../config.yaml")
			c.Database.Driver = driver
			c.Database.DSN = dsn
			c.Database.AutoMigrate = true
			c.Auth.Enabled = false
			c.Auth.DevUser = "DIALECT_TEST"
			c.Logging.Access = false
			c.Modules.Scheduler = false
			db, e := OpenDatabase(c)
			if e != nil {
				t.Fatal(e)
			}
			sql, _ := db.DB()
			defer sql.Close()
			cache := &MemoryCache{}
			a, e := New(context.Background(), c, db, cache)
			if e != nil {
				t.Fatal(e)
			}
			defer a.Close(context.Background())
			name := t.Name()
			db.Where("user_id = ?", c.Auth.DevUser).Delete(&model.BasketNameEntity{})
			id := create(t, a, name)
			cache.Put(context.Background(), c.Cache.Maps["contracts"], "NSE_2188", model.ContractMasterModel{Exch: ptr("NSE"), Token: ptr("2188"), TradingSymbol: ptr("GENCON-EQ"), FormattedInsName: ptr("GENCON-EQ"), LotSize: ptr("1")})
			_, added := request(t, a, "POST", "/basketorder/add/scrips", jsonString(M{"basketId": id, "scrips": json.RawMessage(scripJSON)}))
			ok(t, added)
			_, got := request(t, a, "GET", "/basketorder/get/scrips/"+str(id), "")
			ok(t, got)
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonWrite(w, 200, []Response{success([]M{{"orderNo": "dialect-1", "requestTime": "now"}})})
			}))
			defer srv.Close()
			a.Config.Upstream.OrderURL = srv.URL
			_, executed := request(t, a, "POST", "/basketorder/execute", jsonString(M{"basketId": id, "scrips": []json.RawMessage{json.RawMessage(scripJSON)}}))
			ok(t, executed.([]any)[0])
			master := model.ThematicMaster{BasketName: ptr("integration"), Status: ptr("Open"), ActionType: ptr("SEND_NOW")}
			if e = db.Create(&master).Error; e != nil {
				t.Fatal(e)
			}
			ts := model.ThematicScrip{BasketId: master.Id, Token: ptr("2188"), Exchange: ptr("NSE"), TradingSymbol: ptr("GENCON-EQ"), Qty: ptr("1"), Weightage: ptr("100"), Price: ptr("43.04"), Version: 1}
			if e = db.Create(&ts).Error; e != nil {
				t.Fatal(e)
			}
			if e = db.Create(&model.ReasearchCallUsers{UserId: ptr(c.Auth.DevUser), ThematicBasketId: int(master.Id)}).Error; e != nil {
				t.Fatal(e)
			}
			_, details := request(t, a, "GET", "/thematic/basket/get/"+str(master.Id), "")
			ok(t, details)
			_, preview := request(t, a, "POST", "/thematic/basket/review", jsonString(M{"basketId": master.Id, "lotSize": 1, "scrips": []M{{"token": "2188", "ltp": 43.04}}}))
			ok(t, preview)
			var sr M
			json.Unmarshal([]byte(scripJSON), &sr)
			sr["version"] = 1
			_, invested := request(t, a, "POST", "/thematic/basket/v1/invest", jsonString(M{"basketId": master.Id, "lots": 1, "basketAction": "BUY", "source": "WEB", "scrips": []M{sr}}))
			ok(t, invested.([]any)[0])
			_, hold := request(t, a, "GET", "/thematic/holdings/get/V1", "")
			ok(t, hold)
			rm := model.ResearchMaster{ResearchcallOrderEntity: model.ResearchcallOrderEntity{Category: ptr("EQ"), SubCategory: ptr("Daily"), Status: ptr("open"), ActiveStatus: 1}, AnalystName: ptr("Tester")}
			if e = db.Create(&rm).Error; e != nil {
				t.Fatal(e)
			}
			if e = db.Create(&model.ReasearchCallUsers{UserId: ptr(c.Auth.DevUser), ResearchCallId: int(rm.Id)}).Error; e != nil {
				t.Fatal(e)
			}
			_, research := request(t, a, "POST", "/research/getResearchCall", `{}`)
			ok(t, research)

			_, out := request(t, a, "GET", "/basketorder/get", "")
			ok(t, out)
			if e = db.Create(&model.ThematicMaster{BasketName: ptr("dialect"), ActiveStatus: 1, Status: ptr("Open")}).Error; e != nil {
				t.Fatal(e)
			}
			_, out = request(t, a, "GET", "/research/getUniqStatus", "")
			if out == nil {
				t.Fatal("missing response")
			}
			_, out = request(t, a, "DELETE", "/basketorder/delete/"+str(id), "")
			ok(t, out)
		})
	}
}
