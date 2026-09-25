package app

import (
	"basket/internal/model"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// added by janani for resolving openissue in sonarqube
const invalidBasket = "Invalid basket"
const noRecordsFound = "No records found"
const basketIDWhere = "basket_id = ?"
const basketUserWhere = "basket_id = ? AND user_id = ?"

func validScrip(s model.ScripRequestModel, execute bool) bool {
	for _, f := range []*string{s.Exchange, s.Token, s.Qty, s.Price, s.Product, s.TransType, s.PriceType, s.OrderType, s.Ret, s.Source} {
		if !nonempty(f) {
			return false
		}
	}
	if execute {
		return nonempty(s.TradingSymbol)
	}
	q, e := strconv.Atoi(val(s.Qty))
	return e == nil && q > 0
}
func owned(db *gorm.DB, id int64, user string) (model.BasketNameEntity, error) {
	var b model.BasketNameEntity
	// e := db.Where("basket_id = ? AND user_id = ?", id, user).First(&b).Error
	e := db.Where(basketUserWhere, id, user).First(&b).Error
	return b, e
}
func (a *App) basketReports(ctx context.Context, user string) (Response, error) {
	var rows []struct {
		BasketId   int64            `gorm:"column:basket_id"`
		BasketName *string          `gorm:"column:basket_name"`
		IsExecuted *string          `gorm:"column:is_executed"`
		CreatedOn  *model.Timestamp `gorm:"column:created_on"`
		ScripCount int64            `gorm:"column:scrip_count"`
	}
	db := a.DB.WithContext(ctx)
	if e := db.Model(&model.BasketNameEntity{}).
		Select("basket_id, basket_name, is_executed, created_on, (SELECT COUNT(1) FROM tbl_basket_order_scrip WHERE tbl_basket_order_scrip.basket_id = tbl_basket_order.basket_id) AS scrip_count").
		Where("user_id = ? AND research_call = ?", user, 0).
		Order("basket_id DESC").
		Find(&rows).Error; e != nil {
		return Response{}, e
	}
	out := []M{}
	for _, r := range rows {
		out = append(out, M{"basketId": r.BasketId, "basketName": r.BasketName, "isExecuted": r.IsExecuted, "createdOn": r.CreatedOn, "scripCount": r.ScripCount})
	}
	return success(out), nil
}
func (a *App) scrips(ctx context.Context, id int64, refreshLot bool) (Response, error) {
	out := []model.BasketScripEntity{}
	// if e := a.DB.WithContext(ctx).Where("basket_id = ?", id).Order("id").Find(&out).Error; e != nil {
	if e := a.DB.WithContext(ctx).Where(basketIDWhere, id).Order("id").Find(&out).Error; e != nil {
		return Response{}, e
	}
	if refreshLot {
		if len(out) == 0 {
			//added for sonarqube
			// return message("No records found"), nil
			return message(noRecordsFound), nil
		}
		for i := range out {
			c, e := a.contract(ctx, val(out[i].Exchange), val(out[i].Token))
			if e != nil {
				return Response{}, e
			}
			if nonempty(c.LotSize) {
				out[i].LotSize = c.LotSize
			}
		}
	}
	return success(out), nil
}
func (a *App) prepareScrip(ctx context.Context, s model.ScripRequestModel, id int64, user string, existing *model.BasketScripEntity) (model.BasketScripEntity, error) {
	c, e := a.contract(ctx, val(s.Exchange), val(s.Token))
	if e != nil {
		return model.BasketScripEntity{}, e
	}
	b := model.BasketScripEntity{
		Price:            s.Price,
		Qty:              s.Qty,
		OrderType:        s.OrderType,
		Product:          s.Product,
		PriceType:        s.PriceType,
		TransType:        s.TransType,
		Ret:              s.Ret,
		TriggerPrice:     s.TriggerPrice,
		DisClosedQty:     s.DisClosedQty,
		MktProtection:    s.MktProtection,
		Target:           s.Target,
		StopLoss:         s.StopLoss,
		TrailingStopLoss: s.TrailingStopLoss,
		ValidityDays:     s.ValidityDays,
		ExpiryDate:       s.ExpiryDate,
		SortOrder:        s.SortOrder,
		BasketId:         id,
		Exchange:         c.Exch,
		Token:            c.Token,
		Expiry:           c.Expiry,
		TradingSymbol:    c.TradingSymbol,
		FormattedInsName: c.FormattedInsName,
		WeekTag:          c.WeekTag,
		ActiveStatus:     1,
		CreatedBy:        ptr(user),
		CreatedOn:        model.Now(),
		UpdatedOn:        model.Now(),
	}
	if existing != nil {
		b.Id = existing.Id
		b.CreatedBy = existing.CreatedBy
		b.CreatedOn = existing.CreatedOn
		b.ActiveStatus = existing.ActiveStatus
		b.UpdatedBy = ptr(user)
		b.LotSize = s.LotSize
		if !nonempty(s.ValidityDays) {
			b.ValidityDays = existing.ValidityDays
		}
		if !nonempty(s.ExpiryDate) {
			b.ExpiryDate = existing.ExpiryDate
		}
	}
	return b, nil
}

/*
func (a *App) basket(r *http.Request, action string) (any, error) {
	ctx := r.Context()
	idn := identity(r)
	db := a.DB.WithContext(ctx)
	id, _ := strconv.ParseInt(r.PathValue("basketId"), 10, 64)
	if action == "execute" {
		return a.executeBasket(r)
	}
	if action == "get" {
		resp, e := a.basketReports(ctx, idn.UserID)
		if e == nil && len(resp.Result.([]M)) == 0 {
			//added for sonarqube
			// return message("No records found"), nil
			return message(noRecordsFound), nil
		}
		return resp, e
	}
	if action == "scrips" || action == "reset" || action == "delete" {
		if id < 1 {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		if _, e := owned(db, id, idn.UserID); errors.Is(e, gorm.ErrRecordNotFound) {
			//added for sonarqube
			// return failed("Invalid basket"), nil
			return failed(invalidBasket), nil
		} else if e != nil {
			return nil, e
		}
		switch action {
		case "scrips":
			return a.scrips(ctx, id, true)
		case "reset":
			if e := db.Model(&model.BasketNameEntity{}).Where(basketUserWhere, id, idn.UserID).Updates(M{"is_executed": "0", "updated_by": idn.UserID}).Error; e != nil {
				return nil, e
			}

		case "delete":
			if e := db.Transaction(func(tx *gorm.DB) error { return deleteBasket(tx, id, a.Config.Modules.Notifications) }); e != nil {
				return nil, e
			}
		}
		return a.basketReports(ctx, idn.UserID)
	}
	if action == "updateScripList" {
		var req *model.BasketOrderListUpdateReq
		if decode(r, &req) != nil || req == nil || req.BasketId < 1 || req.Scrips == nil {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		for _, s := range req.Scrips {
			if s.Id < 1 || !validScrip(s, false) {
				//added for sonarqube
				// return failed("Invalid Parameter"), nil
				return failed(invalidParameter), nil
			}
		}
		return a.updateScrips(ctx, int64(req.BasketId), idn.UserID, req.Scrips)
	}
	var req *model.BasketOrderReq
	if decode(r, &req) != nil || req == nil {
		//added for sonarqube
		// return failed("Invalid Parameter"), nil
		return failed(invalidParameter), nil
	}
	id = int64(req.BasketId)
	switch action {
	case "create", "rename":
		if !nonempty(req.BasketName) || (action == "rename" && id < 1) {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		var n int64
		if e := db.Model(&model.BasketNameEntity{}).Where("basket_name = ? AND user_id = ?", req.BasketName, idn.UserID).Count(&n).Error; e != nil {
			return nil, e
		}
		if n > 0 {
			return failed("Basket name already exist"), nil
		}
		if action == "create" {
			b := model.BasketNameEntity{UserId: ptr(idn.UserID), CreatedBy: ptr(idn.UserID), BasketName: req.BasketName, IsExecuted: ptr("0"), ActiveStatus: 1}
			if e := db.Create(&b).Error; e != nil {
				return nil, e
			}
		} else {
			// res := db.Model(&model.BasketNameEntity{}).Where("basket_id = ? AND user_id = ?", id, idn.UserID).Updates(M{"basket_name": req.BasketName, "updated_by": idn.UserID})
			res := db.Model(&model.BasketNameEntity{}).Where(basketUserWhere, id, idn.UserID).Updates(M{"basket_name": req.BasketName, "updated_by": idn.UserID})
			if res.Error != nil {
				return nil, res.Error
			}
			if res.RowsAffected == 0 {
				//added for sonarqube
				// return failed("Invalid basket"), nil
				return failed(invalidBasket), nil
			}
		}
		return a.basketReports(ctx, idn.UserID)
	case "addScrip":
		if id < 1 || !validScrip(req.Scrips, false) {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if _, e := owned(lockBasketRow(tx), id, idn.UserID); e != nil {
				return e
			}
			var n int64
			// if e := tx.Model(&model.BasketScripEntity{}).Where("basket_id = ?", id).Count(&n).Error; e != nil {
			if e := tx.Model(&model.BasketScripEntity{}).Where(basketIDWhere, id).Count(&n).Error; e != nil {
				return e
			}
			if n >= int64(a.Config.Business.MaxScrips) {
				return errLimit
			}
			b, e := a.prepareScrip(ctx, req.Scrips, id, idn.UserID, nil)
			if e != nil {
				return e
			}
			return tx.Create(&b).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//added for sonarqube
			// return failed("Invalid basket"), nil
			return failed(invalidBasket), nil
		}
		if errors.Is(err, errLimit) {
			return failed("Basket reached the maximum limits"), nil
		}
		if err != nil {
			return nil, err
		}
		return a.scrips(ctx, id, false)
	case "updateScrip":
		if id < 1 || req.Scrips.Id < 1 || !validScrip(req.Scrips, false) {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		return a.updateScrips(ctx, id, idn.UserID, []model.ScripRequestModel{req.Scrips})
	case "deleteScrip":
		if id < 1 || len(req.ScripsId) == 0 {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		if _, e := owned(db, id, idn.UserID); errors.Is(e, gorm.ErrRecordNotFound) {
			//added for sonarqube
			// return failed("Invalid basket"), nil
			return failed(invalidBasket), nil
		} else if e != nil {
			return nil, e
		}
		res := db.Where("basket_id = ? AND id IN ?", id, req.ScripsId).Delete(&model.BasketScripEntity{})
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			// return failed("No records found"), nil
			return failed(noRecordsFound), nil
		}
		return a.scrips(ctx, id, false)
	}
	//added for sonarqube
	// return failed("Invalid Parameter"), nil
	return failed(invalidParameter), nil
}
*/

type basketScripJoinRow struct {
	BasketId         int64            `gorm:"column:b_basket_id"`
	Id               *int64           `gorm:"column:id"`
	ScripBasketId    *int64           `gorm:"column:basket_id"`
	LotSize          *string          `gorm:"column:lot_size"`
	SortOrder        *string          `gorm:"column:sort_order"`
	Exchange         *string          `gorm:"column:exchange"`
	Token            *string          `gorm:"column:token"`
	TradingSymbol    *string          `gorm:"column:trading_symbol"`
	Qty              *string          `gorm:"column:qty"`
	Price            *string          `gorm:"column:price"`
	Expiry           *model.Timestamp `gorm:"column:expiry"`
	Product          *string          `gorm:"column:product"`
	TransType        *string          `gorm:"column:trans_type"`
	PriceType        *string          `gorm:"column:price_type"`
	OrderType        *string          `gorm:"column:order_type"`
	Ret              *string          `gorm:"column:ret"`
	TriggerPrice     *string          `gorm:"column:trigger_price"`
	DisClosedQty     *string          `gorm:"column:dis_closed_qty"`
	MktProtection    *string          `gorm:"column:mkt_protection"`
	Target           *string          `gorm:"column:target"`
	StopLoss         *string          `gorm:"column:stop_loss"`
	TrailingStopLoss *string          `gorm:"column:trailing_stop_loss"`
	FormattedInsName *string          `gorm:"column:formatted_ins_name"`
	WeekTag          *string          `gorm:"column:week_tag"`
	ValidityDays     *string          `gorm:"column:validity_days"`
	ExpiryDate       *string          `gorm:"column:expiry_date"`
}

func (a *App) getBasketScripsOptimized(ctx context.Context, id int64, userID string) (Response, error) {
	if id < 1 {
		return failed(invalidParameter), nil
	}
	var rows []basketScripJoinRow
	db := a.DB.WithContext(ctx)
	err := db.Table("tbl_basket_order b").
		Select("b.basket_id AS b_basket_id, s.id, s.basket_id, s.lot_size, s.sort_order, s.exchange, s.token, s.trading_symbol, s.qty, s.price, s.expiry, s.product, s.trans_type, s.price_type, s.order_type, s.ret, s.trigger_price, s.dis_closed_qty, s.mkt_protection, s.target, s.stop_loss, s.trailing_stop_loss, s.formatted_ins_name, s.week_tag, s.validity_days, s.expiry_date").
		Joins("LEFT JOIN tbl_basket_order_scrip s ON b.basket_id = s.basket_id").
		Where("b.basket_id = ? AND b.user_id = ?", id, userID).
		Order("s.id").
		Find(&rows).Error
	if err != nil {
		return Response{}, err
	}
	if len(rows) == 0 {
		return failed(invalidBasket), nil
	}
	if len(rows) == 1 && (rows[0].Id == nil || *rows[0].Id == 0) {
		return message(noRecordsFound), nil
	}
	out := make([]model.BasketScripEntity, len(rows))
	for i := range rows {
		out[i] = model.BasketScripEntity{
			Id:               val(rows[i].Id),
			BasketId:         val(rows[i].ScripBasketId),
			LotSize:          rows[i].LotSize,
			SortOrder:        rows[i].SortOrder,
			Exchange:         rows[i].Exchange,
			Token:            rows[i].Token,
			TradingSymbol:    rows[i].TradingSymbol,
			Qty:              rows[i].Qty,
			Price:            rows[i].Price,
			Expiry:           rows[i].Expiry,
			Product:          rows[i].Product,
			TransType:        rows[i].TransType,
			PriceType:        rows[i].PriceType,
			OrderType:        rows[i].OrderType,
			Ret:              rows[i].Ret,
			TriggerPrice:     rows[i].TriggerPrice,
			DisClosedQty:     rows[i].DisClosedQty,
			MktProtection:    rows[i].MktProtection,
			Target:           rows[i].Target,
			StopLoss:         rows[i].StopLoss,
			TrailingStopLoss: rows[i].TrailingStopLoss,
			FormattedInsName: rows[i].FormattedInsName,
			WeekTag:          rows[i].WeekTag,
			ValidityDays:     rows[i].ValidityDays,
			ExpiryDate:       rows[i].ExpiryDate,
		}
		if nonempty(out[i].Exchange) && nonempty(out[i].Token) {
			c, e := a.contract(ctx, val(out[i].Exchange), val(out[i].Token))
			if e != nil {
				return Response{}, e
			}
			if nonempty(c.LotSize) {
				out[i].LotSize = c.LotSize
			}
		}
	}
	return success(out), nil
}

// added for sonarqube
func (a *App) handleBasketScripActions(ctx context.Context, db *gorm.DB, action string, id int64, userID string) (any, error) {
	if id < 1 {
		return failed(invalidParameter), nil
	}
	if action == "scrips" {
		return a.getBasketScripsOptimized(ctx, id, userID)
	}
	if _, e := owned(db, id, userID); errors.Is(e, gorm.ErrRecordNotFound) {
		return failed(invalidBasket), nil
	} else if e != nil {
		return nil, e
	}
	switch action {
	case "reset":
		if e := db.Model(&model.BasketNameEntity{}).Where(basketUserWhere, id, userID).Updates(M{"is_executed": "0", "updated_by": userID}).Error; e != nil {
			return nil, e
		}
	case "delete":
		if e := db.Transaction(func(tx *gorm.DB) error { return deleteBasket(tx, id, a.Config.Modules.Notifications) }); e != nil {
			return nil, e
		}
	}
	return a.basketReports(ctx, userID)
}

// added for sonarqube
func (a *App) handleBasketUpdateScripList(ctx context.Context, r *http.Request, userID string) (any, error) {
	var req *model.BasketOrderListUpdateReq
	if decode(r, &req) != nil || req == nil || req.BasketId < 1 || req.Scrips == nil {
		return failed(invalidParameter), nil
	}
	for _, s := range req.Scrips {
		if s.Id < 1 || !validScrip(s, false) {
			return failed(invalidParameter), nil
		}
	}
	return a.updateScrips(ctx, int64(req.BasketId), userID, req.Scrips)
}

// added for sonarqube
func (a *App) handleBasketCreateRename(ctx context.Context, db *gorm.DB, action string, id int64, userID string, req *model.BasketOrderReq) (any, error) {
	if !nonempty(req.BasketName) || (action == "rename" && id < 1) {
		return failed(invalidParameter), nil
	}
	var n int64
	if e := db.Model(&model.BasketNameEntity{}).Where("basket_name = ? AND user_id = ?", req.BasketName, userID).Count(&n).Error; e != nil {
		return nil, e
	}
	if n > 0 {
		return failed("Basket name already exist"), nil
	}
	if action == "create" {
		b := model.BasketNameEntity{UserId: ptr(userID), CreatedBy: ptr(userID), BasketName: req.BasketName, IsExecuted: ptr("0"), ActiveStatus: 1}
		if e := db.Create(&b).Error; e != nil {
			return nil, e
		}
	} else {
		res := db.Model(&model.BasketNameEntity{}).Where(basketUserWhere, id, userID).Updates(M{"basket_name": req.BasketName, "updated_by": userID})
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			return failed(invalidBasket), nil
		}
	}
	return a.basketReports(ctx, userID)
}

