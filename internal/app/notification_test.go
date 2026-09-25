package app

import (
	"basket/internal/model"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFCMV1Payload(t *testing.T) {
	a := testApp(t)
	sends := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			jsonWrite(w, 200, M{"access_token": "fixture-access", "expires_in": 3600, "token_type": "Bearer"})
			return
		}
		sends++
		if r.Header.Get("Authorization") != "Bearer fixture-access" {
			t.Error("FCM token header")
		}
		var payload M
		json.NewDecoder(r.Body).Decode(&payload)
		m := payload["message"].(map[string]any)
		data := m["data"].(map[string]any)
		if data["type"] != "Basket" || data["basketId"] != "12" {
			t.Error(data)
		}
		jsonWrite(w, 200, M{"name": "projects/fixture/messages/1"})
	}))
	defer srv.Close()
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		t.Fatal(e)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	file := filepath.Join(t.TempDir(), "credentials.json")
	if e = os.WriteFile(file, []byte(jsonString(M{"type": "service_account", "project_id": "fixture", "private_key_id": "fixture", "private_key": string(keyPEM), "client_email": "fixture@example.test", "client_id": "1", "token_uri": srv.URL + "/token"})), 0600); e != nil {
		t.Fatal(e)
	}
	a.Config.Notifications.CredentialsFile = file
	a.Config.Notifications.Provider = "fcm"
	a.Config.Notifications.URL = srv.URL + "/send"
	if e = a.notify(context.Background(), []string{"device1", "device2"}, "Title", "Message", M{"type": "Basket", "basketId": "12"}); e != nil {
		t.Fatal(e)
	}
	if sends != 2 {
		t.Fatal(sends)
	}
}
func TestAdminPushAndNotificationSwitch(t *testing.T) {
	a := testApp(t)
	a.DB.Create(&model.DeviceMappingEntity{UserId: ptr("USER1"), DeviceId: ptr(" device ")})
	sends := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sends++
		var m M
		json.NewDecoder(r.Body).Decode(&m)
		ids := m["registration_ids"].([]any)
		if len(ids) != 1 || ids[0] != "device" {
			t.Error(ids)
		}
		jsonWrite(w, 200, M{"success": 1, "failure": 0})
	}))
	defer srv.Close()
	a.Config.Notifications.Provider = "legacy"
	a.Config.Notifications.URL = srv.URL
	req := model.AdminBasketOrderReq{BasketName: ptr("Admin"), PushNotification: 1, Title: ptr("Title"), Message: ptr("Body")}
	if e := a.createAdminBasket(context.Background(), "USER1", req, nil, false); e != nil {
		t.Fatal(e)
	}
	if sends != 1 {
		t.Fatal(sends)
	}
	a.Config.Modules.Notifications = false
	if e := a.createAdminBasket(context.Background(), "USER1", req, nil, false); e != nil {
		t.Fatal(e)
	}
	if sends != 1 {
		t.Fatal(sends)
	}
	var n int64
	a.DB.Model(&model.UserNotification{}).Count(&n)
	if n != 1 {
		t.Fatal(n)
	}
}
