package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"context"
	"os"
	"testing"
)

func TestHazelcastJavaIntegration(t *testing.T) {
	if os.Getenv("BASKET_TEST_HAZELCAST") != "1" {
		t.Skip("start CacheFixture and set BASKET_TEST_HAZELCAST=1")
	}
	c, _ := config.Load("../../config.yaml")
	c.Cache.Cluster = "basket-go-parity"
	c.Cache.Addresses = []string{"127.0.0.1:15701"}
	cache, e := NewCache(context.Background(), c)
	if e != nil {
		t.Fatal(e)
	}
	defer cache.Close(context.Background())
	var contract model.ContractMasterModel
	if e = cache.Get(context.Background(), "contractMaster", "NFO_47310", &contract); e != nil {
		t.Fatal(e)
	}
	if val(contract.LotSize) != "65" || contract.Expiry.Time().UnixMilli() != 1789410600000 {
		t.Fatal("contract differs")
	}
	var customer M
	if e = cache.Get(context.Background(), "userKeyMap", "USER1", &customer); e != nil {
		t.Fatal(e)
	}
	if customer["stringPkey4"] != "fixture-key" {
		t.Fatal("customer differs")
	}
	if e = cache.Put(context.Background(), "deviceMappingDetails", "USER1", []model.DeviceMappingEntity{{CommonEntity: model.CommonEntity{Id: 7, ActiveStatus: 1}, UserId: ptr("USER1")}}); e != nil {
		t.Fatal(e)
	}
}