// added for sonarqube
func (a *App) handleBasketAddScrip(ctx context.Context, db *gorm.DB, id int64, userID string, req *model.BasketOrderReq) (any, error) {
	if id < 1 || !validScrip(req.Scrips, false) {
		return failed(invalidParameter), nil
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if _, e := owned(lockBasketRow(tx), id, userID); e != nil {
			return e
		}
		var n int64
		if e := tx.Model(&model.BasketScripEntity{}).Where(basketIDWhere, id).Count(&n).Error; e != nil {
			return e
		}
		if n >= int64(a.Config.Business.MaxScrips) {
			return errLimit
		}
		b, e := a.prepareScrip(ctx, req.Scrips, id, userID, nil)
		if e != nil {
			return e
		}
		return tx.Create(&b).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return failed(invalidBasket), nil
	}
	if errors.Is(err, errLimit) {
		return failed("Basket reached the maximum limits"), nil
	}
	if err != nil {
		return nil, err
	}
	return a.scrips(ctx, id, false)
}

// added for sonarqube
func (a *App) handleBasketDeleteScrip(ctx context.Context, db *gorm.DB, id int64, userID string, req *model.BasketOrderReq) (any, error) {
	if id < 1 || len(req.ScripsId) == 0 {
		return failed(invalidParameter), nil
	}
	if _, e := owned(db, id, userID); errors.Is(e, gorm.ErrRecordNotFound) {
		return failed(invalidBasket), nil
	} else if e != nil {
		return nil, e
	}
	res := db.Where("basket_id = ? AND id IN ?", id, req.ScripsId).Delete(&model.BasketScripEntity{})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return failed(noRecordsFound), nil
	}
	return a.scrips(ctx, id, false)
}

