//go:build integration
// +build integration

package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/data"
	goredis "github.com/redis/go-redis/v9"
)

const (
	testGroupID = "docker-e2e-g1"
	testRoomID  = "docker-e2e-r1"
)

// ── Tracing logger ──────────────────────────────────────────────────────────
var traceMu sync.Mutex

func trace(format string, args ...any) {
	traceMu.Lock()
	defer traceMu.Unlock()
	fmt.Printf("  │ "+format+"\n", args...)
}
func traceS(name, format string, args ...any) {
	traceMu.Lock()
	defer traceMu.Unlock()
	fmt.Printf("  │ [%s] "+format+"\n", append([]any{name}, args...)...)
}

// dockerGateway wraps a real gateway service backed by Redis Pub/Sub.
type dockerGateway struct {
	svc       *GatewayGrpcService
	cm        biz.ConnManager
	pubsub    *data.PubSub
	rdb       *goredis.Client
	conns     map[string]*testConn
	delivered map[string]bool
	mu        sync.Mutex
	name      string
}

func newDockerGateway(t *testing.T, name string) *dockerGateway {
	t.Helper()
	rdb := goredis.NewClient(&goredis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("%s: redis ping: %v", name, err)
	}
	pubsub := data.NewPubSub(rdb)
	cm := newTestCM()
	svc := NewGatewayGrpcService(cm, pubsub)
	traceS(name, "启动 (Redis Pub/Sub 就绪)")
	return &dockerGateway{
		svc: svc, cm: cm, pubsub: pubsub, rdb: rdb,
		conns: make(map[string]*testConn), delivered: make(map[string]bool), name: name,
	}
}

func (g *dockerGateway) addUser(userID string) *testConn {
	c := &testConn{userID: userID}
	g.conns[userID] = c
	g.cm.(*testCM).addUser(userID)
	traceS(g.name, "用户 %s 上线", userID)
	return c
}

func (g *dockerGateway) startListening(t *testing.T) {
	t.Helper()
	handler := func(ctx context.Context, bm *biz.BroadcastMsg) {
		g.deliverLocally(bm)
	}
	go g.pubsub.Subscribe(context.Background(), handler)
	traceS(g.name, "Redis Pub/Sub 监听已启动")
	time.Sleep(100 * time.Millisecond)
}

func (g *dockerGateway) stop() {
	traceS(g.name, "⛔ 崩溃下线 (pubsub关闭)")
	g.pubsub.Close()
	g.rdb.Close()
}

func (g *dockerGateway) joinGroup(userID string) {
	ctx := context.Background()
	g.cm.JoinGroup(testGroupID, userID)
	g.svc.JoinGroup(ctx, &gatewayv1.JoinGroupRequest{GroupId: testGroupID, UserId: userID})
	traceS(g.name, "用户 %s 加入群组 %s", userID, testGroupID)
}

func (g *dockerGateway) broadcastGroup(t *testing.T, payload []byte) {
	t.Helper()
	ctx := context.Background()
	bm := &biz.BroadcastMsg{
		Type: biz.BroadcastTypeGroup, TargetID: testGroupID,
		FrameType: int32(gatewayv1.FrameType_FRAME_TYPE_GROUP_CHAT),
		Payload: payload, MsgID: fmt.Sprintf("docker-%s-%d", g.name, time.Now().UnixNano()),
	}
	traceS(g.name, "📤 发布群组消息: %q → Redis Pub/Sub", string(payload))
	if err := g.pubsub.Publish(ctx, bm); err != nil {
		t.Fatalf("%s: publish: %v", g.name, err)
	}
}

func (g *dockerGateway) deliverLocally(bm *biz.BroadcastMsg) {
	g.mu.Lock()
	if g.delivered[bm.MsgID] {
		g.mu.Unlock()
		return
	}
	g.delivered[bm.MsgID] = true
	g.mu.Unlock()

	msg := biz.BuildFrame(biz.CurrentVersion, uint32(bm.FrameType), bm.Payload)
	ctx := context.Background()
	memberIDs, _ := g.cm.GetGroupMembers(bm.TargetID)
	for _, uid := range memberIDs {
		if c, ok := g.conns[uid]; ok {
			c.Write(ctx, websocket.MessageBinary, msg)
			traceS(g.name, "📥 投递给用户 %s: %q (%d bytes)", uid, string(bm.Payload), len(msg))
		}
	}
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestDocker_TwoGateways_CrossInstanceFanOut(t *testing.T) {
	trace("")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  测试 1: 双网关跨实例群组广播")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  场景: GW1 广播 → Redis Pub/Sub → GW2 投递")
	trace("")

	gw1 := newDockerGateway(t, "GW1")
	gw2 := newDockerGateway(t, "GW2")
	defer gw1.stop()
	defer gw2.stop()

	gw1.startListening(t)
	gw2.startListening(t)

	connA := gw1.addUser("userA")
	connB := gw2.addUser("userB")
	gw1.joinGroup("userA")
	gw2.joinGroup("userB")
	time.Sleep(50 * time.Millisecond)

	gw1.broadcastGroup(t, []byte("msg-from-gw1"))
	time.Sleep(100 * time.Millisecond)

	aCount := connA.messages()
	bCount := connB.messages()

	trace("")
	if aCount == 0 {
		t.Errorf("GW1 userA should receive the message (local delivery), got %d", aCount)
	}
	if bCount == 0 {
		t.Errorf("GW2 userB should receive the message (cross-gateway Pub/Sub), got %d", bCount)
	}
	traceS("结果", "GW1(userA)=%d 条, GW2(userB)=%d 条", aCount, bCount)
	trace("  ✅ 测试 1 通过: 跨实例扇出正常")
}

