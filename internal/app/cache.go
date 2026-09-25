package app

import (
	"basket/internal/config"
	"basket/internal/javaser"
	"basket/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	hz "github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/redis/go-redis/v9"
)

// added for sonarqube
const errCacheMissingFmt = "cache entry missing: %s"

// Cache values use JSON on the language-neutral boundary. The HTTP provider can
// bridge existing Java-serialized Hazelcast values without changing publishers.
type Cache interface {
	Get(context.Context, string, string, any) error
	Put(context.Context, string, string, any) error
	Delete(context.Context, string, string) error
	Close(context.Context) error
}
type MemoryCache struct {
	mu     sync.RWMutex
	Values map[string]map[string]json.RawMessage
}

func (c *MemoryCache) Get(ctx context.Context, m, k string, v any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	b := c.Values[m][k]
	if len(b) == 0 {
		//added for sonarqube
		// return fmt.Errorf("cache entry missing: %s", m)
		return fmt.Errorf(errCacheMissingFmt, m)
	}
	return json.Unmarshal(b, v)
}
func (c *MemoryCache) Put(ctx context.Context, m, k string, v any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Values == nil {
		c.Values = map[string]map[string]json.RawMessage{}
	}
	if c.Values[m] == nil {
		c.Values[m] = map[string]json.RawMessage{}
	}
	c.Values[m][k] = b
	return nil
}
func (c *MemoryCache) Delete(ctx context.Context, m, k string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Values != nil && c.Values[m] != nil {
		delete(c.Values[m], k)
	}
	return nil
}
func (c *MemoryCache) Close(context.Context) error { return nil }

type hzCache struct{ client *hz.Client }

func (c *hzCache) Get(ctx context.Context, m, k string, v any) error {
	mp, e := c.client.GetMap(ctx, m)
	if e != nil {
		return e
	}
	raw, e := mp.Get(ctx, k)
	if e != nil {
		return e
	}
	if raw == nil {
		//added for sonarqube
		// return fmt.Errorf("cache entry missing: %s", m)
		return fmt.Errorf(errCacheMissingFmt, m)
	}
	var b []byte
	switch x := raw.(type) {
	case string:
		if _, ok := v.(*string); ok {
			b, _ = json.Marshal(x)
		} else {
			b = []byte(x)
		}
	case serialization.JSON:
		b = []byte(x)
	case []byte:
		b = x
	default:
		b, e = json.Marshal(x)
	}
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
func (c *hzCache) Put(ctx context.Context, m, k string, v any) error {
	mp, e := c.client.GetMap(ctx, m)
	if e != nil {
		return e
	}
	switch v.(type) {
	case string, bool:
	case []model.DeviceMappingEntity:
	default:
		v = jsonString(v)
	}
	return mp.Set(ctx, k, v)
}
func (c *hzCache) Delete(ctx context.Context, m, k string) error {
	mp, e := c.client.GetMap(ctx, m)
	if e != nil {
		return e
	}
	return mp.Delete(ctx, k)
}
func (c *hzCache) Close(ctx context.Context) error { return c.client.Shutdown(ctx) }

type httpCache struct {
	url, token string
	client     *http.Client
}

func (c *httpCache) request(ctx context.Context, method, m, k string, v any) error {
	u := strings.TrimRight(c.url, "/") + "/maps/" + url.PathEscape(m) + "/" + url.PathEscape(k)
	body := ""
	if method == "PUT" {
		body = jsonString(v)
	}
	r, e := http.NewRequestWithContext(ctx, method, u, strings.NewReader(body))
	if e != nil {
		return e
	}
	r.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		r.Header.Set("Authorization", "Bearer "+c.token)
	}
	res, e := c.client.Do(r)
	if e != nil {
		return e
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("cache bridge HTTP %d", res.StatusCode)
	}
	if method == "GET" {
		return json.NewDecoder(res.Body).Decode(v)
	}
	return nil
}
func (c *httpCache) Get(ctx context.Context, m, k string, v any) error {
	return c.request(ctx, "GET", m, k, v)
}
func (c *httpCache) Put(ctx context.Context, m, k string, v any) error {
	return c.request(ctx, "PUT", m, k, v)
}
func (c *httpCache) Delete(ctx context.Context, m, k string) error {
	return c.request(ctx, "DELETE", m, k, nil)
}
func (c *httpCache) Close(context.Context) error { return nil }

type redisCache struct{ client *redis.Client }

func (c *redisCache) Get(ctx context.Context, m, k string, v any) error {
	val, err := c.client.HGet(ctx, m, k).Result()
	if errors.Is(err, redis.Nil) {
		//added for sonarqube
		// return fmt.Errorf("cache entry missing: %s", m)
		return fmt.Errorf(errCacheMissingFmt, m)
	}
	if err != nil {
		return err
	}
	if s, ok := v.(*string); ok {
		*s = val
		return nil
	}
	return json.Unmarshal([]byte(val), v)
}

func (c *redisCache) Put(ctx context.Context, m, k string, v any) error {
	var data string
	switch x := v.(type) {
	case string:
		data = x
	case bool:
		data = strconv.FormatBool(x)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		data = string(b)
	}
	return c.client.HSet(ctx, m, k, data).Err()
}

func (c *redisCache) Delete(ctx context.Context, m, k string) error {
	return c.client.HDel(ctx, m, k).Err()
}