// added for sonarqube
func (a *App) handleBasketScripMutations(ctx context.Context, db *gorm.DB, action string, id int64, userID string, req *model.BasketOrderReq) (any, error) {
	switch action {
	case "create", "rename":
		return a.handleBasketCreateRename(ctx, db, action, id, userID, req)
	case "addScrip":
		return a.handleBasketAddScrip(ctx, db, id, userID, req)
	case "updateScrip":
		if id < 1 || req.Scrips.Id < 1 || !validScrip(req.Scrips, false) {
			return failed(invalidParameter), nil
		}
		return a.updateScrips(ctx, id, userID, []model.ScripRequestModel{req.Scrips})
	case "deleteScrip":
		return a.handleBasketDeleteScrip(ctx, db, id, userID, req)
	}
	return failed(invalidParameter), nil
}

// added for sonarqube
func (a *App) basket(r *http.Request, action string) (any, error) {
	ctx := r.Context()
	idn := identity(r)
	db := a.DB.WithContext(ctx)
	id, _ := strconv.ParseInt(r.PathValue("basketId"), 10, 64)
	if action == "execute" {
		return a.executeBasket(r)
	}
	if action == "get" {
		resp, e := a.basketReports(ctx, idn.UserID)
		if e == nil && len(resp.Result.([]M)) == 0 {
			return message(noRecordsFound), nil
		}
		return resp, e
	}
	if action == "scrips" || action == "reset" || action == "delete" {
		return a.handleBasketScripActions(ctx, db, action, id, idn.UserID)
	}
	if action == "updateScripList" {
		return a.handleBasketUpdateScripList(ctx, r, idn.UserID)
	}
	var req *model.BasketOrderReq
	if decode(r, &req) != nil || req == nil {
		return failed(invalidParameter), nil
	}
	id = int64(req.BasketId)
	return a.handleBasketScripMutations(ctx, db, action, id, idn.UserID, req)
}

