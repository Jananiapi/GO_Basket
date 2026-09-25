package app

import (
	"basket/internal/model"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const invalidParameter = "Invalid Parameter"

func (a *App) midnight() time.Time {
	n := a.Now().In(a.location)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, a.location)
}
func (a *App) deleteExpiredScrips(ctx context.Context) Response {
	r := a.DB.WithContext(ctx).Where("expiry < ?", a.midnight()).Delete(&model.BasketScripEntity{})
	if r.Error != nil {
		a.Log.Error("expiry cleanup", "error", r.Error)
		return failed("Failed to deleted")
	}
	return message(strconv.FormatInt(r.RowsAffected, 10) + "-Record Deleted")
}
func (a *App) deleteExpiredBaskets(ctx context.Context) Response {
	var bs []model.BasketNameEntity
	db := a.DB.WithContext(ctx)
	if e := db.Where("expiry_date < ?", a.midnight()).Find(&bs).Error; e != nil {
		a.Log.Error("expired baskets", "error", e)
		return failed("Failed")
	}
	if len(bs) == 0 {
		return failed("No records found")
	}
	if e := db.Transaction(func(tx *gorm.DB) error {
		for _, b := range bs {
			if e := deleteBasket(tx, b.BasketId, a.Config.Modules.Notifications); e != nil {
				return e
			}
		}
		return nil
	}); e != nil {
		a.Log.Error("expired basket deletion", "error", e)
		return failed("Failed")
	}
	return message("Success")
}
/*
func (a *App) admin(r *http.Request, action string) (any, error) {
	if action == "deleteExpired" {
		return a.deleteExpiredBaskets(r.Context()), nil
	}
	if !adminAuthorized(r.Header.Get("Authorization"), a.Config.Auth.AdminToken) {
		return httpResult{401, Response{}}, nil
	}
	var req *model.AdminBasketOrderReq
	if decode(r, &req) != nil || req == nil {
		return failed(invalidParameter), nil
	}
	var vendor model.VendorAppEntity
	e := a.DB.WithContext(r.Context()).Where("api_key = ?", req.ApiKey).First(&vendor).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return failed("Your not a vendor"), nil
	}
	if e != nil {
		return nil, e
	}
	if vendor.TppAuthorization != 1 {
		return failed("Your not authorized by admin"), nil
	}
	if !nonempty(req.BasketName) || req.ExpiryDate == nil || req.Scrips == nil {
		return failed(invalidParameter), nil
	}
	if len(req.Scrips) > a.Config.Business.MaxScrips {
		return failed("Scrip is more than maximum size"), nil
	}
	for _, s := range req.Scrips {
		if !validScrip(s, true) {
			return failed(invalidParameter), nil
		}
		valid := false
		for _, x := range a.Config.Business.AdminExchanges {
			valid = valid || strings.EqualFold(x, val(s.Exchange))
		}
		if !valid {
			return failed(invalidParameter), nil
		}
	}
	users := req.UserId
	if len(users) == 0 {
		if e = a.DB.WithContext(r.Context()).Model(&model.DeviceMappingEntity{}).Distinct("user_id").Pluck("user_id", &users).Error; e != nil {
			return nil, e
		}
	}
	scrips := make([]model.BasketScripEntity, 0, len(req.Scrips))
	for _, s := range req.Scrips {
		var b model.BasketScripEntity
		if e = copyJSON(s, &b); e != nil {
			return nil, e
		}
		b.Id = 0
		b.Qty = nil
		c, e := a.contract(r.Context(), val(s.Exchange), val(s.Token))
		if e != nil {
			return nil, e
		}
		if nonempty(c.LotSize) {
			qty, e := strconv.Atoi(val(s.Qty))
			if e != nil {
				return failed(invalidParameter), nil
			}
			lot, e := strconv.Atoi(val(c.LotSize))
			if e != nil {
				return nil, e
			}
			b.LotSize = c.LotSize
			b.Qty = ptr(strconv.Itoa(qty * lot))
		}
		scrips = append(scrips, b)
		if e = a.Cache.Put(r.Context(), a.Config.Cache.Maps["ideas"], val(s.Exchange)+"_"+val(s.Token), true); e != nil {
			return nil, e
		}
	}
	// Queue one campaign atomically so queue pressure cannot partially accept users.
	job := func() {
		for _, u := range users {
			if e := a.createAdminBasket(a.ctx, u, *req, scrips, action == "adminTemp"); e != nil {
				a.Log.Error("admin basket creation", "user", u, "error", e)
			}
		}
	}
	select {
	case a.jobs <- job:
		return message("Success"), nil
	default:
		return failed("Notification queue is full"), nil
	}
}
*/

