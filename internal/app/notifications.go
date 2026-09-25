package app

import (
	"basket/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strings"
)

/*
func (a *App) notify(ctx context.Context, devices []string, title, body string, data M) error {
	if len(devices) == 0 || !a.Config.Modules.Notifications {
		return nil
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, a.HTTP)
	c := a.Config.Notifications
	if c.Provider == "legacy" {
		for start := 0; start < len(devices); start += c.BatchSize {
			end := min(start+c.BatchSize, len(devices))
			payload := M{"registration_ids": devices[start:end], "notification": M{"title": title, "body": body}, "data": data}
			req, e := http.NewRequestWithContext(ctx, "POST", c.URL, strings.NewReader(jsonString(payload)))
			if e != nil {
				return e
			}
			req.Header.Set("Authorization", "key="+c.APIKey)
			req.Header.Set("Content-Type", "application/json")
			resp, e := a.HTTP.Do(req)
			if e != nil {
				return e
			}
			var result M
			e = json.NewDecoder(resp.Body).Decode(&result)
			resp.Body.Close()
			if e != nil {
				return e
			}
			if resp.StatusCode >= 300 || integer(result["failure"]) > 0 {
				return fmt.Errorf("notification batch failed: HTTP %d, failures %d", resp.StatusCode, integer(result["failure"]))
			}
		}
		return nil
	}
	if c.Provider != "fcm" {
		return fmt.Errorf("unsupported notification provider %q", c.Provider)
	}
	credentials, e := os.ReadFile(c.CredentialsFile)
	if e != nil {
		return e
	}
	creds, e := google.CredentialsFromJSON(ctx, credentials, "https://www.googleapis.com/auth/firebase.messaging")
	if e != nil {
		return e
	}
	token, e := creds.TokenSource.Token()
	if e != nil {
		return e
	}
	project := c.ProjectID
	if project == "" {
		project = creds.ProjectID
	}
	u := c.URL
	if u == "" {
		u = "https://fcm.googleapis.com/v1/projects/" + project + "/messages:send"
	}
	var firstErr error
	for start := 0; start < len(devices); start += c.BatchSize {
		for _, device := range devices[start:min(start+c.BatchSize, len(devices))] {
			payload := M{"message": M{"token": device, "notification": M{"title": title, "body": body}, "data": data}}
			var out M
			if _, e := a.post(ctx, u, payload, token.AccessToken, &out); e != nil {
				a.Log.Error("notification delivery failed", "error", e)
				if firstErr == nil {
					firstErr = e
				}
			}
		}
	}
	return firstErr
}
*/

// added for sonarqube
func (a *App) notifyLegacy(ctx context.Context, devices []string, title, body string, data M) error {
	c := a.Config.Notifications
	for start := 0; start < len(devices); start += c.BatchSize {
		end := min(start+c.BatchSize, len(devices))
		payload := M{"registration_ids": devices[start:end], "notification": M{"title": title, "body": body}, "data": data}
		req, e := http.NewRequestWithContext(ctx, "POST", c.URL, strings.NewReader(jsonString(payload)))
		if e != nil {
			return e
		}
		req.Header.Set("Authorization", "key="+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		resp, e := a.HTTP.Do(req)
		if e != nil {
			return e
		}
		var result M
		e = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if e != nil {
			return e
		}
		if resp.StatusCode >= 300 || integer(result["failure"]) > 0 {
			return fmt.Errorf("notification batch failed: HTTP %d, failures %d", resp.StatusCode, integer(result["failure"]))
		}
	}
	return nil
}

// added for sonarqube
func (a *App) notifyFCM(ctx context.Context, devices []string, title, body string, data M) error {
	c := a.Config.Notifications
	credentials, e := os.ReadFile(c.CredentialsFile)
	if e != nil {
		return e
	}
	creds, e := google.CredentialsFromJSON(ctx, credentials, "https://www.googleapis.com/auth/firebase.messaging")
	if e != nil {
		return e
	}
	token, e := creds.TokenSource.Token()
	if e != nil {
		return e
	}
	project := c.ProjectID
	if project == "" {
		project = creds.ProjectID
	}
	u := c.URL
	if u == "" {
		u = "https://fcm.googleapis.com/v1/projects/" + project + "/messages:send"
	}
	var firstErr error
	for start := 0; start < len(devices); start += c.BatchSize {
		for _, device := range devices[start:min(start+c.BatchSize, len(devices))] {
			payload := M{"message": M{"token": device, "notification": M{"title": title, "body": body}, "data": data}}
			var out M
			if _, e := a.post(ctx, u, payload, token.AccessToken, &out); e != nil {
				a.Log.Error("notification delivery failed", "error", e)
				if firstErr == nil {
					firstErr = e
				}
			}
		}
	}
	return firstErr
}

// added for sonarqube
func (a *App) notify(ctx context.Context, devices []string, title, body string, data M) error {
	if len(devices) == 0 || !a.Config.Modules.Notifications {
		return nil
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, a.HTTP)
	c := a.Config.Notifications
	switch c.Provider {
	case "legacy":
		return a.notifyLegacy(ctx, devices, title, body, data)
	case "fcm":
		return a.notifyFCM(ctx, devices, title, body, data)
	default:
		return fmt.Errorf("unsupported notification provider %q", c.Provider)
	}
}

// SaveNotification and NotificationList port the non-HTTP notification service.
func (a *App) SaveNotification(ctx context.Context, req model.SendNoficationReqModel) error {
	if !a.Config.Modules.Notifications {
		return fmt.Errorf("notifications module disabled")
	}
	users := req.UserId
	if strings.EqualFold(val(req.UserType), "all") {
		users = []string{"all"}
	}
	return a.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, u := range users {
			n := model.UserNotification{Message: req.Message, Url: req.Url, Title: req.Title, MessageType: req.MessageType, UserId: ptr(u), UserType: req.UserType, Icon: req.Icon, Validity: req.Validity, OrderRecommendation: ptr(jsonString(req.OrderRecommendation)), BasketId: req.BasketId, CreatedBy: ptr("Admin"), UpdatedBy: ptr("Admin"), ActiveStatus: ptr(true)}
			if e := tx.Create(&n).Error; e != nil {
				return e
			}
		}
		return nil
	})
}
func (a *App) NotificationList(ctx context.Context, user string) ([]model.UserNotification, error) {
	out := []model.UserNotification{}
	e := a.DB.WithContext(ctx).Where("user_id = ? OR (UPPER(user_type) = ? AND active_status = ?)", user, "ALL", true).Find(&out).Error
	return out, e
}