var errLimit = errors.New("scrip limit")

func (a *App) updateScrips(ctx context.Context, id int64, user string, scrips []model.ScripRequestModel) (any, error) {
	db := a.DB.WithContext(ctx)
	if _, e := owned(db, id, user); errors.Is(e, gorm.ErrRecordNotFound) {
		//added for sonarqube
		// return failed("Invalid basket"), nil
		return failed(invalidBasket), nil
	} else if e != nil {
		return nil, e
	}
	e := db.Transaction(func(tx *gorm.DB) error {
		for _, s := range scrips {
			var old model.BasketScripEntity
			if e := tx.Where("id = ? AND basket_id = ?", s.Id, id).First(&old).Error; e != nil {
				return e
			}
			b, e := a.prepareScrip(ctx, s, id, user, &old)
			if e != nil {
				return e
			}
			if e = tx.Save(&b).Error; e != nil {
				return e
			}
		}
		return nil
	})
	if errors.Is(e, gorm.ErrRecordNotFound) {
		// return message("Invalid basket"), nil
		return message(invalidBasket), nil
	}
	if e != nil {
		return nil, e
	}
	return Response{Status: "Ok", Message: "Success", Result: "Scrip updated successfully"}, nil

}
func deleteBasket(tx *gorm.DB, id int64, notifications bool) error {
	// if e := tx.Where("basket_id = ?", id).Delete(&model.BasketScripEntity{}).Error; e != nil {
	if e := tx.Where(basketIDWhere, id).Delete(&model.BasketScripEntity{}).Error; e != nil {
		return e
	}
	if notifications {
		// if e := tx.Where("basket_id = ?", strconv.FormatInt(id, 10)).Delete(&model.UserNotification{}).Error; e != nil {
		if e := tx.Where(basketIDWhere, strconv.FormatInt(id, 10)).Delete(&model.UserNotification{}).Error; e != nil {
			return e
		}
	}
	// return tx.Where("basket_id = ?", id).Delete(&model.BasketNameEntity{}).Error
	return tx.Where(basketIDWhere, id).Delete(&model.BasketNameEntity{}).Error
}

