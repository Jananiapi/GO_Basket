package app

import (
	"basket/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// added for sonarqube
const formUrlEncodedContentType = "application/x-www-form-urlencoded"

func (a *App) post(ctx context.Context, u string, body any, token string, out any) (int, error) {
	return a.call(ctx, u, jsonString(body), "application/json", token, out)
}
func (a *App) call(ctx context.Context, u, body, contentType, token string, out any) (int, error) {
	start := a.Now()
	var responseBody string
	defer func() { a.restLog(ctx, u, body, responseBody, start) }()
	if u == "" {
		return 0, fmt.Errorf("upstream endpoint is not configured")
	}
	r, e := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(body))
	if e != nil {
		return 0, e
	}
	r.Header.Set("Content-Type", contentType)
	r.Header.Set("Accept", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	res, e := a.HTTP.Do(r)
	if e != nil {
		return 0, e
	}
	defer res.Body.Close()
	b, e := io.ReadAll(io.LimitReader(res.Body, a.Config.Upstream.MaxResponseBytes+1))
	if e != nil {
		return res.StatusCode, e
	}
	responseBody = string(b)
	if int64(len(b)) > a.Config.Upstream.MaxResponseBytes {
		return res.StatusCode, fmt.Errorf("upstream response exceeds limit")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return res.StatusCode, fmt.Errorf("upstream HTTP %d", res.StatusCode)
	}
	if out != nil {
		e = json.Unmarshal(b, out)
	}
	return res.StatusCode, e
}

// added for sonarqube
func (a *App) nestSpanMargin(r *http.Request, id Identity) (any, error) {
	ctx := r.Context()
	var req []model.BasketMarginRequest
	if decode(r, &req) != nil || len(req) == 0 {
		//added for sonarqube
		// return failed("Invalid Parameter"), nil
		return failed(invalidParameter), nil
	}
	for _, s := range req {
		if !nonempty(s.Exchange) || !nonempty(s.Qty) {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
	}
	var customer M
	if e := a.Cache.Get(ctx, a.Config.Cache.Maps["customers"], id.UserID, &customer); e != nil {
		return nil, e
	}
	legs := []M{}
	for i, s := range req {
		if a.Config.Business.LegacyNestFirstOnly && i > 0 {
			break
		}
		legs = append(legs, M{"exch": s.Exchange, "symbol": s.Symbol, "netQty": s.Qty, "buyQty": "", "sellQty": ""})
	}
	u, e := url.Parse(a.Config.Upstream.BasketMarginURL)
	if e != nil {
		return nil, e
	}
	q := u.Query()
	q.Set("jData", jsonString(legs))
	q.Set("jKey", str(customer["stringPkey4"]))
	q.Set("jsessionid", "."+str(customer["tomcatcount"]))
	u.RawQuery = q.Encode()
	var resp M
	status, e := a.call(ctx, u.String(), "[]", formUrlEncodedContentType, "", &resp)
	if status == 401 {
		return httpResult{401, Response{}}, nil
	}
	if e != nil {
		return nil, e
	}
	if strings.EqualFold(str(resp["stat"]), "Ok") {
		return success(M{"span": resp["spanRequirement"]}), nil
	}
	return failed(str(resp["Emsg"])), nil
}

/*
func (a *App) margin(r *http.Request, action string) (any, error) {
	ctx := r.Context()
	id := identity(r)
	if action == "nestSpan" {
		return a.nestSpanMargin(r, id)
	}
	var req []model.SpanMarginReq
	if decode(r, &req) != nil || req == nil {
		//added for sonarqube
		// return failed("Invalid Parameter"), nil
		return failed(invalidParameter), nil
	}
	var session string
	if e := a.Cache.Get(ctx, a.Config.Cache.Maps["sessions"], id.UserID+"_REST_SESSION", &session); e != nil || session == "" {
		return httpResult{401, Response{}}, nil
	}
	total := decimal.Zero
	pos := []M{}
	divisor, e := decimal.NewFromString(a.Config.Business.EquityMarginDivisor)
	if e != nil || divisor.IsZero() {
		return nil, fmt.Errorf("invalid equity margin divisor")
	}
	for _, s := range req {
		if !nonempty(s.Exchange) || !nonempty(s.Token) || !nonempty(s.Qty) || !nonempty(s.Price) || !nonempty(s.TransType) {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		qty, e := decimal.NewFromString(val(s.Qty))
		if e != nil {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		price, e := decimal.NewFromString(val(s.Price))
		if e != nil {
			//added for sonarqube
			// return failed("Invalid Parameter"), nil
			return failed(invalidParameter), nil
		}
		exch := strings.ToUpper(val(s.Exchange))
		if exch == "NSE" || exch == "BSE" {
			total = total.Add(qty.Mul(price).Div(divisor))
			continue
		}
		c, e := a.contract(ctx, exch, val(s.Token))
		if e != nil {
			return nil, e
		}
		if exch == "MCX" {
			qty = qty.Mul(dec(c.LotSize))
		}
		p := M{"prd": a.Config.Business.SpanProduct, "exch": exch, "exd": nil, "strprc": c.StrikePrice, "instname": c.InsType, "symname": c.Symbol, "buyqty": nil, "sellqty": nil, "netqty": nil, "optt": "XX"}
		if c.Expiry != nil {
			p["exd"] = strings.ToUpper(c.Expiry.Time().In(a.location).Format("02-Jan-2006"))
		}
		ins := val(c.InsType)
		if strings.HasPrefix(ins, "OPT") || strings.HasSuffix(ins, "OPT") {
			p["optt"] = c.OptionType
		}
		switch strings.ToUpper(val(s.TransType)) {
		case "B", "BUY":
			p["buyqty"] = qty.String()
		case "S", "SELL":
			p["sellqty"] = qty.String()
		}
		pos = append(pos, p)
	}
	if len(pos) > 0 {
		var resp M
		body := "jData=" + jsonString(M{"actid": a.Config.Business.SpanAccount, "pos": pos})
		//added for sonarqube
		// _, e := a.call(ctx, a.Config.Upstream.SpanURL, body, "application/x-www-form-urlencoded", "", &resp)
		_, e := a.call(ctx, a.Config.Upstream.SpanURL, body, formUrlEncodedContentType, "", &resp)
		if e == nil && strings.EqualFold(str(resp["stat"]), "Ok") {
			total = total.Add(dec(resp["span_trade"])).Add(dec(resp["expo_trade"]))
		} else {
			a.Log.Error("span margin upstream failed", "error", e)
		}
	}
	return success(M{"span": total.RoundBank(2).StringFixed(2)}), nil
}
*/

// added for sonarqube
func isValidMarginScrip(s model.SpanMarginReq) bool {
	return nonempty(s.Exchange) && nonempty(s.Token) && nonempty(s.Qty) && nonempty(s.Price) && nonempty(s.TransType)
}

// added for sonarqube
func prepareSingleSpanPosition(c model.ContractMasterModel, exch string, qty decimal.Decimal, transType string, loc *time.Location, spanProduct string) M {
	p := M{"prd": spanProduct, "exch": exch, "exd": nil, "strprc": c.StrikePrice, "instname": c.InsType, "symname": c.Symbol, "buyqty": nil, "sellqty": nil, "netqty": nil, "optt": "XX"}
	if c.Expiry != nil {
		p["exd"] = strings.ToUpper(c.Expiry.Time().In(loc).Format("02-Jan-2006"))
	}
	ins := val(c.InsType)
	if strings.HasPrefix(ins, "OPT") || strings.HasSuffix(ins, "OPT") {
		p["optt"] = c.OptionType
	}
	switch strings.ToUpper(transType) {
	case "B", "BUY":
		p["buyqty"] = qty.String()
	case "S", "SELL":
		p["sellqty"] = qty.String()
	}
	return p
}

// added for sonarqube
func (a *App) prepareMarginItems(ctx context.Context, req []model.SpanMarginReq, divisor decimal.Decimal) (decimal.Decimal, []M, any, error) {
	total := decimal.Zero
	pos := []M{}
	for _, s := range req {
		if !isValidMarginScrip(s) {
			return decimal.Zero, nil, failed(invalidParameter), nil
		}
		qty, e := decimal.NewFromString(val(s.Qty))
		if e != nil {
			return decimal.Zero, nil, failed(invalidParameter), nil
		}
		price, e := decimal.NewFromString(val(s.Price))
		if e != nil {
			return decimal.Zero, nil, failed(invalidParameter), nil
		}
		exch := strings.ToUpper(val(s.Exchange))
		if exch == "NSE" || exch == "BSE" {
			total = total.Add(qty.Mul(price).Div(divisor))
			continue
		}
		c, e := a.contract(ctx, exch, val(s.Token))
		if e != nil {
			return decimal.Zero, nil, nil, e
		}
		if exch == "MCX" {
			qty = qty.Mul(dec(c.LotSize))
		}
		pos = append(pos, prepareSingleSpanPosition(c, exch, qty, val(s.TransType), a.location, a.Config.Business.SpanProduct))
	}
	return total, pos, nil, nil
}

// added for sonarqube
func (a *App) margin(r *http.Request, action string) (any, error) {
	ctx := r.Context()
	id := identity(r)
	if action == "nestSpan" {
		return a.nestSpanMargin(r, id)
	}
	var req []model.SpanMarginReq
	if decode(r, &req) != nil || req == nil {
		return failed(invalidParameter), nil
	}
	var session string
	if e := a.Cache.Get(ctx, a.Config.Cache.Maps["sessions"], id.UserID+"_REST_SESSION", &session); e != nil || session == "" {
		return httpResult{401, Response{}}, nil
	}
	divisor, e := decimal.NewFromString(a.Config.Business.EquityMarginDivisor)
	if e != nil || divisor.IsZero() {
		return nil, fmt.Errorf("invalid equity margin divisor")
	}
	total, pos, resp, err := a.prepareMarginItems(ctx, req, divisor)
	if resp != nil || err != nil {
		return resp, err
	}
	if len(pos) > 0 {
		var resp M
		body := "jData=" + jsonString(M{"actid": a.Config.Business.SpanAccount, "pos": pos})
		_, e := a.call(ctx, a.Config.Upstream.SpanURL, body, formUrlEncodedContentType, "", &resp)
		if e == nil && strings.EqualFold(str(resp["stat"]), "Ok") {
			total = total.Add(dec(resp["span_trade"])).Add(dec(resp["expo_trade"]))
		} else {
			a.Log.Error("span margin upstream failed", "error", e)
		}
	}
	return success(M{"span": total.RoundBank(2).StringFixed(2)}), nil
}
