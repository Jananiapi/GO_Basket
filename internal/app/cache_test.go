package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMemoryCacheOperations(t *testing.T) {
	ctx := context.Background()
	c := &MemoryCache{}

	// Put
	if err := c.Put(ctx, "testMap", "key1", "hello"); err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Get
	var val string
	if err := c.Get(ctx, "testMap", "key1", &val); err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}

	// Delete
	if err := c.Delete(ctx, "testMap", "key1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Get after Delete
	if err := c.Get(ctx, "testMap", "key1", &val); err == nil {
		t.Fatalf("expected error after delete, got val=%q", val)
	}
}

// mockRedisServer provides a simple in-memory Redis server responding to RESP commands
type mockRedisServer struct {
	listener net.Listener
	data     map[string]map[string]string
	mu       sync.Mutex
	closed   chan struct{}
}

func startMockRedis(t *testing.T) *mockRedisServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := &mockRedisServer{
		listener: ln,
		data:     make(map[string]map[string]string),
		closed:   make(chan struct{}),
	}
	go s.serve()
	return s
}

func (s *mockRedisServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *mockRedisServer) Close() {
	s.listener.Close()
	<-s.closed
}

func (s *mockRedisServer) serve() {
	defer close(s.closed)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *mockRedisServer) handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		cmd, err := readRESP(reader)
		if err != nil {
			return
		}
		if len(cmd) == 0 {
			continue
		}
		s.dispatch(conn, cmd)
	}
}

func readRESP(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "*") {
		return nil, fmt.Errorf("expected *, got %q", line)
	}
	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}
	args := make([]string, count)
	for i := 0; i < count; i++ {
		strLenLine, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		strLenLine = strings.TrimSpace(strLenLine)
		if !strings.HasPrefix(strLenLine, "$") {
			return nil, fmt.Errorf("expected $, got %q", strLenLine)
		}
		strLen, err := strconv.Atoi(strLenLine[1:])
		if err != nil {
			return nil, err
		}
		buf := make([]byte, strLen+2) // data + \r\n
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		args[i] = string(buf[:strLen])
	}
	return args, nil
}

func (s *mockRedisServer) dispatch(conn net.Conn, cmd []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch strings.ToUpper(cmd[0]) {
	case "HELLO":
		conn.Write([]byte("-ERR unknown command 'HELLO'\r\n"))
	case "CLIENT":
		conn.Write([]byte("+OK\r\n"))
	case "PING":
		conn.Write([]byte("+PONG\r\n"))
	case "HSET":
		if len(cmd) < 4 {
			conn.Write([]byte("-ERR wrong number of arguments for 'hset' command\r\n"))
			return
		}
		m, k, v := cmd[1], cmd[2], cmd[3]
		if s.data[m] == nil {
			s.data[m] = make(map[string]string)
		}
		s.data[m][k] = v
		conn.Write([]byte(":1\r\n"))
	case "HGET":
		if len(cmd) < 3 {
			conn.Write([]byte("-ERR wrong number of arguments for 'hget' command\r\n"))
			return
		}
		m, k := cmd[1], cmd[2]
		if s.data[m] != nil {
			if val, ok := s.data[m][k]; ok {
				conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)))
				return
			}
		}
		conn.Write([]byte("$-1\r\n"))
	case "HDEL":
		if len(cmd) < 3 {
			conn.Write([]byte("-ERR wrong number of arguments for 'hdel' command\r\n"))
			return
		}
		m, k := cmd[1], cmd[2]
		count := 0
		if s.data[m] != nil {
			if _, ok := s.data[m][k]; ok {
				delete(s.data[m], k)
				count = 1
			}
		}
		conn.Write([]byte(fmt.Sprintf(":%d\r\n", count)))
	default:
		conn.Write([]byte("+OK\r\n"))
	}
}

