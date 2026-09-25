package app

import (
	"basket/internal/model"
	"context"
	"gorm.io/gorm"
	"strconv"
)

// These methods retain DAO/service operations that are present in Java but are
// not published by its REST controllers. Callers can reuse them without adding APIs.
func (a *App) CreateThematicBasket(ctx context.Context, req model.ThematicBasketRequest, createdBy string) (int64, error) {
	m := model.ThematicMaster{Category: req.Category, SubCategory: req.SubCategory, BasketName: req.BasketName, ShortDescription: req.ShortDescription, LongDescription: req.LongDescription, Tag: req.Tag, RiskLevel: req.RiskLevel, Benchmark: req.Benchmark, Exposure: req.Exposure, InvestmentDuration: req.InvestmentDuration, ReviewFrequency: req.ReviewFrequency, ScripCount: req.ScripCount, CreatedBy: ptr(createdBy)}
	e := a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&m).Error; e != nil {
			return e
		}
		for _, s := range req.Scrips {
			if _, e := strconv.Atoi(val(s.Qty)); e != nil {
				return e
			}
			v := model.ThematicScrip{BasketId: m.Id, Exchange: s.Exchange, Token: s.Token, Qty: s.Qty, TransType: s.TransType, Price: s.Price, CreatedBy: ptr(createdBy)}
			if e := tx.Create(&v).Error; e != nil {
				return e
			}
		}
		return nil
	})
	return m.Id, e
}
func (a *App) ThematicScripsForExecution(ctx context.Context, id int64, defaults bool) ([]model.ScripRequestModel, error) {
	rows, e := rows(a.DB.WithContext(ctx).Table("tbl_thematic_basket_scrips").Where("basket_id = ?", id))
	if e != nil {
		return nil, e
	}
	out := []model.ScripRequestModel{}
	for _, r := range rows {
		m := project(r, map[string]string{"exchange": "exchange", "token": "token", "qty": "qty", "transType": "trans_type", "price": "price", "orderType": "order_type", "priceType": "price_type", "ret": "ret", "source": "source", "triggerPrice": "trigger_price", "disClosedQty": "disclosed_qty", "mktProtection": "mkt_protection", "target": "target", "stopLoss": "stop_loss", "trailingStopLoss": "trailing_stop_loss", "tradingSymbol": "trading_symbol"})
		if defaults {
			side := ""
			if buy(r["trans_type"]) {
				side = "B"
			}
			m["transType"] = side
			m["product"] = "CNC"
			m["ret"] = "DAY"
			m["triggerPrice"] = "0"
			for _, k := range []string{"disClosedQty", "mktProtection", "target", "stopLoss", "trailingStopLoss"} {
				m[k] = ""
			}
		}
		var s model.ScripRequestModel
		if e = copyJSON(m, &s); e != nil {
			return nil, e
		}
		out = append(out, s)
	}
	return out, nil
}
func (a *App) PendingAndIncompleteOrders(ctx context.Context) ([]model.ExecutionDetail, error) {
	out := []model.ExecutionDetail{}
	e := a.DB.WithContext(ctx).Where("LOWER(order_status) = ? OR (order_status = ? AND (executed_price IS NULL OR executed_price IN ? OR executed_on IS NULL))", "pending", "EXECUTED", []string{"0", "0.0"}).Find(&out).Error
	return out, e
}
func (a *App) OrderStatusByOrderNo(ctx context.Context, no string) (*model.OrderStatusFeedEntity, error) {
	var f model.OrderStatusFeedEntity
	e := a.DB.WithContext(ctx).Where("order_no = ?", no).Order("updated_on DESC").First(&f).Error
	if e != nil {
		return nil, e
	}
	f.ActiveStatus = 0
	return &f, nil
}

// ResearchData retains the older DAO's category/tag/expiry filtering; the active
// getResearchCall API instead uses its analyst/status/seven-day implementation.
func (a *App) ResearchData(ctx context.Context, user string, req *model.ResearchCallRequest) ([]model.ResearchMaster, error) {
	q := a.DB.WithContext(ctx).Where("user_id IN ?", []string{user, "ALL"})
	if req != nil {
		if nonempty(req.Category) {
			q = q.Where("category = ?", req.Category)
		}
		if nonempty(req.SubCategory) {
			q = q.Where("subcategory = ?", req.SubCategory)
		}
		if len(req.Tags) > 0 {
			q = q.Where("tags IN ?", req.Tags)
		}
		from, to := val(req.FromDate), val(req.ToDate)
		switch {
		case from != "" && to != "":
			q = q.Where("expiry_date BETWEEN ? AND ?", from, to)
		case from != "":
			q = q.Where("expiry_date = ?", from)
		default:
			q = q.Where("expiry_date >= ?", a.midnight().AddDate(0, 0, -30))
			if to == "" {
				q = q.Where("expiry_date <= ?", a.midnight())
			}
		}
		if nonempty(req.Status) {
			q = q.Where("status = ?", req.Status)
		}
	}
	out := []model.ResearchMaster{}
	e := q.Find(&out).Error
	return out, e
}
func (a *App) MaxBasketID(ctx context.Context) (int64, error) {
	var id int64
	e := a.DB.WithContext(ctx).Model(&model.BasketNameEntity{}).Select("COALESCE(MAX(basket_id), 0)").Scan(&id).Error
	return id + 1, e
}
