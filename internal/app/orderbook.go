package app

import (
	"basket/internal/model"
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// added for sonarqube
const orderTypeAMO = "AMO"

/*
func (a *App) refreshOrderBook(ctx context.Context, executionID int64, user string, results []M) error {
	if a.Config.Upstream.OrderBookURL == "" {
		return nil
	}
	wanted := map[string]bool{}
	for _, r := range results {
		if nonempty(r["orderNo"]) {
			wanted[normalizeOrder(str(r["orderNo"]))] = true
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	var customer M
	if e := a.Cache.Get(ctx, a.Config.Cache.Maps["customers"], user, &customer); e != nil {
		return e
	}
	aliases := a.Config.Business.ProductAliases
	if settings, ok := customer["userSettingDto"].(map[string]any); ok {
		if s, ok := settings["s_prdt_ali"].(string); ok && strings.TrimSpace(s) == "" {
			aliases = s
		}
	}
	u, e := url.Parse(a.Config.Upstream.OrderBookURL)
	if e != nil {
		return e
	}
	q := u.Query()
	q.Set("jsessionid", "."+str(customer["tomcatcount"]))
	q.Set("jData", jsonString(M{"uid": user, "s_prdt_ali": aliases}))
	q.Set("jKey", str(customer["stringPkey4"]))
	u.RawQuery = q.Encode()
	var raw []M
	if _, e = a.call(ctx, u.String(), "", "application/json", "", &raw); e != nil {
		return e
	}
	for _, entry := range raw {
		ob := a.bindOrderBook(ctx, entry)
		key := normalizeOrder(str(ob["orderNo"]))
		if !wanted[key] {
			continue
		}
		price := str(ob["avgTradePrice"])
		if price == "" || dec(price).IsZero() {
			price = str(ob["price"])
		}
		if price == "" {
			price = "0.0"
		}
		if e = a.updateExecutionDetailRow(ctx, executionID, key, ob, price); e != nil {
			return e
		}
	}
	return nil
}
*/

// added for sonarqube
func (a *App) fetchUpstreamOrderBook(ctx context.Context, user string) ([]M, error) {
	var customer M
	if e := a.Cache.Get(ctx, a.Config.Cache.Maps["customers"], user, &customer); e != nil {
		return nil, e
	}
	aliases := a.Config.Business.ProductAliases
	if settings, ok := customer["userSettingDto"].(map[string]any); ok {
		if s, ok := settings["s_prdt_ali"].(string); ok && strings.TrimSpace(s) == "" {
			aliases = s
		}
	}
	u, e := url.Parse(a.Config.Upstream.OrderBookURL)
	if e != nil {
		return nil, e
	}
	q := u.Query()
	q.Set("jsessionid", "."+str(customer["tomcatcount"]))
	q.Set("jData", jsonString(M{"uid": user, "s_prdt_ali": aliases}))
	q.Set("jKey", str(customer["stringPkey4"]))
	u.RawQuery = q.Encode()
	var raw []M
	if _, e = a.call(ctx, u.String(), "", "application/json", "", &raw); e != nil {
		return nil, e
	}
	return raw, nil
}

// added for sonarqube
func resolveOrderBookPrice(ob M) string {
	price := str(ob["avgTradePrice"])
	if price == "" || dec(price).IsZero() {
		price = str(ob["price"])
	}
	if price == "" {
		price = "0.0"
	}
	return price
}

// added for sonarqube
func (a *App) refreshOrderBook(ctx context.Context, executionID int64, user string, results []M) error {
	if a.Config.Upstream.OrderBookURL == "" {
		return nil
	}
	wanted := map[string]bool{}
	for _, r := range results {
		if nonempty(r["orderNo"]) {
			wanted[normalizeOrder(str(r["orderNo"]))] = true
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	raw, e := a.fetchUpstreamOrderBook(ctx, user)
	if e != nil {
		return e
	}
	for _, entry := range raw {
		ob := a.bindOrderBook(ctx, entry)
		key := normalizeOrder(str(ob["orderNo"]))
		if !wanted[key] {
			continue
		}
		price := resolveOrderBookPrice(ob)
		if e = a.updateExecutionDetailRow(ctx, executionID, key, ob, price); e != nil {
			return e
		}
	}
	return nil
}

// added for sonarqube
func (a *App) updateExecutionDetailRow(ctx context.Context, executionID int64, key string, ob M, price string) error {
	var ds []model.ExecutionDetail
	if e := a.DB.WithContext(ctx).Where("execution_id = ?", strconv.FormatInt(executionID, 10)).Find(&ds).Error; e != nil {
		return e
	}
	for _, d := range ds {
		if normalizeOrder(val(d.OrderNo)) != key {
			continue
		}
		if e := a.DB.WithContext(ctx).Model(&model.ExecutionDetail{}).Where("id = ?", d.Id).Updates(M{"order_status": ob["orderStatus"], "exch": ob["exchange"], "trading_symbol": ob["tradingSymbol"], "token": ob["token"], "trans_type": ob["transType"], "executed_qty": integer(ob["qty"]), "executed_price": price, "updated_by": "SYSTEM", "updated_on": a.Now()}).Error; e != nil {
			return e
		}
	}
	return nil
}
func normalizeOrder(s string) string {
	return strings.ToUpper(strings.TrimRight(strings.TrimSpace(strings.ReplaceAll(s, ";", "")), ":"))
}
func (a *App) bindOrderBook(ctx context.Context, r M) M {
	m := project(r, map[string]string{"actId": "accountId", "avgTradePrice": "Avgprc", "disclosedQty": "Dscqty", "exchOrderId": "ExchOrdID", "exchUpdateTime": "ExchConfrmtime", "exchange": "Exchange", "mktProtection": "marketprotectionpercentage", "multiplier": "multiplier", "orderNo": "Nstordno", "priceType": "Prctype", "product": "Pcode", "rejectedReason": "RejReason", "ret": "Validity", "tickSize": "ticksize", "token": "token", "tradingSymbol": "Trsym", "triggerPrice": "Trgprc", "userId": "user"})
	qty := integer(r["Qty"])
	c, e := a.contract(ctx, str(r["Exchange"]), str(r["token"]))
	if e == nil {
		m["lotSize"] = c.LotSize
		m["companyName"] = c.CompanyName
		if strings.EqualFold(str(r["Exchange"]), "MCX") && integer(c.LotSize) > 0 {
			qty *= integer(c.LotSize)
		}
	}
	m["qty"] = strconv.Itoa(qty)
	m["formattedInsName"] = r["Scripname"]
	if nonempty(c.FormattedInsName) {
		m["formattedInsName"] = strings.ToUpper(val(c.FormattedInsName))
	}
	status := str(r["Status"])
	switch strings.ToUpper(status) {
	case "TRIGGER PENDING":
		m["orderStatus"] = "pending"
	case "CANCELLED AFTER MARKET ORDER":
		m["orderStatus"] = "Cancelled AMO"
	default:
		m["orderStatus"] = status
	}
	if t, e := time.ParseInLocation("02/01/2006 15:04:05", str(r["OrderedTime"]), a.location); e == nil {
		m["orderTime"] = t.Format("2006-01-02 03:04:05")
	}
	//added for sonarqube
	// if strings.EqualFold(str(r["ordergenerationtype"]), "AMO") {
	// 	m["orderType"] = "AMO"
	if strings.EqualFold(str(r["ordergenerationtype"]), orderTypeAMO) {
		m["orderType"] = orderTypeAMO
	} else {
		m["orderType"] = map[string]string{"BO": "Bracket", "CO": "Cover", "MTF": "MTF", "CNC": "DELIVERY", "MIS": "INTRADAY", "NRML": "Normal"}[strings.ToUpper(str(r["Pcode"]))]
	}
	m["price"] = r["Prc"]
	if dec(r["Prc"]).IsZero() || strings.EqualFold(status, "COMPLETE") {
		m["price"] = r["Avgprc"]
	}
	side := strings.ToUpper(str(r["Trantype"]))
	if side == "B" {
		side = "BUY"
	} else if side == "S" {
		side = "SELL"
	}
	m["transType"] = side
	for _, key := range []string{"rprice", "rqty", "stopLoss", "target", "ltp", "trailingPrice"} {
		m[key] = "0"
	}
	return m
}

// UpdateFromFeed is the port of ThematicDao.updateThematicExecDetailsFromFeed.
func (a *App) UpdateFromFeed(ctx context.Context, executionID int64, orderNo string, f model.OrderStatusFeedEntity) error {
	return a.DB.WithContext(ctx).Model(&model.ExecutionDetail{}).Where("execution_id = ? AND order_no = ?", strconv.FormatInt(executionID, 10), orderNo).Updates(M{"order_status": f.OrderStatus, "exch": f.Exch, "executed_qty": integer(f.Qty), "executed_price": f.TradedPrice, "rejected_reason": f.Reason, "updated_by": "SYSTEM", "updated_on": a.Now()}).Error
}