func TestRedisCacheOperations(t *testing.T) {
	mock := startMockRedis(t)
	defer mock.Close()

	ctx := context.Background()
	cfg := config.Config{
		Cache: config.Cache{
			Provider: "redis",
			Redis: config.RedisConfig{
				Address: mock.Addr(),
			},
		},
	}

	cache, err := NewCache(ctx, cfg)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}
	defer cache.Close(ctx)

	// Test 1: String value Put, Get, Delete
	mapName := "sessions"
	keyName := "1234_REST_SESSION"
	sessionVal := "active-session-token-xyz"

	if err := cache.Put(ctx, mapName, keyName, sessionVal); err != nil {
		t.Fatalf("Put string failed: %v", err)
	}

	var fetchedSession string
	if err := cache.Get(ctx, mapName, keyName, &fetchedSession); err != nil {
		t.Fatalf("Get string failed: %v", err)
	}
	if fetchedSession != sessionVal {
		t.Fatalf("expected %q, got %q", sessionVal, fetchedSession)
	}

	if err := cache.Delete(ctx, mapName, keyName); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if err := cache.Get(ctx, mapName, keyName, &fetchedSession); err == nil {
		t.Fatalf("expected error for deleted key, got %q", fetchedSession)
	}

	// Test 2: Struct Put, Get, Delete
	contractKey := "NSE_2188"
	contract := model.ContractMasterModel{
		Exch:   ptr("NSE"),
		Token:  ptr("2188"),
		Symbol: ptr("GENCON-EQ"),
	}

	if err := cache.Put(ctx, "contractMaster", contractKey, contract); err != nil {
		t.Fatalf("Put struct failed: %v", err)
	}

	var fetchedContract model.ContractMasterModel
	if err := cache.Get(ctx, "contractMaster", contractKey, &fetchedContract); err != nil {
		t.Fatalf("Get struct failed: %v", err)
	}
	if val(fetchedContract.Symbol) != "GENCON-EQ" {
		t.Fatalf("expected symbol 'GENCON-EQ', got %q", val(fetchedContract.Symbol))
	}

	if err := cache.Delete(ctx, "contractMaster", contractKey); err != nil {
		t.Fatalf("Delete struct failed: %v", err)
	}

	// Test 3: Map Put, Get, Delete
	custKey := "USER123"
	custData := M{"stringPkey4": "key-456", "tomcatcount": "3"}

	if err := cache.Put(ctx, "userKeyMap", custKey, custData); err != nil {
		t.Fatalf("Put map failed: %v", err)
	}

	var fetchedCust M
	if err := cache.Get(ctx, "userKeyMap", custKey, &fetchedCust); err != nil {
		t.Fatalf("Get map failed: %v", err)
	}
	if str(fetchedCust["stringPkey4"]) != "key-456" {
		t.Fatalf("expected key-456, got %v", fetchedCust["stringPkey4"])
	}
}

func TestCacheProviderSwitching(t *testing.T) {
	ctx := context.Background()

	// Unsupported provider
	invalidCfg := config.Config{
		Cache: config.Cache{Provider: "unknown_provider"},
	}
	if _, err := NewCache(ctx, invalidCfg); err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}

	// Redis unreachable
	unreachableCfg := config.Config{
		Cache: config.Cache{
			Provider: "redis",
			Redis: config.RedisConfig{
				Address: "127.0.0.1:1", // guaranteed unreachable port
			},
		},
	}
	if _, err := NewCache(ctx, unreachableCfg); err == nil {
		t.Fatal("expected error for unreachable redis, got nil")
	}
}

func TestLiveRealRedis(t *testing.T) {
	ctx := context.Background()
	addr := "127.0.0.1:6379"
	conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
	if err != nil {
		t.Skipf("Live Redis not reachable at %s: %v", addr, err)
	}
	conn.Close()

	cfg := config.Config{
		Cache: config.Cache{
			Provider: "redis",
			Redis: config.RedisConfig{
				Address: addr,
			},
		},
	}
	cache, err := NewCache(ctx, cfg)
	if err != nil {
		t.Fatalf("NewCache with live Redis failed: %v", err)
	}
	defer cache.Close(ctx)

	testKey := "live_test_key"
	testVal := "live_test_val_12345"
	if err := cache.Put(ctx, "test_hash", testKey, testVal); err != nil {
		t.Fatalf("Put failed on live Redis: %v", err)
	}
	var out string
	if err := cache.Get(ctx, "test_hash", testKey, &out); err != nil {
		t.Fatalf("Get failed on live Redis: %v", err)
	}
	if out != testVal {
		t.Fatalf("expected %q, got %q", testVal, out)
	}
	if err := cache.Delete(ctx, "test_hash", testKey); err != nil {
		t.Fatalf("Delete failed on live Redis: %v", err)
	}
	if err := cache.Get(ctx, "test_hash", testKey, &out); err == nil {
		t.Fatalf("expected error after delete, got %q", out)
	}
}

