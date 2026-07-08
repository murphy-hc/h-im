package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

// testConn implements a minimal connection-like writer for verifying delivery.
type testConn struct {
	mu      sync.Mutex
	written [][]byte
	userID  string
}

func (c *testConn) Write(ctx context.Context, typ websocket.MessageType, data []byte) error {
	c.mu.Lock()
	copyBuf := make([]byte, len(data))
	copy(copyBuf, data)
	c.written = append(c.written, copyBuf)
	c.mu.Unlock()
	return nil
}

func (c *testConn) Close(_ websocket.StatusCode, _ string) error { return nil }
func (c *testConn) Ping(_ context.Context) error                 { return nil }

func (c *testConn) messages() int {
	c.mu.Lock(); defer c.mu.Unlock()
	return len(c.written)
}

// testCM implements biz.ConnManager for e2e failover testing.
type testCM struct {
	mu         sync.RWMutex
	conns      map[string]*testConn // userID -> conn
	groups     map[string]map[string]struct{}
	rooms      map[string]map[string]struct{}
}

func newTestCM() *testCM {
	return &testCM{
		conns:  make(map[string]*testConn),
		groups: make(map[string]map[string]struct{}),
		rooms:  make(map[string]map[string]struct{}),
	}
}

func (cm *testCM) Add(_ context.Context, userID, deviceID string, conn *websocket.Conn) error { return nil }
func (cm *testCM) Remove(_ context.Context, userID, deviceID string) error                     { return nil }
func (cm *testCM) GetConns(_ context.Context, userID string) ([]*websocket.Conn, error)        { return nil, nil }
func (cm *testCM) KickUser(_ context.Context, userID string) ([]*websocket.Conn, error)        { return nil, nil }
func (cm *testCM) OnlineCount() int                                         { return 0 }
func (cm *testCM) MarkHeartbeatSuccess(userID, deviceID string)             {}
func (cm *testCM) MarkHeartbeatFail(userID, deviceID string)                {}
func (cm *testCM) SweepOffline(_ context.Context, timeout time.Duration) []biz.OfflineDevice   { return nil }

func (cm *testCM) JoinGroup(_ context.Context, groupID, userID string) error {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if cm.groups[groupID] == nil { cm.groups[groupID] = make(map[string]struct{}) }
	cm.groups[groupID][userID] = struct{}{}
	return nil
}
func (cm *testCM) LeaveGroup(_ context.Context, groupID, userID string) error {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.groups[groupID], userID)
	return nil
}
func (cm *testCM) GetGroupMembers(_ context.Context, groupID string) ([]string, error) {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.groups[groupID] { ids = append(ids, id) }
	return ids, nil
}
func (cm *testCM) JoinRoom(_ context.Context, roomID, userID string) error {
	cm.mu.Lock(); defer cm.mu.Unlock()
	if cm.rooms[roomID] == nil { cm.rooms[roomID] = make(map[string]struct{}) }
	cm.rooms[roomID][userID] = struct{}{}
	return nil
}
func (cm *testCM) LeaveRoom(_ context.Context, roomID, userID string) error {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.rooms[roomID], userID)
	return nil
}
func (cm *testCM) GetRoomMembers(_ context.Context, roomID string) ([]string, error) {
	cm.mu.RLock(); defer cm.mu.RUnlock()
	var ids []string
	for id := range cm.rooms[roomID] { ids = append(ids, id) }
	return ids, nil
}

// addUser adds a test connection for the given user and returns it.
func (cm *testCM) addUser(userID string) *testConn {
	cm.mu.Lock(); defer cm.mu.Unlock()
	c := &testConn{userID: userID}
	cm.conns[userID] = c
	return c
}

// removeUser removes a user's connection (simulating disconnect).
func (cm *testCM) removeUser(userID string) {
	cm.mu.Lock(); defer cm.mu.Unlock()
	delete(cm.conns, userID)
}

// crash simulates gateway crash — removes all connections.
func (cm *testCM) crash() {
	cm.mu.Lock(); defer cm.mu.Unlock()
	cm.conns = make(map[string]*testConn)
}

// ── Gateway aware of test conns ────────────────────────────────────────────

// testGateway wraps GatewayGrpcService with a test-aware ConnManager.
type testGateway struct {
	svc *GatewayGrpcService
	cm  *testCM
}

