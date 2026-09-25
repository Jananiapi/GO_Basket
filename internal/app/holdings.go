package app

import (
	"basket/internal/model"
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type position struct {
	token, exchange, symbol   string
	qty, recommended, version int
	buyQty, buyValue          decimal.Decimal
	last                      any
}

func executed(v any) bool { return strings.EqualFold(str(v), "EXECUTED") }
func buy(v any) bool      { return strings.EqualFold(str(v), "BUY") }
func timestamp(v any) time.Time {
	var t model.Timestamp
	if e := t.Scan(v); e != nil {
		return time.Time{}
	}
	return t.Time()
}
func holdingDate(v any) any {
	if v == nil {
		return nil
	}
	t := timestamp(v)
	if t.IsZero() {
		return nil
	}
	return t.Format("2006-01-02 15:04:05.0")
}
func determineAction(current, recommended int) any {
	switch {
	case current == 0 && recommended > 0:
		return "Add New"
	case current > 0 && recommended > current:
		return "Add More"
	case current > 0 && recommended > 0 && recommended < current:
		return "Reduce"
	case current > 0 && recommended == 0:
		return "Exit"
	default:
		return nil
	}
}

// added for sonarqube
const tblUserThematicExecDetails = "tbl_user_thematic_exec_details"

func (a *App) holdings(r *http.Request, v1 bool) (any, error) {
	ctx := r.Context()
	user := identity(r).UserID
	db := a.DB.WithContext(ctx)
	var ids []int
	e := db.Table(tblUserThematicExecDetails).Where("user_id = ? AND UPPER(order_status) = ? AND active_status = ?", user, "EXECUTED", 1).Distinct("research_id").Pluck("research_id", &ids).Error
	if e != nil {
		return M{"status": "FAILURE", "message": "Internal server error", "data": []any{}}, e
	}
	if len(ids) == 0 {
		return M{"status": "FAILURE", "message": "User info not found", "data": []any{}}, nil
	}
	sort.Ints(ids)

	inactiveMap := map[int]bool{}
	if v1 {
		var inactiveIDs []int
		if e := db.Table("tbl_user_thematic_master").Where("user_id = ? AND basket_id IN ? AND (UPPER(basket_action) = ? OR active_status <> ?)", user, ids, "SELL", 1).Pluck("basket_id", &inactiveIDs).Error; e != nil {
			return nil, e
		}
		for _, inID := range inactiveIDs {
			inactiveMap[inID] = true
		}
	}

	masters, e := rows(db.Table(tblThematicBasketMaster).Where("id IN ? AND LOWER(status) <> ?", ids, "closed"))
	if e != nil {
		return nil, e
	}
	masterMap := map[int]M{}
	for _, m := range masters {
		masterMap[integer(m["id"])] = m
	}

	allDetails, e := rows(db.Table(tblUserThematicExecDetails).Where("user_id = ? AND research_id IN ?", user, ids).Order("id"))
	if e != nil {
		return nil, e
	}
	detailsByResearchID := map[int][]M{}
	for _, d := range allDetails {
		rid := integer(d["research_id"])
		detailsByResearchID[rid] = append(detailsByResearchID[rid], d)
	}

	anyIDs := make([]any, len(ids))
	for i, id := range ids {
		anyIDs[i] = id
	}
	scripsByBasket, e := latestScripsBatch(db, tblThematicBasketScrips, anyIDs)
	if e != nil {
		return nil, e
	}

	out := []M{}
	for _, id := range ids {
		if v1 && inactiveMap[id] {
			continue
		}
		master := masterMap[id]
		if master == nil {
			continue
		}
		details := detailsByResearchID[id]
		if len(details) == 0 {
			continue
		}
		lots, firstBuy := calculateLotsAndFirstBuy(details)
		if lots == 0 {
			continue
		}
		positions := aggregatePositions(details, v1)
		userByToken := map[string]*position{}
		for _, p := range positions {
			userByToken[p.token] = p
		}
		recs := scripsByBasket[str(id)]
		if recs == nil {
			recs = []M{}
		}
		latest, current := calculateLatestAndCurrentVersions(recs, userByToken)
		scripList, total := buildHoldingsScripList(recs, userByToken, lots, latest, current, v1)
		amount, _ := total.Round(2).Float64()
		rebalance := 0
		if latest > current {
			rebalance = 1
		}
		out = append(out, M{"basketId": id, "basketName": master["basket_name"], "userId": user, "investedAmount": amount, "currentVersion": current, "rebalancedVersion": latest, "isRebalanced": rebalance, "scripList": scripList, "lotSize": lots, "executedDate": holdingDate(firstBuy)})
	}
	return success(out), nil
}

// added for sonarqube
func isDetailEligible(d M, v1 bool) bool {
	status := str(d["order_status"])
	if v1 {
		return integer(d["active_status"]) == 1 && (executed(status) || strings.EqualFold(status, "complete"))
	}
	return executed(status)
}

// added for sonarqube
func updatePositionVersion(p *position, d M, v1 bool) {
	ver := integer(d["version"])
	recomm := integer(d["recomm_qty"])
	if v1 {
		if ver > p.version {
			p.version = ver
			p.recommended = recomm
			p.symbol = str(d["trading_symbol"])
		} else if ver == p.version && recomm > p.recommended {
			p.recommended = recomm
		}
		return
	}
	if ver > p.version {
		p.version = ver
	}
	if recomm > p.recommended {
		p.recommended = recomm
	}
}

// added for sonarqube
func updatePositionFromDetail(p *position, d M, v1 bool) {
	q := integer(d["executed_qty"])
	if buy(d["trans_type"]) {
		p.qty += q
		p.buyQty = p.buyQty.Add(decimal.NewFromInt(int64(q)))
		p.buyValue = p.buyValue.Add(dec(d["executed_price"]).Mul(decimal.NewFromInt(int64(q))))
	} else {
		p.qty -= q
	}
	updatePositionVersion(p, d, v1)
	if d["executed_on"] != nil && (p.last == nil || timestamp(d["executed_on"]).After(timestamp(p.last))) {
		p.last = d["executed_on"]
	}
}

// added for sonarqube
func getOrCreatePosition(positions map[string]*position, d M) *position {
	key := str(d["token"]) + "_" + str(d["exch"])
	p := positions[key]
	if p == nil {
		p = &position{token: str(d["token"]), exchange: str(d["exch"]), symbol: str(d["trading_symbol"])}
		positions[key] = p
	}
	return p
}

// added for sonarqube
func aggregatePositions(details []M, v1 bool) map[string]*position {
	positions := map[string]*position{}
	for _, d := range details {
		if isDetailEligible(d, v1) {
			updatePositionFromDetail(getOrCreatePosition(positions, d), d, v1)
		}
	}
	return positions
}

// added for sonarqube
func updateExecLotAndDate(d M, lotsByExec map[string]int, firstBuy any) any {
	if integer(d["active_status"]) == 1 {
		key := str(d["execution_id"])
		if integer(d["lot_size"]) > lotsByExec[key] {
			lotsByExec[key] = integer(d["lot_size"])
		}
	}
	if d["executed_on"] != nil && (firstBuy == nil || timestamp(d["executed_on"]).Before(timestamp(firstBuy))) {
		return d["executed_on"]
	}
	return firstBuy
}

// added for sonarqube
func calculateLotsAndFirstBuy(details []M) (int, any) {
	lotsByExec := map[string]int{}
	var firstBuy any
	for _, d := range details {
		if executed(d["order_status"]) && buy(d["basket_action"]) && buy(d["trans_type"]) {
			firstBuy = updateExecLotAndDate(d, lotsByExec, firstBuy)
		}
	}
	lots := 0
	for _, l := range lotsByExec {
		lots += l
	}
	return lots, firstBuy
}

// added for sonarqube
func makeHoldingScrip(token, exch, symbol string, p *position, recommended, latest, lots int) (M, decimal.Decimal) {
	m := M{"token": token, "exchange": exch, "tradingSymbol": symbol, "currentQty": 0, "orgRecoQty": 0, "executedPrice": 0.0, "executedOn": nil, "version": 1, "recommendedQty": recommended, "recommendedVersion": latest, "rebalancedAction": nil, "rebalancedQty": recommended * lots}
	val := decimal.Zero
	if p != nil {
		avg := decimal.Zero
		if !p.buyQty.IsZero() {
			avg = p.buyValue.Div(p.buyQty)
		}
		m["currentQty"] = p.qty
		m["orgRecoQty"] = p.recommended
		m["executedPrice"], _ = avg.Float64()
		m["executedOn"] = holdingDate(p.last)
		m["version"] = p.version
		val = avg.Mul(decimal.NewFromInt(int64(p.qty)))
	}
	return m, val
}

// added for sonarqube
func appendExitHoldings(out []M, userByToken map[string]*position, seen map[string]bool, latest, lots int) []M {
	tokens := []string{}
	for token := range userByToken {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	for _, token := range tokens {
		if !seen[token] {
			p := userByToken[token]
			m, _ := makeHoldingScrip(token, p.exchange, p.symbol, p, 0, latest, lots)
			m["rebalancedAction"] = "Exit"
			out = append(out, m)
		}
	}
	return out
}

// added for sonarqube
func buildHoldingsScripList(recs []M, userByToken map[string]*position, lots, latest, current int, v1 bool) ([]M, decimal.Decimal) {
	total := decimal.Zero
	out := []M{}
	seen := map[string]bool{}
	for _, s := range recs {
		token := str(s["token"])
		seen[token] = true
		p := userByToken[token]
		qty := integer(s["qty"])
		m, scripVal := makeHoldingScrip(token, str(s["exchange"]), str(s["trading_symbol"]), p, qty, latest, lots)
		total = total.Add(scripVal)
		if latest > current {
			cq := integer(m["currentQty"])
			if v1 {
				cq = integer(m["orgRecoQty"])
			}
			m["rebalancedAction"] = determineAction(cq, qty)
		}
		if integer(m["currentQty"]) > 0 || qty > 0 {
			out = append(out, m)
		}
	}
	out = appendExitHoldings(out, userByToken, seen, latest, lots)
	return out, total
}

// added for sonarqube
func isV1BasketActive(db *gorm.DB, user string, id int) (bool, error) {
	var n int64
	if e := db.Table("tbl_user_thematic_master").Where("user_id = ? AND basket_id = ? AND (UPPER(basket_action) = ? OR active_status <> ?)", user, id, "SELL", 1).Count(&n).Error; e != nil {
		return false, e
	}
	return n == 0, nil
}

// added for sonarqube
func calculateLatestAndCurrentVersions(recs []M, userByToken map[string]*position) (int, int) {
	latest, current := 1, 1
	for _, s := range recs {
		if integer(s["version"]) > latest {
			latest = integer(s["version"])
		}
		if p := userByToken[str(s["token"])]; p != nil && p.version > current {
			current = p.version
		}
	}
	return latest, current
}

func (a *App) buildHolding(ctx context.Context, user string, id int, v1 bool) (M, error) {
	db := a.DB.WithContext(ctx)
	if v1 {
		active, e := isV1BasketActive(db, user, id)
		if e != nil || !active {
			return nil, e
		}
	}
	//added for sonarqube
	// master, e := first(db.Table("tbl_thematic_basket_master").Where("id = ? AND LOWER(status) <> ?", id, "closed"))
	master, e := first(db.Table(tblThematicBasketMaster).Where(idWhere+" AND LOWER(status) <> ?", id, "closed"))
	if notFound(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	//added for sonarqube
	// details, e := rows(db.Table("tbl_user_thematic_exec_details").Where("user_id = ? AND research_id = ?", user, id).Order("id"))
	details, e := rows(db.Table(tblUserThematicExecDetails).Where("user_id = ? AND research_id = ?", user, id).Order("id"))
	if e != nil {
		return nil, e
	}
	lots, firstBuy := calculateLotsAndFirstBuy(details)
	if lots == 0 {
		return nil, nil
	}
	positions := aggregatePositions(details, v1)
	userByToken := map[string]*position{}
	for _, p := range positions {
		userByToken[p.token] = p
	}
	//added for sonarqube
	// recs, e := latestScrips(db, "tbl_thematic_basket_scrips", id)
	recs, e := latestScrips(db, tblThematicBasketScrips, id)
	if e != nil {
		return nil, e
	}
	latest, current := calculateLatestAndCurrentVersions(recs, userByToken)
	out, total := buildHoldingsScripList(recs, userByToken, lots, latest, current, v1)
	amount, _ := total.Round(2).Float64()
	rebalance := 0
	if latest > current {
		rebalance = 1
	}
	return M{"basketId": id, "basketName": master["basket_name"], "userId": user, "investedAmount": amount, "currentVersion": current, "rebalancedVersion": latest, "isRebalanced": rebalance, "scripList": out, "lotSize": lots, "executedDate": holdingDate(firstBuy)}, nil
}