func TestLiveHazelcast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := "127.0.0.1:5701"
	conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
	if err != nil {
		t.Skipf("Hazelcast not reachable at %s: %v", addr, err)
	}
	conn.Close()

	cfg := config.Config{
		Cache: config.Cache{
			Provider:  "hazelcast",
			Addresses: []string{addr},
		},
	}
	cache, err := NewCache(ctx, cfg)
	if err != nil {
		t.Fatalf("NewCache with Hazelcast failed: %v", err)
	}
	defer cache.Close(ctx)

	testKey := "hz_live_key"
	testVal := "hz_live_val_98765"
	if err := cache.Put(ctx, "test_hz_map", testKey, testVal); err != nil {
		t.Fatalf("Hazelcast Put failed: %v", err)
	}
	var out string
	if err := cache.Get(ctx, "test_hz_map", testKey, &out); err != nil {
		t.Fatalf("Hazelcast Get failed: %v", err)
	}
	if out != testVal {
		t.Fatalf("expected %q, got %q", testVal, out)
	}
	if err := cache.Delete(ctx, "test_hz_map", testKey); err != nil {
		t.Fatalf("Hazelcast Delete failed: %v", err)
	}
	if err := cache.Get(ctx, "test_hz_map", testKey, &out); err == nil {
		t.Fatalf("expected error after delete, got %q", out)
	}
	t.Log("Hazelcast client Put/Get/Delete tested successfully")
}

func TestCacheProviderSwitchingEndToEnd(t *testing.T) {
	hzConn, err := net.DialTimeout("tcp", "127.0.0.1:5701", 200*time.Millisecond)
	if err != nil {
		t.Skipf("Hazelcast not reachable at 127.0.0.1:5701: %v", err)
	}
	hzConn.Close()

	redisConn, err := net.DialTimeout("tcp", "127.0.0.1:6379", 200*time.Millisecond)
	if err != nil {
		t.Skipf("Live Redis not reachable at 127.0.0.1:6379: %v", err)
	}
	redisConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Verify Hazelcast ON, Redis OFF
	hzCfg := config.Config{
		Cache: config.Cache{
			Provider:  "hazelcast",
			Addresses: []string{"127.0.0.1:5701"},
			Redis: config.RedisConfig{
				Address: "127.0.0.1:1", // Guaranteed unreachable port; must NOT be dialed
			},
		},
	}
	hzInstance, err := NewCache(ctx, hzCfg)
	if err != nil {
		t.Fatalf("Hazelcast initialization failed: %v", err)
	}
	if _, ok := hzInstance.(*hzCache); !ok {
		t.Fatalf("expected *hzCache, got %T", hzInstance)
	}
	if err := hzInstance.Put(ctx, "contractMaster", "TEST_HZ_KEY", "VAL_HZ"); err != nil {
		t.Fatalf("Hazelcast Put failed: %v", err)
	}
	var hzOut string
	if err := hzInstance.Get(ctx, "contractMaster", "TEST_HZ_KEY", &hzOut); err != nil || hzOut != "VAL_HZ" {
		t.Fatalf("Hazelcast Get failed, got %q, err: %v", hzOut, err)
	}
	_ = hzInstance.Delete(ctx, "contractMaster", "TEST_HZ_KEY")
	hzInstance.Close(ctx)

	// 2. Verify Redis ON, Hazelcast OFF
	redisCfg := config.Config{
		Cache: config.Cache{
			Provider:  "redis",
			Addresses: []string{"127.0.0.1:1"}, // Guaranteed unreachable port; must NOT be dialed
			Redis: config.RedisConfig{
				Address: "127.0.0.1:6379",
			},
		},
	}
	redisInstance, err := NewCache(ctx, redisCfg)
	if err != nil {
		t.Fatalf("Redis initialization failed: %v", err)
	}
	rc, ok := redisInstance.(*redisCache)
	if !ok {
		t.Fatalf("expected *redisCache, got %T", redisInstance)
	}
	if err := redisInstance.Put(ctx, "contractMaster", "TEST_REDIS_KEY", "VAL_REDIS"); err != nil {
		t.Fatalf("Redis Put failed: %v", err)
	}
	var redisOut string
	if err := redisInstance.Get(ctx, "contractMaster", "TEST_REDIS_KEY", &redisOut); err != nil || redisOut != "VAL_REDIS" {
		t.Fatalf("Redis Get failed, got %q, err: %v", redisOut, err)
	}
	// Verify Redis Hash: key=contractMaster, field=TEST_REDIS_KEY
	rawHashVal, err := rc.client.HGet(ctx, "contractMaster", "TEST_REDIS_KEY").Result()
	if err != nil || rawHashVal != "VAL_REDIS" {
		t.Fatalf("Redis raw HGet failed, got %q, err: %v", rawHashVal, err)
	}
	_ = redisInstance.Delete(ctx, "contractMaster", "TEST_REDIS_KEY")
	redisInstance.Close(ctx)
}