// Shared order payload: disClosedQty in Java request DTO becomes disclosedQty upstream.
func (a *App) orderRequest(scrips []model.ScripRequestModel, basketID int64) []M {
	out := []M{}
	now := a.Now().In(a.location)
	open := false
	for _, day := range a.Config.Business.MarketDays {
		if int(now.Weekday()) == day {
			open = true
		}
	}
	tm := now.Format("15:04:05")
	open = open && tm >= a.Config.Business.MarketOpen+":00" && tm <= a.Config.Business.MarketClose+":00"
	for _, s := range scrips {
		m := M{}
		for k, v := range map[string]*string{"exchange": s.Exchange, "tradingSymbol": s.TradingSymbol, "qty": s.Qty, "price": s.Price, "orderType": s.OrderType, "product": s.Product, "priceType": s.PriceType, "transType": s.TransType, "ret": s.Ret, "triggerPrice": s.TriggerPrice, "disclosedQty": s.DisClosedQty, "mktProtection": s.MktProtection, "target": s.Target, "stopLoss": s.StopLoss, "trailingStopLoss": s.TrailingStopLoss, "source": s.Source, "token": s.Token} {
			m[k] = v
		}
		m["remark"] = nil
		m["orderNo"] = nil
		if basketID > 0 {
			m["remark"] = "TBK:" + strconv.FormatInt(basketID, 10)
			if !open {
				m["orderType"] = "AMO"
			}
		}
		out = append(out, m)
	}
	return out
}
func (a *App) executeBasket(r *http.Request) (any, error) {
	var req *model.ExecuteBasketOrderReq
	if decode(r, &req) != nil || req == nil || req.BasketId < 1 || len(req.Scrips) == 0 {
		//added for sonarqube
		// return failed("Invalid Parameter"), nil
		return failed(invalidParameter), nil
	}
	for _, s := range req.Scrips {
		if !validScrip(s, true) {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
	}
	idn := identity(r)
	if strings.TrimSpace(idn.Token) == "" {
		return httpResult{401, []any{}}, nil
	}
	if _, e := owned(a.DB.WithContext(r.Context()), int64(req.BasketId), idn.UserID); errors.Is(e, gorm.ErrRecordNotFound) {
		//added for sonarqube
		// return failed("Invalid basket"), nil
		return failed(invalidBasket), nil
	} else if e != nil {
		return nil, e
	}
	var response []Response
	status, e := a.post(r.Context(), a.Config.Upstream.OrderURL, a.orderRequest(req.Scrips, 0), idn.Token, &response)
	if status == 401 {
		return httpResult{401, []any{}}, nil
	}
	if e != nil {
		return nil, e
	}
	if response == nil {
		return failed("Failed"), nil
	}
	//added for sonarqube
	// e = a.DB.WithContext(r.Context()).Model(&model.BasketNameEntity{}).Where("basket_id = ? AND user_id = ?", req.BasketId, idn.UserID).Updates(M{"is_executed": "1", "updated_by": idn.UserID}).Error
	e = a.DB.WithContext(r.Context()).Model(&model.BasketNameEntity{}).Where(basketUserWhere, req.BasketId, idn.UserID).Updates(M{"is_executed": "1", "updated_by": idn.UserID}).Error
	return message(a.Config.Business.ExecutionMessage), e
}

// GORM's generic FOR UPDATE clause is not accepted by SQL Server. Table hints
// hold the same update lock until transaction completion on that dialect.
func lockBasketRow(tx *gorm.DB) *gorm.DB {
	if tx.Dialector.Name() == "sqlserver" {
		return tx.Table("tbl_basket_order WITH (UPDLOCK, ROWLOCK)")
	}
	return tx.Clauses(clause.Locking{Strength: "UPDATE"})
}