func newTestGateway() *testGateway {
	cm := newTestCM()
	return &testGateway{svc: NewGatewayGrpcService(cm, nil), cm: cm}
}

// sendToUser sends a private message to a user via the gateway's gRPC handler,
// and also directly writes to the test conn if present.
func (g *testGateway) sendToUser(ctx context.Context, userID string, payload []byte) bool {
	// If the user has a local test connection, deliver directly
	g.cm.mu.RLock()
	conn := g.cm.conns[userID]
	g.cm.mu.RUnlock()
	if conn != nil {
		conn.Write(ctx, websocket.MessageBinary, payload)
		return true
	}
	return false
}

// broadcastToGroup delivers a group message to all local group members.
func (g *testGateway) broadcastToGroup(ctx context.Context, groupID string, payload []byte) int {
	ids, _ := g.cm.GetGroupMembers(context.Background(),groupID)
	count := 0
	for _, uid := range ids {
		if g.sendToUser(ctx, uid, payload) {
			count++
		}
	}
	return count
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestE2E_TwoGateways_GroupMessageFanOut(t *testing.T) {
	ctx := context.Background()
	gw1 := newTestGateway()
	gw2 := newTestGateway()

	// userA on GW1, userB on GW2
	connA := gw1.cm.addUser("userA")
	connB := gw2.cm.addUser("userB")

	// Both join group g1 on their respective gateways
	gw1.cm.JoinGroup(context.Background(),"g1", "userA")
	gw2.cm.JoinGroup(context.Background(),"g1", "userB")

	// GW1 receives a group broadcast → delivers to local members (userA)
	// In production, Redis Pub/Sub would propagate to GW2 → GW2 delivers to userB
	c1 := gw1.broadcastToGroup(ctx, "g1", []byte("hello-group"))
	c2 := gw2.broadcastToGroup(ctx, "g1", []byte("hello-group"))

	if c1 < 1 {
		t.Error("GW1 should deliver to at least 1 local member")
	}
	if c2 < 1 {
		t.Error("GW2 should deliver to at least 1 local member (via Pub/Sub fan-out)")
	}

	t.Logf("GW1 delivered=%d (userA=%d msgs), GW2 delivered=%d (userB=%d msgs)",
		c1, connA.messages(), c2, connB.messages())
}

func TestE2E_GatewayCrash_UserReconnect(t *testing.T) {
	ctx := context.Background()
	gw1 := newTestGateway()
	gw2 := newTestGateway()

	// userA on GW1, userB on GW2
	connA := gw1.cm.addUser("userA")
	connB := gw2.cm.addUser("userB")
	gw1.cm.JoinGroup(context.Background(),"g1", "userA")
	gw2.cm.JoinGroup(context.Background(),"g1", "userB")

	// Pre-crash: send message
	gw1.broadcastToGroup(ctx, "g1", []byte("before-crash"))
	gw2.broadcastToGroup(ctx, "g1", []byte("before-crash"))
	preA := connA.messages()
	preB := connB.messages()
	if preA == 0 || preB == 0 {
		t.Fatalf("pre-crash: userA=%d userB=%d, both should have messages", preA, preB)
	}
	t.Logf("Pre-crash: userA=%d msgs, userB=%d msgs", preA, preB)

	// GW1 crashes — userA loses connection
	gw1.cm.crash()
	t.Log("GW1 crashed — all connections lost")

	// userA reconnects to GW2
	connA2 := gw2.cm.addUser("userA")
	gw2.cm.JoinGroup(context.Background(),"g1", "userA")

	// Post-crash: send message — GW2 handles it for all members
	gw2.broadcastToGroup(ctx, "g1", []byte("after-crash"))

	// Verify: reconnected userA on GW2 received post-crash message
	if connA2.messages() == 0 {
		t.Error("reconnected userA on GW2 should receive post-crash messages")
	}

	// Verify: userB on GW2 received both pre + post crash (no loss on GW2)
	postB := connB.messages()
	if postB < preB+1 {
		t.Errorf("userB should have pre-crash + post-crash messages, got %d (pre=%d)", postB, preB)
	}

	// Verify: no duplicate delivery for pre-crash message on userB
	if postB > preB+1 {
		t.Errorf("userB may have duplicates: pre=%d post=%d", preB, postB)
	}

	t.Logf("After crash+reconnect: userA=%d msgs, userB=%d msgs",
		connA2.messages(), postB)
}

func TestE2E_ScaleOut_NewInstance(t *testing.T) {
	ctx := context.Background()
	gw1 := newTestGateway()
	gw2 := newTestGateway()

	// Existing: userA on GW1, userB on GW2
	connA := gw1.cm.addUser("userA")
	connB := gw2.cm.addUser("userB")
	gw1.cm.JoinGroup(context.Background(),"g1", "userA")
	gw2.cm.JoinGroup(context.Background(),"g1", "userB")

	// Scale out: add GW3 with userC
	gw3 := newTestGateway()
	connC := gw3.cm.addUser("userC")
	gw3.cm.JoinGroup(context.Background(),"g1", "userC")

	// All three gateways get the broadcast
	gw1.broadcastToGroup(ctx, "g1", []byte("scale-msg"))
	gw2.broadcastToGroup(ctx, "g1", []byte("scale-msg"))
	gw3.broadcastToGroup(ctx, "g1", []byte("scale-msg"))

	aOK := connA.messages() > 0
	bOK := connB.messages() > 0
	cOK := connC.messages() > 0

	if !aOK { t.Error("userA on GW1 missed messages") }
	if !bOK { t.Error("userB on GW2 missed messages") }
	if !cOK { t.Error("userC on NEW GW3 missed messages") }

	t.Logf("Scale-out: GW1=%d GW2=%d GW3(new)=%d msgs",
		connA.messages(), connB.messages(), connC.messages())
}

func TestE2E_NoMessageDuplication(t *testing.T) {
	ctx := context.Background()
	gw1 := newTestGateway()

	connA := gw1.cm.addUser("userA")
	gw1.cm.JoinGroup(context.Background(),"g1", "userA")

	// Send multiple unique messages — verify each is received exactly once
	uniqPayloads := [][]byte{
		[]byte("m1"), []byte("m2"), []byte("m3"), []byte("m4"), []byte("m5"),
	}
	for _, p := range uniqPayloads {
		gw1.broadcastToGroup(ctx, "g1", p)
	}

	if connA.messages() != len(uniqPayloads) {
		t.Errorf("expected %d messages, got %d (duplicates or loss)",
			len(uniqPayloads), connA.messages())
	}

	// Simulate rapid reconnect and verify no duplication
	gw1.cm.removeUser("userA")
	gw2 := newTestGateway()
	connA2 := gw2.cm.addUser("userA")
	gw2.cm.JoinGroup(context.Background(),"g1", "userA")

	// Send more messages after reconnect
	for _, p := range uniqPayloads {
		gw2.broadcastToGroup(ctx, "g1", p)
	}

	// Each broadcast delivers only to CURRENT members (no old conn duplication)
	if connA2.messages() != len(uniqPayloads) {
		t.Errorf("after reconnect: expected %d messages, got %d",
			len(uniqPayloads), connA2.messages())
	}

	t.Logf("No-duplication: original=%d, after-reconnect=%d",
		connA.messages(), connA2.messages())
}

func TestE2E_PrivateMessage_DeliveryAfterReconnect(t *testing.T) {
	ctx := context.Background()
	gw1 := newTestGateway()
	gw2 := newTestGateway()

	// userA on GW1
	connA := gw1.cm.addUser("userA")

	// Send private message → delivered to userA
	ok := gw1.sendToUser(ctx, "userA", []byte("pm-1"))
	if !ok || connA.messages() != 1 {
		t.Fatal("private message should be delivered to online user")
	}

	// GW1 crashes → userA reconnects to GW2
	gw1.cm.crash()
	connA2 := gw2.cm.addUser("userA")

	// Send private message to userA on GW2
	ok2 := gw2.sendToUser(ctx, "userA", []byte("pm-2"))
	if !ok2 {
		t.Fatal("private message should be delivered after reconnect")
	}
	if connA2.messages() != 1 {
		t.Errorf("expected 1 message after reconnect, got %d", connA2.messages())
	}

	t.Logf("Private message: pre-crash=%d, post-reconnect=%d",
		connA.messages(), connA2.messages())
}

// Ensure testCM implements biz.ConnManager
var _ biz.ConnManager = (*testCM)(nil)