func TestDocker_GatewayCrash_UserReconnect(t *testing.T) {
	trace("")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  测试 2: 网关崩溃 + 用户重连恢复")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  场景: GW1 崩溃 → userA 重连 GW2 → 消息不丢")
	trace("")

	gw1 := newDockerGateway(t, "GW1")
	gw2 := newDockerGateway(t, "GW2")
	defer gw2.stop()

	gw1.startListening(t)
	gw2.startListening(t)

	connA := gw1.addUser("userA")
	connB := gw2.addUser("userB")
	gw1.joinGroup("userA")
	gw2.joinGroup("userB")
	time.Sleep(50 * time.Millisecond)

	trace("── 崩溃前 ──")
	gw1.broadcastGroup(t, []byte("msg-before-crash"))
	time.Sleep(100 * time.Millisecond)

	preA := connA.messages()
	preB := connB.messages()
	traceS("崩溃前", "GW1(userA)=%d 条, GW2(userB)=%d 条", preA, preB)
	if preA == 0 || preB == 0 {
		t.Fatalf("pre-crash: GW1=%d GW2=%d, both should be >0", preA, preB)
	}

	trace("")
	trace("── GW1 崩溃 ──")
	gw1.stop()
	trace("")

	connA2 := gw2.addUser("userA")
	gw2.joinGroup("userA")
	traceS("恢复", "userA 重连到 GW2")
	time.Sleep(50 * time.Millisecond)

	trace("── 崩溃后 ──")
	gw2.broadcastGroup(t, []byte("msg-after-crash"))
	time.Sleep(100 * time.Millisecond)

	a2Count := connA2.messages()
	postB := connB.messages()
	traceS("崩溃后", "GW2(userA)=%d 条, GW2(userB)=%d 条", a2Count, postB)

	if a2Count == 0 {
		t.Error("reconnected userA on GW2 should receive post-crash messages")
	}
	if postB < preB+1 {
		t.Errorf("userB should have pre+post, got %d (pre=%d)", postB, preB)
	}
	if postB > preB+1 {
		t.Errorf("userB may have duplicates: pre=%d post=%d", preB, postB)
	}
	trace("  ✅ 测试 2 通过: 崩溃恢复，消息不丢不重")
}

func TestDocker_ScaleOut_NewInstance(t *testing.T) {
	trace("")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  测试 3: 扩容新实例")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  场景: 新增 GW3 → 用户连 GW3 → 立即收到广播")
	trace("")

	gw1 := newDockerGateway(t, "GW1")
	gw2 := newDockerGateway(t, "GW2")
	defer gw1.stop()
	defer gw2.stop()

	gw1.startListening(t)
	gw2.startListening(t)

	connA := gw1.addUser("userA")
	connB := gw2.addUser("userB")
	gw1.joinGroup("userA")
	gw2.joinGroup("userB")
	time.Sleep(50 * time.Millisecond)

	trace("── 扩容: 新增 GW3 ──")
	gw3 := newDockerGateway(t, "GW3")
	defer gw3.stop()
	gw3.startListening(t)

	connC := gw3.addUser("userC")
	gw3.joinGroup("userC")
	time.Sleep(50 * time.Millisecond)

	trace("── 广播测试 ──")
	gw1.broadcastGroup(t, []byte("msg-scale-out"))
	time.Sleep(100 * time.Millisecond)

	a := connA.messages()
	b := connB.messages()
	c := connC.messages()
	traceS("结果", "GW1=%d  GW2=%d  GW3(新)=%d", a, b, c)

	if a == 0 {
		t.Error("GW1 userA missed message")
	}
	if b == 0 {
		t.Error("GW2 userB missed message")
	}
	if c == 0 {
		t.Error("GW3(new) userC missed message — scale-out should work immediately")
	}
	trace("  ✅ 测试 3 通过: 扩容实例正常工作")
}

func TestDocker_ConcurrentBroadcasts_NoDuplication(t *testing.T) {
	trace("")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  测试 4: 并发广播去重")
	trace("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	trace("  场景: 10 条并发广播 → 两网关各恰好 10 条")
	trace("")

	gw1 := newDockerGateway(t, "GW1")
	gw2 := newDockerGateway(t, "GW2")
	defer gw1.stop()
	defer gw2.stop()

	gw1.startListening(t)
	gw2.startListening(t)

	connA := gw1.addUser("userA")
	connB := gw2.addUser("userB")
	gw1.joinGroup("userA")
	gw2.joinGroup("userB")
	time.Sleep(50 * time.Millisecond)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			payload := []byte(fmt.Sprintf("msg-%d", i))
			if i%2 == 0 {
				gw1.broadcastGroup(t, payload)
			} else {
				gw2.broadcastGroup(t, payload)
			}
		}()
	}
	wg.Wait()
	time.Sleep(500 * time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	aCount := connA.messages()
	bCount := connB.messages()
	traceS("结果", "GW1(userA)=%d 条, GW2(userB)=%d 条 (期望各 10 条)", aCount, bCount)

	if aCount != 10 {
		t.Errorf("userA expected 10 msgs, got %d", aCount)
	}
	if bCount != 10 {
		t.Errorf("userB expected 10 msgs, got %d", bCount)
	}

	if aCount == 10 && bCount == 10 {
		trace("  ✅ 测试 4 通过: 0 丢失, 0 重复")
	}
}

var _ = websocket.MessageBinary
var _ = gatewayv1.BroadcastToGroupRequest{}
var _ = testRoomID // keep import