// added for sonarqube
func (a *App) isAdminExchangeValid(exchange string) bool {
	for _, x := range a.Config.Business.AdminExchanges {
		if strings.EqualFold(x, exchange) {
			return true
		}
	}
	return false
}

// added for sonarqube
func (a *App) validateAdminRequest(r *http.Request, req *model.AdminBasketOrderReq) (any, error) {
	var vendor model.VendorAppEntity
	e := a.DB.WithContext(r.Context()).Where("api_key = ?", req.ApiKey).First(&vendor).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return failed("Your not a vendor"), nil
	}
	if e != nil {
		return nil, e
	}
	if vendor.TppAuthorization != 1 {
		return failed("Your not authorized by admin"), nil
	}
	if !nonempty(req.BasketName) || req.ExpiryDate == nil || req.Scrips == nil {
		return failed(invalidParameter), nil
	}
	if len(req.Scrips) > a.Config.Business.MaxScrips {
		return failed("Scrip is more than maximum size"), nil
	}
	for _, s := range req.Scrips {
		if !validScrip(s, true) || !a.isAdminExchangeValid(val(s.Exchange)) {
			return failed(invalidParameter), nil
		}
	}
	return nil, nil
}

// added for sonarqube
func (a *App) prepareAdminScrips(ctx context.Context, scripReqs []model.ScripRequestModel) ([]model.BasketScripEntity, any, error) {
	scrips := make([]model.BasketScripEntity, 0, len(scripReqs))
	for _, s := range scripReqs {
		b := model.BasketScripEntity{
			Price:            s.Price,
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
			Exchange:         s.Exchange,
			Token:            s.Token,
			TradingSymbol:    s.TradingSymbol,
		}
		c, e := a.contract(ctx, val(s.Exchange), val(s.Token))
		if e != nil {
			return nil, nil, e
		}
		if nonempty(c.LotSize) {
			qty, e := strconv.Atoi(val(s.Qty))
			if e != nil {
				return nil, failed(invalidParameter), nil
			}
			lot, e := strconv.Atoi(val(c.LotSize))
			if e != nil {
				return nil, nil, e
			}
			b.LotSize = c.LotSize
			b.Qty = ptr(strconv.Itoa(qty * lot))
		}
		scrips = append(scrips, b)
		if e := a.Cache.Put(ctx, a.Config.Cache.Maps["ideas"], val(s.Exchange)+"_"+val(s.Token), true); e != nil {
			return nil, nil, e
		}
	}
	return scrips, nil, nil
}

// added for sonarqube
func (a *App) resolveAdminUsers(ctx context.Context, users []string) ([]string, error) {
	if len(users) > 0 {
		return users, nil
	}
	if e := a.DB.WithContext(ctx).Model(&model.DeviceMappingEntity{}).Distinct("user_id").Pluck("user_id", &users).Error; e != nil {
		return nil, e
	}
	return users, nil
}

// added for sonarqube
func (a *App) queueAdminBasketJob(users []string, req model.AdminBasketOrderReq, scrips []model.BasketScripEntity, temp bool) Response {
	job := func() {
		for _, u := range users {
			if e := a.createAdminBasket(a.ctx, u, req, scrips, temp); e != nil {
				a.Log.Error("admin basket creation", "user", u, "error", e)
			}
		}
	}
	select {
	case a.jobs <- job:
		return message("Success")
	default:
		return failed("Notification queue is full")
	}
}

