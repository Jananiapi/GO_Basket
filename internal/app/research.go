package app

import (
	"errors"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// added for sonarqube
const (
	noDataFound          = "No data found"
	actionResearchBasket = "researchBasket"
)

/*
func (a *App) research(r *http.Request, action string) (any, error) {
	db := a.DB.WithContext(r.Context())
	var req M
	if r.Method == "POST" {
		if decode(r, &req) != nil || req == nil {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
	}
	switch action {
	case "status":
		var statuses []*string
		e := db.Table("tbl_researchcall_master").Distinct("status").Pluck("status", &statuses).Error
		if e != nil {
			return nil, e
		}
		if len(statuses) == 0 {
			//added for sonarqube
			// return failed("No data found"), nil
			return failed(noDataFound), nil
		}
		return success(statuses), nil
	case "sectors", "sector", "reports":
		q := db.Table("tbl_sector_reports").Where("active_status = ?", 1)
		if action == "sector" {
			q = q.Where("id = ?", integer(r.PathValue("id")))
		}
		if action == "reports" && nonempty(req["basketType"]) {
			q = q.Where("type = ?", req["basketType"])
		}
		rs, e := rows(q)
		if e != nil {
			return nil, e
		}
		if len(rs) == 0 && action != "reports" {
			//added for sonarqube
			// return failed("No data found"), nil
			return failed(noDataFound), nil
		}
		out := []M{}
		for _, row := range rs {
			m := project(row, map[string]string{"id": "id", "title": "name", "attachment": "attachment", "description": "description", "type": "type", "url": "url", "createdOn": "created_on"})
			m["createdOn"] = epoch(m["createdOn"])
			if action == "sector" {
				delete(m, "title")
				m["name"] = row["name"]
				m["createdBy"] = integer(row["created_by"])
				m["updatedBy"] = integer(row["updated_by"])
				m["updatedOn"] = epoch(row["updated_on"])
				m["activeStatus"] = integer(row["active_status"])
			}
			out = append(out, m)
		}
		return success(out), nil
	}
	var ids []int64
	if e := db.Table("tbl_researchcall_usermapping").Where("user_id IN ?", []string{identity(r).UserID, "ALL"}).Distinct("researchcall_id").Pluck("researchcall_id", &ids).Error; e != nil {
		return nil, e
	}
	if len(ids) == 0 {
		//added for sonarqube
		// if action == "researchBasket" {
		if action == actionResearchBasket {
			//added for sonarqube
			// return failed("No data found"), nil
			return failed(noDataFound), nil
		}
		return failed("No data found for this user"), nil
	}
	q := db.Table("tbl_researchcall_master").Where("active_status = ? AND id IN ?", 1, ids)
	if action == "research" {
		if nonempty(req["analystName"]) {
			q = q.Where("analyst_name = ?", strings.TrimSpace(str(req["analystName"])))
		}
		if nonempty(req["status"]) {
			q = q.Where("LOWER(status) = ?", strings.ToLower(strings.TrimSpace(str(req["status"]))))
		} else {
			q = q.Where("LOWER(status) = ? OR (LOWER(status) = ? AND created_on >= ?)", "open", "closed", a.Now().AddDate(0, 0, -a.Config.Business.ClosedResearchDays))
		}
		q = q.Order("created_on DESC")
	}
	orders, e := rows(q)
	if e != nil {
		return nil, e
	}
	if len(orders) == 0 {
		//added for sonarqube
		// return failed("No data found"), nil
		return failed(noDataFound), nil
	}
	out, e := groupResearchOrders(db, orders, action)
	if e != nil {
		return nil, e
	}
	return success(out), nil
}
*/

// added for sonarqube
func handleResearchStatus(db *gorm.DB) (any, error) {
	var statuses []*string
	e := db.Table("tbl_researchcall_master").Distinct("status").Pluck("status", &statuses).Error
	if e != nil {
		return nil, e
	}
	if len(statuses) == 0 {
		return failed(noDataFound), nil
	}
	return success(statuses), nil
}

// added for sonarqube
func handleResearchSectorAndReports(r *http.Request, db *gorm.DB, action string, req M) (any, error) {
	q := db.Table("tbl_sector_reports").Where("active_status = ?", 1)
	if action == "sector" {
		q = q.Where("id = ?", integer(r.PathValue("id")))
	}
	if action == "reports" && nonempty(req["basketType"]) {
		q = q.Where("type = ?", req["basketType"])
	}
	rs, e := rows(q)
	if e != nil {
		return nil, e
	}
	if len(rs) == 0 && action != "reports" {
		return failed(noDataFound), nil
	}
	out := []M{}
	for _, row := range rs {
		m := project(row, map[string]string{"id": "id", "title": "name", "attachment": "attachment", "description": "description", "type": "type", "url": "url", "createdOn": "created_on"})
		m["createdOn"] = epoch(m["createdOn"])
		if action == "sector" {
			delete(m, "title")
			m["name"] = row["name"]
			m["createdBy"] = integer(row["created_by"])
			m["updatedBy"] = integer(row["updated_by"])
			m["updatedOn"] = epoch(row["updated_on"])
			m["activeStatus"] = integer(row["active_status"])
		}
		out = append(out, m)
	}
	return success(out), nil
}

// added for sonarqube
func (a *App) handleResearchOrders(r *http.Request, db *gorm.DB, action string, req M) (any, error) {
	var ids []int64
	if e := db.Table("tbl_researchcall_usermapping").Where("user_id IN ?", []string{identity(r).UserID, "ALL"}).Distinct("researchcall_id").Pluck("researchcall_id", &ids).Error; e != nil {
		return nil, e
	}
	if len(ids) == 0 {
		if action == actionResearchBasket {
			return failed(noDataFound), nil
		}
		return failed("No data found for this user"), nil
	}
	q := db.Table("tbl_researchcall_master").Where("active_status = ? AND id IN ?", 1, ids)
	if action == "research" {
		if nonempty(req["analystName"]) {
			q = q.Where("analyst_name = ?", strings.TrimSpace(str(req["analystName"])))
		}
		if nonempty(req["status"]) {
			q = q.Where("LOWER(status) = ?", strings.ToLower(strings.TrimSpace(str(req["status"]))))
		} else {
			q = q.Where("LOWER(status) = ? OR (LOWER(status) = ? AND created_on >= ?)", "open", "closed", a.Now().AddDate(0, 0, -a.Config.Business.ClosedResearchDays))
		}
		q = q.Order("created_on DESC")
	}
	orders, e := rows(q)
	if e != nil {
		return nil, e
	}
	if len(orders) == 0 {
		return failed(noDataFound), nil
	}
	out, e := groupResearchOrders(db, orders, action)
	if e != nil {
		return nil, e
	}
	return success(out), nil
}

// added for sonarqube
func (a *App) research(r *http.Request, action string) (any, error) {
	db := a.DB.WithContext(r.Context())
	var req M
	if r.Method == "POST" {
		if decode(r, &req) != nil || req == nil {
			return failed(invalidParameter), nil
		}
	}
	switch action {
	case "status":
		return handleResearchStatus(db)
	case "sectors", "sector", "reports":
		return handleResearchSectorAndReports(r, db, action, req)
	}
	return a.handleResearchOrders(r, db, action, req)
}

// added for sonarqube
func fetchResearchScrips(db *gorm.DB, orderID any, action string) ([]M, error) {
	ss, e := rows(db.Table("tbl_research_scrip").Where("researchcall_id = ?", orderID).Order("created_on DESC"))
	if e != nil {
		return nil, e
	}
	scrips := []M{}
	for _, s := range ss {
		mapping := map[string]string{"lotSize": "lot_size", "exchange": "exchange", "token": "token", "tradingSymbol": "trading_symbol", "qty": "qty", "expiry": "expiry", "product": "product", "transType": "trans_type", "priceType": "price_type", "orderType": "order_type", "ret": "retention", "triggerPrice": "trigger_price", "disClosedQty": "dis_closed_qty", "mktProtection": "mkt_protection", "trailingStopLoss": "trailing_stop_loss", "formattedInsName": "formatted_ins_name", "weekTag": "week_tag", "validityDays": "validity_days"}
		if action == actionResearchBasket {
			mapping["expiryDate"] = "expiry_date"
		} else {
			for k, c := range map[string]string{"scripId": "id", "priceUpperBound": "price_upper_bound", "priceLowerBound": "price_lower_bound", "targetUpperBound": "target_upper_bound", "targetLowerBound": "target_lower_bound", "stopLossUpperBound": "stoploss_upper_bound", "stopLossLowerBound": "stoploss_lower_bound"} {
				mapping[k] = c
			}
		}
		m := project(s, mapping)
		m["expiry"] = s["expiry"]
		if action == actionResearchBasket {
			m["basketId"] = orderID
		}
		scrips = append(scrips, m)
	}
	return scrips, nil
}

// added for sonarqube
func buildResearchOrder(o M, scrips []M) M {
	m := project(o, map[string]string{"basketId": "id", "createdOn": "created_on", "expiryDate": "expiry_date", "status": "status", "speclizationTag": "speclization_tag", "description": "longdescription", "investmentDuration": "investment_duration", "shortDesc": "shortdescription", "analystName": "analyst_name", "attachement": "attachement"})
	m["createdOn"] = str(o["created_on"])
	m["expiryDate"] = epoch(o["expiry_date"])
	if !nonempty(m["attachement"]) {
		m["attachement"] = ""
	}
	m["sortOrder"] = 0
	m["remarks"] = nil
	m["scripdetails"] = scrips
	return m
}

// added for sonarqube
func groupResearchOrders(db *gorm.DB, orders []M, action string) ([]M, error) {
	if len(orders) == 0 {
		return []M{}, nil
	}
	orderIDs := make([]any, len(orders))
	for i, o := range orders {
		orderIDs[i] = o["id"]
	}
	allScrips, e := rows(db.Table("tbl_research_scrip").Where("researchcall_id IN ?", orderIDs).Order("created_on DESC"))
	if e != nil {
		return nil, e
	}
	scripsByOrder := map[string][]M{}
	for _, s := range allScrips {
		orderID := s["researchcall_id"]
		mapping := map[string]string{"lotSize": "lot_size", "exchange": "exchange", "token": "token", "tradingSymbol": "trading_symbol", "qty": "qty", "expiry": "expiry", "product": "product", "transType": "trans_type", "priceType": "price_type", "orderType": "order_type", "ret": "retention", "triggerPrice": "trigger_price", "disClosedQty": "dis_closed_qty", "mktProtection": "mkt_protection", "trailingStopLoss": "trailing_stop_loss", "formattedInsName": "formatted_ins_name", "weekTag": "week_tag", "validityDays": "validity_days"}
		if action == actionResearchBasket {
			mapping["expiryDate"] = "expiry_date"
		} else {
			for k, c := range map[string]string{"scripId": "id", "priceUpperBound": "price_upper_bound", "priceLowerBound": "price_lower_bound", "targetUpperBound": "target_upper_bound", "targetLowerBound": "target_lower_bound", "stopLossUpperBound": "stoploss_upper_bound", "stopLossLowerBound": "stoploss_lower_bound"} {
				mapping[k] = c
			}
		}
		m := project(s, mapping)
		m["expiry"] = s["expiry"]
		if action == actionResearchBasket {
			m["basketId"] = orderID
		}
		key := str(orderID)
		scripsByOrder[key] = append(scripsByOrder[key], m)
	}

	grouped := map[string]map[string][]M{}
	categories := []string{}
	for _, o := range orders {
		key := str(o["id"])
		scrips := scripsByOrder[key]
		if scrips == nil {
			scrips = []M{}
		}
		if action == actionResearchBasket && len(scrips) == 0 {
			continue
		}
		category, sub := str(o["category"]), str(o["subcategory"])
		if grouped[category] == nil {
			categories = append(categories, category)
			grouped[category] = map[string][]M{}
		}
		if grouped[category][sub] == nil {
			grouped[category][sub] = []M{}
		}
		if action == actionResearchBasket {
			grouped[category][sub] = append(grouped[category][sub], scrips...)
		} else {
			grouped[category][sub] = append(grouped[category][sub], buildResearchOrder(o, scrips))
		}
	}
	out := []M{}
	for _, c := range categories {
		out = append(out, M{c: grouped[c]})
	}
	return out, nil
}
func notFound(e error) bool { return errors.Is(e, gorm.ErrRecordNotFound) }