func (c *redisCache) Close(ctx context.Context) error {
	return c.client.Close()
}

/*
func NewCache(ctx context.Context, c config.Config) (Cache, error) {
	switch c.Cache.Provider {
	case "file":
		m := &MemoryCache{Values: map[string]map[string]json.RawMessage{}}
		if c.Cache.File != "" {
			b, e := os.ReadFile(c.Cache.File)
			if e != nil {
				return nil, e
			}
			if e = json.Unmarshal(b, &m.Values); e != nil {
				return nil, e
			}
		}
		return m, nil
	case "http":
		if c.Cache.URL == "" {
			return nil, fmt.Errorf("cache.url is required")
		}
		return &httpCache{c.Cache.URL, c.Cache.Token, &http.Client{Timeout: c.Upstream.Timeout}}, nil
	case "hazelcast":
		hc := hz.NewConfig()
		hc.Serialization.SetGlobalSerializer(javaCacheSerializer{})
		hc.Cluster.Name = c.Cache.Cluster
		hc.Cluster.Network.SetAddresses(c.Cache.Addresses...)
		client, e := hz.StartNewClientWithConfig(ctx, hc)
		if e != nil {
			return nil, e
		}
		return &hzCache{client}, nil
	case "redis":
		addr := c.Cache.Redis.Address
		if addr == "" {
			addr = os.Getenv("BASKET_REDIS_ADDR")
		}
		if addr == "" {
			addr = "127.0.0.1:6379"
		}
		pass := c.Cache.Redis.Password
		if pass == "" {
			pass = os.Getenv("BASKET_REDIS_PASSWORD")
		}
		db := c.Cache.Redis.DB
		client := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: pass,
			DB:       db,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			client.Close()
			return nil, fmt.Errorf("redis connection failed: %w", err)
		}
		return &redisCache{client: client}, nil
	default:
		return nil, fmt.Errorf("unsupported cache provider %q", c.Cache.Provider)
	}
}
*/

// added for sonarqube
func newFileCache(c config.Config) (Cache, error) {
	m := &MemoryCache{Values: map[string]map[string]json.RawMessage{}}
	if c.Cache.File != "" {
		b, e := os.ReadFile(c.Cache.File)
		if e != nil {
			return nil, e
		}
		if e = json.Unmarshal(b, &m.Values); e != nil {
			return nil, e
		}
	}
	return m, nil
}

// added for sonarqube
func newHTTPCache(c config.Config) (Cache, error) {
	if c.Cache.URL == "" {
		return nil, fmt.Errorf("cache.url is required")
	}
	return &httpCache{c.Cache.URL, c.Cache.Token, &http.Client{Timeout: c.Upstream.Timeout}}, nil
}

// added for sonarqube
func newHazelcastCache(ctx context.Context, c config.Config) (Cache, error) {
	hc := hz.NewConfig()
	hc.Serialization.SetGlobalSerializer(javaCacheSerializer{})
	hc.Cluster.Name = c.Cache.Cluster
	hc.Cluster.Network.SetAddresses(c.Cache.Addresses...)
	client, e := hz.StartNewClientWithConfig(ctx, hc)
	if e != nil {
		return nil, e
	}
	return &hzCache{client}, nil
}

// added for sonarqube
func newRedisCache(ctx context.Context, c config.Config) (Cache, error) {
	addr := c.Cache.Redis.Address
	if addr == "" {
		addr = os.Getenv("BASKET_REDIS_ADDR")
	}
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	pass := c.Cache.Redis.Password
	if pass == "" {
		pass = os.Getenv("BASKET_REDIS_PASSWORD")
	}
	db := c.Cache.Redis.DB
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return &redisCache{client: client}, nil
}

// added for sonarqube
func NewCache(ctx context.Context, c config.Config) (Cache, error) {
	switch c.Cache.Provider {
	case "file":
		return newFileCache(c)
	case "http":
		return newHTTPCache(c)
	case "hazelcast":
		return newHazelcastCache(ctx, c)
	case "redis":
		return newRedisCache(ctx, c)
	default:
		return nil, fmt.Errorf("unsupported cache provider %q", c.Cache.Provider)
	}
}
func (a *App) contract(ctx context.Context, exch, token string) (model.ContractMasterModel, error) {
	key := exch + "_" + token
	if v, ok := a.contracts.Load(key); ok {
		return v.(model.ContractMasterModel), nil
	}
	var c model.ContractMasterModel
	e := a.Cache.Get(ctx, a.Config.Cache.Maps["contracts"], key, &c)
	if e == nil {
		a.contracts.Store(key, c)
	}
	return c, e
}

// Hazelcast reserves -100 for Java Serializable. The global serializer registry
// accepts this type ID, allowing existing Java DTO values to be read directly.
type javaCacheSerializer struct{}

func (javaCacheSerializer) ID() int32 { return -100 }
func (javaCacheSerializer) Read(in serialization.DataInput) any {
	v, e := javaser.Decode(in.ReadByte)
	if e != nil {
		panic(e)
	}
	return v
}
func (javaCacheSerializer) Write(out serialization.DataOutput, v any) {
	d, ok := v.([]model.DeviceMappingEntity)
	if !ok {
		panic("unsupported Java cache output type")
	}
	b, e := javaser.EncodeDevices(d)
	if e != nil {
		panic(e)
	}
	for _, c := range b {
		out.WriteByte(c)
	}
}