// added for sonarqube
func (a *App) admin(r *http.Request, action string) (any, error) {
	if action == "deleteExpired" {
		return a.deleteExpiredBaskets(r.Context()), nil
	}
	if !adminAuthorized(r.Header.Get("Authorization"), a.Config.Auth.AdminToken) {
		return httpResult{401, Response{}}, nil
	}
	var req *model.AdminBasketOrderReq
	if decode(r, &req) != nil || req == nil {
		return failed(invalidParameter), nil
	}
	if resp, err := a.validateAdminRequest(r, req); resp != nil || err != nil {
		return resp, err
	}
	users, err := a.resolveAdminUsers(r.Context(), req.UserId)
	if err != nil {
		return nil, err
	}
	scrips, resp, err := a.prepareAdminScrips(r.Context(), req.Scrips)
	if resp != nil || err != nil {
		return resp, err
	}
	return a.queueAdminBasketJob(users, *req, scrips, action == "adminTemp"), nil
}
func (a *App) createAdminBasket(ctx context.Context, user string, req model.AdminBasketOrderReq, scrips []model.BasketScripEntity, temp bool) error {
	var devices []model.DeviceMappingEntity
	db := a.DB.WithContext(ctx)
	if e := db.Where("user_id = ?", user).Find(&devices).Error; e != nil {
		return e
	}
	if e := a.Cache.Put(ctx, a.Config.Cache.Maps["devices"], user, devices); e != nil {
		return e
	}
	name := val(req.BasketName) + " " + a.Now().In(a.location).Format("02-01 15:04")
	b := model.BasketNameEntity{UserId: ptr(user), CreatedBy: ptr(user), BasketName: ptr(name), Description: req.Description, ExpiryDate: req.ExpiryDate, ResearchCall: 1, IsExecuted: ptr("0"), ActiveStatus: 1}
	//added by janani for sonarqube open issue resoving
	e := db.Transaction(func(tx *gorm.DB) error {
		return a.createAdminBasketTransaction(tx, b, user, req, scrips, devices, name)
	})
	if e != nil {
		return e
	}
	if !temp && a.Config.Modules.Notifications && req.PushNotification == 1 {
		ids := []string{}
		for _, d := range devices {
			if nonempty(d.DeviceId) {
				ids = append(ids, strings.TrimSpace(val(d.DeviceId)))
			}
		}
		return a.notify(ctx, ids, val(req.Title), val(req.Message), M{"type": "Basket", "basketId": strconv.FormatInt(b.BasketId, 10)})
	}
	return nil
}

// added by janani for sonarqube open issue resoving
func (a *App) createAdminBasketTransaction(
	tx *gorm.DB,
	b model.BasketNameEntity,
	user string,
	req model.AdminBasketOrderReq,
	scrips []model.BasketScripEntity,
	devices []model.DeviceMappingEntity,
	name string,
) error {
	if e := tx.Create(&b).Error; e != nil {
		return e
	}

	for _, s := range scrips {
		s.Id = 0
		s.BasketId = b.BasketId
		s.ActiveStatus = 1

		if e := tx.Create(&s).Error; e != nil {
			return e
		}
	}

	if a.Config.Modules.Notifications &&
		req.PushNotification == 1 && len(devices) > 0 {
		n := model.UserNotification{
			Message:             req.Message,
			Title:               ptr(name),
			MessageType:         ptr("ResearchCall"),
			UserId:              ptr(user),
			UserType:            ptr("individual"),
			BasketId:            ptr(strconv.FormatInt(b.BasketId, 10)),
			OrderRecommendation: ptr("null"),
			CreatedBy:           ptr("Admin"),
			UpdatedBy:           ptr("Admin"),
			ActiveStatus:        ptr(true),
		}

		return tx.Create(&n).Error
	}

	return nil
}
