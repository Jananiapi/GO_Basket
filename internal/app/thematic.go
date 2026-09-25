package app

import (
	"basket/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// added for sonarqube
const (
	tblThematicBasketScrips = "tbl_thematic_basket_scrips"
	tblThematicBasketMaster = "tbl_thematic_basket_master"
	msgFailed               = "Failed"
	idWhere                 = "id = ?"
)

func latestScrips(db *gorm.DB, table string, id any) ([]M, error) {
	sub := db.Table(table).Select("MAX(version)").Where(basketIDWhere, id)
	return rows(db.Table(table).Where(basketIDWhere+" AND version = (?)", id, sub))
}

func latestScripsBatch(db *gorm.DB, table string, basketIDs []any) (map[string][]M, error) {
	if len(basketIDs) == 0 {
		return map[string][]M{}, nil
	}
	allScrips, e := rows(db.Table(table).Where("basket_id IN ?", basketIDs).Order("id ASC"))
	if e != nil {
		return nil, e
	}
	maxVersion := map[string]int{}
	for _, s := range allScrips {
		bid := str(s["basket_id"])
		v := integer(s["version"])
		if v > maxVersion[bid] {
			maxVersion[bid] = v
		}
	}
	result := map[string][]M{}
	for _, s := range allScrips {
		bid := str(s["basket_id"])
		if integer(s["version"]) == maxVersion[bid] {
			result[bid] = append(result[bid], s)
		}
	}
	return result, nil
}
func thematicScrip(s M) M {
	return project(s, map[string]string{"exchange": "exchange", "token": "token", "qty": "qty", "holdQty": "hold_qty", "transType": "trans_type", "weightage": "weightage", "tradingSymbol": "trading_symbol", "formattedInsName": "formatted_ins_name", "price": "price", "pdc": "pdc", "marketCap": "exposure", "version": "version", "orderType": "order_type", "priceType": "price_type"})
}
/*
func (a *App) thematic(r *http.Request, action string) (any, error) {
	ctx := r.Context()
	db := a.DB.WithContext(ctx)
	idn := identity(r)
	switch action {
	case "invest", "investV1":
		return a.invest(r, action == "investV1")
	case "legacyHoldings":
		return a.legacyHoldings(ctx, idn.UserID)
	case "thematicAll":
		var ids []int64
		if e := db.Table("tbl_researchcall_usermapping").Where("user_id IN ? AND thematic_basket_id IS NOT NULL", []string{idn.UserID, "ALL"}).Pluck("thematic_basket_id", &ids).Error; e != nil {
			return nil, e
		}
		if len(ids) == 0 {
			return failed("No data found for this user"), nil
		}
		//added for sonarqube
		// bs, e := rows(db.Table("tbl_thematic_basket_master").Where("id IN ? AND active_status = ? AND action_type = ? AND LOWER(status) = ?", ids, 1, "SEND_NOW", "open").Order("created_on DESC"))
		bs, e := rows(db.Table(tblThematicBasketMaster).Where("id IN ? AND active_status = ? AND action_type = ? AND LOWER(status) = ?", ids, 1, "SEND_NOW", "open").Order("created_on DESC"))
		if e != nil {
			return nil, e
		}
		out := []M{}
		for _, b := range bs {
			m := project(b, map[string]string{"category": "category", "subCategory": "sub_category", "basketId": "id", "basketName": "basket_name", "shortDescription": "short_description", "longDescription": "long_description", "tag": "tag", "totalInvstAmt": "total_invst_amt", "curReturn": "current_returns", "minInvestmentAmount": "min_invst_amnt"})
			m["basketId"] = str(b["id"])
			m["createdOn"] = dateOnly(b["created_on"])
			m["expiryDate"] = dateOnly(b["expiry_date"])
			//added for sonarqube
			// ss, e := latestScrips(db, "tbl_thematic_basket_scrips", b["id"])
			ss, e := latestScrips(db, tblThematicBasketScrips, b["id"])
			if e != nil {
				return nil, e
			}
			outS := []M{}
			for _, s := range ss {
				v := project(s, map[string]string{"exchange": "exchange", "token": "token", "price": "price", "qty": "qty"})
				c, e := a.contract(ctx, str(s["exchange"]), str(s["token"]))
				if e == nil {
					v["pdc"] = c.Pdc
				} else {
					v["pdc"] = nil
				}
				outS = append(outS, v)
			}
			m["scrips"] = outS
			out = append(out, m)
		}
		return success(out), nil
	case "thematicDetails":
		id := r.PathValue("id")
		//added for sonarqube
		// b, e := first(db.Table("tbl_thematic_basket_master").Where("id = ?", id))
		b, e := first(db.Table(tblThematicBasketMaster).Where(idWhere, id))
		if notFound(e) {
			return failed("Basket not found for ID: " + id), nil
		}
		if e != nil {
			return nil, e
		}
		m := project(b, map[string]string{"basketName": "basket_name", "shortDescription": "short_description", "longDescription": "long_description", "tag": "tag", "tagDescription": "tag_desc", "risk": "risk_level", "benchmark": "benchmark", "investmentDuration": "investment_duration", "reviewFrequency": "review_frequency", "minInvstAmt": "total_invst_amt", "curReturn": "current_returns", "methodology": "methodology", "overallWeightage": "overall_weightage", "launchDate": "created_on", "rationale": "rationale", "minInvestmentAmount": "min_invst_amnt", "riskDescription": "risk_description", "type": "type", "analyst": "analyst_name", "category": "category", "subcategory": "sub_category", "expectedReturn": "expected_return"})
		m["exposure"] = nil
		m["minInvstAmt"] = number(b["total_invst_amt"])
		m["launchDate"] = str(b["created_on"])
		m["createdOn"] = dateOnly(b["created_on"])
		m["expiryDate"] = dateOnly(b["expiry_date"])
		//added for sonarqube
		// ss, e := latestScrips(db, "tbl_thematic_basket_scrips", id)
		ss, e := latestScrips(db, tblThematicBasketScrips, id)
		if e != nil {
			return nil, e
		}
		out := []M{}
		for _, s := range ss {
			out = append(out, thematicScrip(s))
		}
		m["scrips"] = out
		m["documents"] = nil
		reports, e := rows(db.Table("tbl_sector_reports").Where("referance_id = ? AND active_status = ?", id, 1))
		if e != nil {
			return nil, e
		}
		if len(reports) > 0 {
			docs := []M{}
			for _, d := range reports {
				docs = append(docs, project(d, map[string]string{"id": "id", "name": "name", "description": "description", "attachment": "attachment", "type": "type"}))
			}
			m["documents"] = docs
		}
		return success(m), nil
	}
	var req *model.ExecuteBasketOrderReq
	if decode(r, &req) != nil || req == nil {
		//added for sonarqube
		// return failed("Invalid Parameter"), nil
		return failed(invalidParameter), nil
	}
	switch action {
	case "view":
		e := db.Create(&model.ThematicExecution{UserId: idn.UserID, BasketId: int64(req.BasketId), IsViewed: req.IsViewed, CreatedBy: idn.UserID}).Error
		return message("Basket updated successfully"), e
	case "rebalance":
		var count int64
		//added for sonarqube
		// e := db.Table("tbl_thematic_basket_master AS b").Joins("JOIN tbl_user_thematic_exec e ON e.basket_id = b.id").Where("b.rebalance_available = ? AND b.id = ? AND e.user_id = ? AND e.is_executed = ?", 1, req.BasketId, idn.UserID, 1).Count(&count).Error
		e := db.Table(tblThematicBasketMaster+" AS b").Joins("JOIN tbl_user_thematic_exec e ON e.basket_id = b.id").Where("b.rebalance_available = ? AND b.id = ? AND e.user_id = ? AND e.is_executed = ?", 1, req.BasketId, idn.UserID, 1).Count(&count).Error
		if e != nil {
			return nil, e
		}
		if count == 0 {
			return failed("Rebalance not available for this user or basket."), nil
		}
		ss, e := latestScrips(db, "tbl_thematic_basket_rebalance_scrips", req.BasketId)
		if e != nil {
			return nil, e
		}
		if len(ss) == 0 {
			return failed("No rebalance data available."), nil
		}
		out := []M{}
		for _, s := range ss {
			m := thematicScrip(s)
			for _, k := range []string{"weightage", "marketCap", "version", "pdc"} {
				m[k] = nil
			}
			out = append(out, m)
		}
		return success(out), nil
	case "review":
		return a.thematicReview(db, req)
	}
	//added for sonarqube
	// return failed("Invalid Parameter"), nil
	return failed(invalidParameter), nil
}
*/

// added for sonarqube
func (a *App) handleThematicAll(ctx context.Context, db *gorm.DB, userID string) (any, error) {
	var ids []int64
	if e := db.Table("tbl_researchcall_usermapping").Where("user_id IN ? AND thematic_basket_id IS NOT NULL", []string{userID, "ALL"}).Pluck("thematic_basket_id", &ids).Error; e != nil {
		return nil, e
	}
	if len(ids) == 0 {
		return failed("No data found for this user"), nil
	}
	bs, e := rows(db.Table(tblThematicBasketMaster).Where("id IN ? AND active_status = ? AND action_type = ? AND LOWER(status) = ?", ids, 1, "SEND_NOW", "open").Order("created_on DESC"))
	if e != nil {
		return nil, e
	}
	bIDs := make([]any, len(bs))
	for i, b := range bs {
		bIDs[i] = b["id"]
	}
	scripsByBasket, e := latestScripsBatch(db, tblThematicBasketScrips, bIDs)
	if e != nil {
		return nil, e
	}
	out := make([]M, 0, len(bs))
	for _, b := range bs {
		m := project(b, map[string]string{"category": "category", "subCategory": "sub_category", "basketId": "id", "basketName": "basket_name", "shortDescription": "short_description", "longDescription": "long_description", "tag": "tag", "totalInvstAmt": "total_invst_amt", "curReturn": "current_returns", "minInvestmentAmount": "min_invst_amnt"})
		bidStr := str(b["id"])
		m["basketId"] = bidStr
		m["createdOn"] = dateOnly(b["created_on"])
		m["expiryDate"] = dateOnly(b["expiry_date"])
		ss := scripsByBasket[bidStr]
		outS := make([]M, 0, len(ss))
		for _, s := range ss {
			v := project(s, map[string]string{"exchange": "exchange", "token": "token", "price": "price", "qty": "qty"})
			c, e := a.contract(ctx, str(s["exchange"]), str(s["token"]))
			if e == nil {
				v["pdc"] = c.Pdc
			} else {
				v["pdc"] = nil
			}
			outS = append(outS, v)
		}
		m["scrips"] = outS
		out = append(out, m)
	}
	return success(out), nil
}

// added for sonarqube
func handleThematicDetails(db *gorm.DB, id string) (any, error) {
	b, e := first(db.Table(tblThematicBasketMaster).Where(idWhere, id))
	if notFound(e) {
		return failed("Basket not found for ID: " + id), nil
	}
	if e != nil {
		return nil, e
	}
	m := project(b, map[string]string{"basketName": "basket_name", "shortDescription": "short_description", "longDescription": "long_description", "tag": "tag", "tagDescription": "tag_desc", "risk": "risk_level", "benchmark": "benchmark", "investmentDuration": "investment_duration", "reviewFrequency": "review_frequency", "minInvstAmt": "total_invst_amt", "curReturn": "current_returns", "methodology": "methodology", "overallWeightage": "overall_weightage", "launchDate": "created_on", "rationale": "rationale", "minInvestmentAmount": "min_invst_amnt", "riskDescription": "risk_description", "type": "type", "analyst": "analyst_name", "category": "category", "subcategory": "sub_category", "expectedReturn": "expected_return"})
	m["exposure"] = nil
	m["minInvstAmt"] = number(b["total_invst_amt"])
	m["launchDate"] = str(b["created_on"])
	m["createdOn"] = dateOnly(b["created_on"])
	m["expiryDate"] = dateOnly(b["expiry_date"])
	ss, e := latestScrips(db, tblThematicBasketScrips, id)
	if e != nil {
		return nil, e
	}
	out := []M{}
	for _, s := range ss {
		out = append(out, thematicScrip(s))
	}
	m["scrips"] = out
	m["documents"] = nil
	reports, e := rows(db.Table("tbl_sector_reports").Where("referance_id = ? AND active_status = ?", id, 1))
	if e != nil {
		return nil, e
	}
	if len(reports) > 0 {
		docs := []M{}
		for _, d := range reports {
			docs = append(docs, project(d, map[string]string{"id": "id", "name": "name", "description": "description", "attachment": "attachment", "type": "type"}))
		}
		m["documents"] = docs
	}
	return success(m), nil
}

// added for sonarqube
func handleThematicRebalance(db *gorm.DB, req *model.ExecuteBasketOrderReq, userID string) (any, error) {
	var count int64
	e := db.Table(tblThematicBasketMaster+" AS b").Joins("JOIN tbl_user_thematic_exec e ON e.basket_id = b.id").Where("b.rebalance_available = ? AND b.id = ? AND e.user_id = ? AND e.is_executed = ?", 1, req.BasketId, userID, 1).Count(&count).Error
	if e != nil {
		return nil, e
	}
	if count == 0 {
		return failed("Rebalance not available for this user or basket."), nil
	}
	ss, e := latestScrips(db, "tbl_thematic_basket_rebalance_scrips", req.BasketId)
	if e != nil {
		return nil, e
	}
	if len(ss) == 0 {
		return failed("No rebalance data available."), nil
	}
	out := []M{}
	for _, s := range ss {
		m := thematicScrip(s)
		for _, k := range []string{"weightage", "marketCap", "version", "pdc"} {
			m[k] = nil
		}
		out = append(out, m)
	}
	return success(out), nil
}

// added for sonarqube
func (a *App) thematic(r *http.Request, action string) (any, error) {
	ctx := r.Context()
	db := a.DB.WithContext(ctx)
	idn := identity(r)
	switch action {
	case "invest", "investV1":
		return a.invest(r, action == "investV1")
	case "legacyHoldings":
		return a.legacyHoldings(ctx, idn.UserID)
	case "thematicAll":
		return a.handleThematicAll(ctx, db, idn.UserID)
	case "thematicDetails":
		return handleThematicDetails(db, r.PathValue("id"))
	}
	var req *model.ExecuteBasketOrderReq
	if decode(r, &req) != nil || req == nil {
		return failed(invalidParameter), nil
	}
	switch action {
	case "view":
		e := db.Create(&model.ThematicExecution{UserId: idn.UserID, BasketId: int64(req.BasketId), IsViewed: req.IsViewed, CreatedBy: idn.UserID}).Error
		return message("Basket updated successfully"), e
	case "rebalance":
		return handleThematicRebalance(db, req, idn.UserID)
	case "review":
		return a.thematicReview(db, req)
	}
	return failed(invalidParameter), nil
}

// added for sonarqube
func (a *App) thematicReview(db *gorm.DB, req *model.ExecuteBasketOrderReq) (any, error) {
	if _, e := first(db.Table(tblThematicBasketMaster).Where(idWhere, req.BasketId)); notFound(e) {
		return failed("Basket not found."), nil
	} else if e != nil {
		return nil, e
	}
	if len(req.Scrips) == 0 {
		return failed("Scrip details (LTP) must be provided in request."), nil
	}
	//fetches scrips belonging to the basket from DB
	ss, e := rows(db.Table(tblThematicBasketScrips).Where(basketIDWhere, req.BasketId))
	if e != nil {
		return nil, e
	}
	if len(ss) == 0 {
		return failed("No scrips found for this basket."), nil
	}
	//Maps DB scrips purely by token string
	byToken := map[string]M{}
	for _, s := range ss {
		token := str(s["token"])
		if byToken[token] != nil {
			return failed(msgFailed), nil
		}
		byToken[token] = s
	}
	total := decimal.Zero
	out := []M{}
	//Loops through client request
	for _, s := range req.Scrips {
		b := byToken[val(s.Token)]
		if b == nil {
			return failed("Scrip not found in basket: " + val(s.Token)), nil
		}
		qty := integer(b["qty"]) * req.LotSize
		price := decimal.NewFromFloat(s.Ltp)
		total = total.Add(price.Mul(decimal.NewFromInt(int64(qty))))
		m := project(b, map[string]string{"formattedInsName": "formatted_ins_name", "exchange": "exchange", "transType": "trans_type", "tradingSymbol": "trading_symbol", "pdc": "pdc", "orderType": "order_type", "priceType": "price_type"})
		m["token"] = s.Token
		m["qty"] = qty
		m["price"] = price.Round(2).StringFixed(2)
		m["weightage"] = dec(b["weightage"]).Round(2).StringFixed(2)
		m["adjWeightage"] = m["weightage"]
		out = append(out, m)
	}
	amount := total.Round(2).StringFixed(2)
	return success(M{"scrips": out, "totalInvested": amount, "balance": "0.00", "investmentAmount": amount, "bufferPercentage": a.Config.Business.InvestBuffer}), nil
}

/*
func (a *App) invest(r *http.Request, v1 bool) (any, error) {
	ctx := r.Context()
	db := a.DB.WithContext(ctx)
	id := identity(r)
	var req *model.ExecuteBasketOrderReq
	if decode(r, &req) != nil || req == nil || req.BasketId < 1 || len(req.Scrips) == 0 {
		//added for sonarqube
		// return failed("Invalid Parameter"), nil
		return failed(invalidParameter), nil
	}
	if v1 {
		if !nonempty(req.BasketAction) {
			return failed("Invalid Parameter basketAction"), nil
		}
		if req.Lots <= 0 {
			return failed("Invalid Parameter lots"), nil
		}
	}
	for i := range req.Scrips {
		req.Scrips[i].Source = req.Source
		if !validScrip(req.Scrips[i], true) {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
	}
	if id.Token == "" {
		return httpResult{401, []any{}}, nil
	}
	//added for sonarqube
	// master, e := first(db.Table("tbl_thematic_basket_master").Where("id = ?", req.BasketId))
	master, e := first(db.Table(tblThematicBasketMaster).Where(idWhere, req.BasketId))
	if notFound(e) {
		//added for sonarqube
		// return failed("Invalid basket"), nil
		return failed(invalidBasket), nil
	}
	if e != nil {
		return nil, e
	}
	orders := a.orderRequest(req.Scrips, int64(req.BasketId))
	var responses []Response
	status, e := a.post(ctx, a.Config.Upstream.OrderURL, orders, id.Token, &responses)
	if status == 401 {
		return httpResult{401, []any{}}, nil
	}
	if e != nil {
		return nil, e
	}
	if responses == nil {
		//added for sonarqube
		// return failed("Failed"), nil
		return failed(msgFailed), nil
	}
	exec := model.ThematicExecution{UserId: id.UserID, BasketId: int64(req.BasketId), Lots: 1, OrderResponse: jsonString(responses), OrderRequest: jsonString(orders), ScripDetails: jsonString(req.Scrips), IsExecuted: 1, CreatedBy: id.UserID, UpdatedBy: id.UserID, InvestedAmount: req.InvestmentAmount}
	if !v1 {
		if e = db.Create(&exec).Error; e != nil {
			return nil, e
		}
		return message(a.Config.Business.ExecutionMessage), nil
	}
	if len(responses) == 0 {
		//added for sonarqube
		// return failed("Failed"), nil
		return failed(msgFailed), nil
	}
	exec.Lots = req.Lots
	exec.OrderResponse = jsonString(responses[0])
	exec.Source = val(req.Source)
	exec.BasketName = str(master["basket_name"])
	exec.AnalystName = str(master["analyst_name"])
	exec.UserName = id.Name
	// Persist a broker response even when it contains fewer results than requested.
	// Never retry order placement because a later database operation failed.
	var results []M
	if e = copyJSON(responses[0].Result, &results); e != nil {
		return nil, e
	}
	//added for sonarqube
	// adminScrips, e := rows(db.Table("tbl_thematic_basket_scrips").Where("basket_id = ?", req.BasketId))
	adminScrips, e := rows(db.Table(tblThematicBasketScrips).Where(basketIDWhere, req.BasketId))
	if e != nil {
		return nil, e
	}
	byToken := map[string]M{}
	for _, s := range adminScrips {
		byToken[str(s["token"])] = s
	}
	e = db.Transaction(func(tx *gorm.DB) error {
		return a.saveInvestExecutionDetails(tx, &exec, id, req, byToken, results)
	})
	if e != nil {
		return nil, fmt.Errorf("orders submitted but execution persistence failed: %w", e)
	}
	select {
	case a.jobs <- func() {
		if e := a.refreshOrderBook(a.ctx, exec.Id, id.UserID, results); e != nil {
			a.Log.Error("order book refresh", "execution", exec.Id, "error", e)
		}
	}:
	default:
		a.Log.Error("order book refresh queue full", "execution", exec.Id)
	}
	return message(a.Config.Business.ExecutionMessage), nil
}
*/

// added for sonarqube
func validateInvestRequest(r *http.Request, v1 bool, id Identity) (*model.ExecuteBasketOrderReq, any, error) {
	var req *model.ExecuteBasketOrderReq
	if decode(r, &req) != nil || req == nil || req.BasketId < 1 || len(req.Scrips) == 0 {
		return nil, failed(invalidParameter), nil
	}
	if v1 {
		if !nonempty(req.BasketAction) {
			return nil, failed("Invalid Parameter basketAction"), nil
		}
		if req.Lots <= 0 {
			return nil, failed("Invalid Parameter lots"), nil
		}
	}
	for i := range req.Scrips {
		req.Scrips[i].Source = req.Source
		if !validScrip(req.Scrips[i], true) {
			return nil, failed(invalidParameter), nil
		}
	}
	if id.Token == "" {
		return nil, httpResult{401, []any{}}, nil
	}
	return req, nil, nil
}

// added for sonarqube
func (a *App) persistThematicV1(ctx context.Context, db *gorm.DB, exec *model.ThematicExecution, id Identity, req *model.ExecuteBasketOrderReq, master M, resp Response) error {
	exec.Lots = req.Lots
	exec.OrderResponse = jsonString(resp)
	exec.Source = val(req.Source)
	exec.BasketName = str(master["basket_name"])
	exec.AnalystName = str(master["analyst_name"])
	exec.UserName = id.Name
	var results []M
	if e := copyJSON(resp.Result, &results); e != nil {
		return e
	}
	adminScrips, e := rows(db.Table(tblThematicBasketScrips).Where(basketIDWhere, req.BasketId))
	if e != nil {
		return e
	}
	byToken := map[string]M{}
	for _, s := range adminScrips {
		byToken[str(s["token"])] = s
	}
	e = db.Transaction(func(tx *gorm.DB) error {
		return a.saveInvestExecutionDetails(tx, exec, id, req, byToken, results)
	})
	if e != nil {
		return fmt.Errorf("orders submitted but execution persistence failed: %w", e)
	}
	select {
	case a.jobs <- func() {
		if e := a.refreshOrderBook(a.ctx, exec.Id, id.UserID, results); e != nil {
			a.Log.Error("order book refresh", "execution", exec.Id, "error", e)
		}
	}:
	default:
		a.Log.Error("order book refresh queue full", "execution", exec.Id)
	}
	return nil
}

// added for sonarqube
func (a *App) invest(r *http.Request, v1 bool) (any, error) {
	ctx := r.Context()
	db := a.DB.WithContext(ctx)
	id := identity(r)
	req, resp, err := validateInvestRequest(r, v1, id)
	if resp != nil || err != nil {
		return resp, err
	}
	master, e := first(db.Table(tblThematicBasketMaster).Where(idWhere, req.BasketId))
	if notFound(e) {
		return failed(invalidBasket), nil
	}
	if e != nil {
		return nil, e
	}
	orders := a.orderRequest(req.Scrips, int64(req.BasketId))
	var responses []Response
	status, e := a.post(ctx, a.Config.Upstream.OrderURL, orders, id.Token, &responses)
	if status == 401 {
		return httpResult{401, []any{}}, nil
	}
	if e != nil {
		return nil, e
	}
	if responses == nil {
		return failed(msgFailed), nil
	}
	exec := model.ThematicExecution{UserId: id.UserID, BasketId: int64(req.BasketId), Lots: 1, OrderResponse: jsonString(responses), OrderRequest: jsonString(orders), ScripDetails: jsonString(req.Scrips), IsExecuted: 1, CreatedBy: id.UserID, UpdatedBy: id.UserID, InvestedAmount: req.InvestmentAmount}
	if !v1 {
		if e = db.Create(&exec).Error; e != nil {
			return nil, e
		}
		return message(a.Config.Business.ExecutionMessage), nil
	}
	if len(responses) == 0 {
		return failed(msgFailed), nil
	}
	if e := a.persistThematicV1(ctx, db, &exec, id, req, master, responses[0]); e != nil {
		return nil, e
	}
	return message(a.Config.Business.ExecutionMessage), nil
}

// added for sonarqube
func (a *App) saveInvestExecutionDetails(tx *gorm.DB, exec *model.ThematicExecution, id Identity, req *model.ExecuteBasketOrderReq, byToken map[string]M, results []M) error {
	if e := tx.Create(exec).Error; e != nil {
		return e
	}
	if e := tx.Where("user_id = ? AND research_type = ? AND basket_id = ?", id.UserID, 2, req.BasketId).Delete(&model.ThematicExeMasterEntity{}).Error; e != nil {
		return e
	}
	version := "null"
	if req.Scrips[0].Version != nil {
		version = strconv.Itoa(*req.Scrips[0].Version)
	}
	m := model.ThematicExeMasterEntity{UserId: ptr(id.UserID), BasketId: int64(req.BasketId), ResearchType: 2, Lots: ptr(req.Lots), BasketName: ptr(exec.BasketName), AnalystName: ptr(exec.AnalystName), BasketAction: req.BasketAction, Source: req.Source, Version: ptr(version), ActiveStatus: 1}
	if e := tx.Create(&m).Error; e != nil {
		return e
	}
	q := tx.Model(&model.ExecutionDetail{}).Where("user_id = ? AND research_id = ? AND research_type = ? AND active_status = ?", id.UserID, req.BasketId, 2, 1)
	action := strings.ToUpper(val(req.BasketAction))
	if action == "BUY" {
		q = q.Where("UPPER(basket_action) = ?", "SELL")
	}
	if action == "BUY" || action == "SELL" {
		if e := q.Updates(M{"active_status": 0, "updated_on": a.Now()}).Error; e != nil {
			return e
		}
	}
	return a.createExecutionDetails(tx, exec, id, req, byToken, results)
}

// added for sonarqube
func (a *App) createExecutionDetails(tx *gorm.DB, exec *model.ThematicExecution, id Identity, req *model.ExecuteBasketOrderReq, byToken map[string]M, results []M) error {
	for i, s := range req.Scrips {
		admin := byToken[val(s.Token)]
		if admin == nil {
			a.Log.Error("admin record missing", "basket", req.BasketId, "token", val(s.Token))
			continue
		}
		result := M{}
		if i < len(results) {
			result = results[i]
		}
		orderNo := str(result["orderNo"])
		orderStatus := "EXECUTED"
		successful := true
		if orderNo == "" {
			orderStatus = "FAILED"
			successful = false
		}
		now := model.Now()
		d := model.ExecutionDetail{ActiveStatus: 1, ThematicExecDetails: model.ThematicExecDetails{ExecutionId: ptr(strconv.FormatInt(exec.Id, 10)), UserId: ptr(id.UserID), ResearchId: ptr(req.BasketId), ResearchType: ptr(2), BasketName: ptr(exec.BasketName), LotSize: ptr(req.Lots), BasketAction: req.BasketAction, Version: s.Version, OrderNo: ptr(orderNo), OrderStatus: ptr(orderStatus), IsSuccessful: ptr(successful), OrderResponse: ptr(jsonString(result)), UserPrice: ptr(float32(number(s.Price))), ExecutedPrice: s.Price, Exch: s.Exchange, Token: s.Token, TradingSymbol: ptr(str(admin["trading_symbol"])), TransType: s.TransType, OriginalQty: ptr(integer(s.Qty)), ExecutedQty: ptr(integer(s.Qty)), RecommQty: ptr(integer(admin["qty"])), CreatedBy: ptr(id.UserID), UpdatedBy: ptr(id.UserID), CreatedOn: now, UpdatedOn: now, ExecutedOn: now}}
		if e := tx.Create(&d).Error; e != nil {
			return e
		}
	}
	return nil
}

func isOrderPlaced(resp []Response) bool {
	for _, rr := range resp {
		if rr.Result == nil {
			continue
		}
		switch res := rr.Result.(type) {
		case []M:
			for _, x := range res {
				if nonempty(x["orderNo"]) {
					return true
				}
			}
		case []any:
			for _, item := range res {
				if xm, ok := item.(map[string]any); ok && nonempty(xm["orderNo"]) {
					return true
				}
			}
		case map[string]any:
			if nonempty(res["orderNo"]) {
				return true
			}
		default:
			var results []M
			if copyJSON(rr.Result, &results) == nil {
				for _, x := range results {
					if nonempty(x["orderNo"]) {
						return true
					}
				}
			}
		}
	}
	return false
}

// added for sonarqube
func aggregateOrderRequestScrips(ss []M, orderReqJSON string) []M {
	var req []M
	if json.Unmarshal([]byte(orderReqJSON), &req) == nil {
		for _, s := range req {
			found := false
			for _, existing := range ss {
				if str(existing["token"]) == str(s["token"]) {
					existing["qty"] = strconv.Itoa(integer(existing["qty"]) + integer(s["qty"]))
					found = true
					break
				}
			}
			if !found {
				ss = append(ss, M{"token": s["token"], "exchange": s["exchange"], "tradingSymbol": s["tradingSymbol"], "qty": str(s["qty"])})
			}
		}
	}
	return ss
}

func (a *App) legacyHoldings(ctx context.Context, user string) (any, error) {
	db := a.DB.WithContext(ctx)
	//added for sonarqube
	// rs, e := rows(db.Table("tbl_user_thematic_exec AS e").Select("e.*, b.basket_name AS master_name, b.created_on AS basket_created_on").Joins("JOIN tbl_thematic_basket_master b ON b.id = e.basket_id").Where("e.user_id = ? AND e.is_executed = ?", user, 1).Order("e.invested_amount DESC"))
	rs, e := rows(db.Table("tbl_user_thematic_exec AS e").Select("e.*, b.basket_name AS master_name, b.created_on AS basket_created_on").Joins("JOIN "+tblThematicBasketMaster+" b ON b.id = e.basket_id").Where("e.user_id = ? AND e.is_executed = ?", user, 1).Order("e.invested_amount DESC"))
	if e != nil {
		return nil, e
	}
	byID := map[int]M{}
	ids := []int{}
	for _, r := range rs {
		var resp []Response
		if json.Unmarshal([]byte(str(r["order_response"])), &resp) != nil {
			continue
		}
		if !isOrderPlaced(resp) {
			continue
		}
		id := integer(r["basket_id"])
		h := byID[id]
		if h == nil {
			h = M{"lotSize": "0", "scripList": []M{}}
			ids = append(ids, id)
			byID[id] = h
		}
		h["basketName"] = r["master_name"]
		h["userId"] = user
		h["investedAmount"] = number(r["invested_amount"])
		h["createdDate"] = a.displayDate(r["basket_created_on"])
		h["executedDate"] = a.displayDate(r["created_on"])
		h["lotSize"] = strconv.Itoa(integer(h["lotSize"]) + integer(r["lots"]))
		h["scripList"] = aggregateOrderRequestScrips(h["scripList"].([]M), str(r["order_request"]))
	}
	out := []M{}
	for _, id := range ids {
		out = append(out, byID[id])
	}
	return success(out), nil
}
func (a *App) displayDate(v any) any {
	if v == nil {
		return nil
	}
	var t model.Timestamp
	if e := t.Scan(v); e != nil {
		return nil
	}
	return t.Time().In(a.location).Format("02 Jan 2006 03:04 PM")
}