func TestHttpCacheAndNewCacheBranches(t *testing.T) {
	ctx := context.Background()
	data := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/maps/"), "/")
		if len(parts) < 2 {
			w.WriteHeader(400)
			return
		}
		k := parts[0] + ":" + parts[1]
		switch r.Method {
		case "GET":
			v, ok := data[k]
			if !ok {
				w.WriteHeader(404)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(v)
		case "PUT":
			var val any
			_ = json.NewDecoder(r.Body).Decode(&val)
			data[k] = fmt.Sprint(val)
			w.WriteHeader(200)
		case "DELETE":
			delete(data, k)
			w.WriteHeader(200)
		}
	}))
	defer srv.Close()

	c := &httpCache{url: srv.URL, token: "test-tok", client: srv.Client()}
	_ = c.Put(ctx, "map1", "key1", "val1")
	var out string
	_ = c.Get(ctx, "map1", "key1", &out)
	_ = c.Delete(ctx, "map1", "key1")
	_ = c.Close(ctx)

	// NewCache branches
	_, err := NewCache(ctx, config.Config{Cache: config.Cache{Provider: "http", URL: ""}})
	if err == nil {
		t.Fatal("expected error for empty http cache url")
	}

	cHttp, err := NewCache(ctx, config.Config{Cache: config.Cache{Provider: "http", URL: srv.URL}})
	if err != nil || cHttp == nil {
		t.Fatalf("expected valid http cache, got err: %v", err)
	}

	_, err = NewCache(ctx, config.Config{Cache: config.Cache{Provider: "invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid cache provider")
	}

	_, err = NewCache(ctx, config.Config{Cache: config.Cache{Provider: "file", File: "non-existent-file.json"}})
	if err == nil {
		t.Fatal("expected error for non-existent cache file")
	}

	// File cache with valid JSON file
	tmpValid, err := os.CreateTemp("", "cache-valid-*.json")
	if err == nil {
		defer os.Remove(tmpValid.Name())
		_, _ = tmpValid.WriteString(`{"contracts":{"NSE_2188":{"Token":"2188","LotSize":"1"}}}`)
		_ = tmpValid.Close()
		fCache, err := NewCache(ctx, config.Config{Cache: config.Cache{Provider: "file", File: tmpValid.Name()}})
		if err != nil || fCache == nil {
			t.Fatalf("expected valid file cache, got %v", err)
		}
		var val map[string]string
		_ = fCache.Get(ctx, "contracts", "NSE_2188", &val)
	}

	// File cache with invalid JSON content
	tmpInvalid, err := os.CreateTemp("", "cache-invalid-*.json")
	if err == nil {
		defer os.Remove(tmpInvalid.Name())
		_, _ = tmpInvalid.WriteString(`{invalid-json`)
		_ = tmpInvalid.Close()
		_, err = NewCache(ctx, config.Config{Cache: config.Cache{Provider: "file", File: tmpInvalid.Name()}})
		if err == nil {
			t.Fatal("expected error for invalid JSON content in cache file")
		}
	}

	// Redis provider branch via mock redis server
	rSrv := startMockRedis(t)
	defer rSrv.Close()
	rCache, err := NewCache(ctx, config.Config{Cache: config.Cache{Provider: "redis", Redis: config.RedisConfig{Address: rSrv.Addr()}}})
	if err != nil || rCache == nil {
		t.Fatalf("expected valid redis cache from mock, got %v", err)
	}
	defer rCache.Close(ctx)

	// Redis Put and Get with string
	if err := rCache.Put(ctx, "rMap", "sKey", "sample"); err != nil {
		t.Fatalf("rCache.Put string failed: %v", err)
	}
	var sVal string
	if err := rCache.Get(ctx, "rMap", "sKey", &sVal); err != nil || sVal != "sample" {
		t.Fatalf("rCache.Get string failed: val=%s, err=%v", sVal, err)
	}

	// Redis Put with bool
	_ = rCache.Put(ctx, "rMap", "bKey", true)

	// Redis Put and Get with struct/map
	_ = rCache.Put(ctx, "rMap", "mKey", map[string]int{"num": 42})
	var mVal map[string]int
	_ = rCache.Get(ctx, "rMap", "mKey", &mVal)

	// Redis Put with unmarshallable type (channel)
	_ = rCache.Put(ctx, "rMap", "badKey", make(chan int))

	// Redis Delete
	_ = rCache.Delete(ctx, "rMap", "sKey")

	// Redis Get non-existent
	_ = rCache.Get(ctx, "rMap", "missing", &sVal)

	// javaCacheSerializer ID
	ser := javaCacheSerializer{}
	if ser.ID() != -100 {
		t.Fatalf("expected ID -100, got %d", ser.ID())
	}
}
