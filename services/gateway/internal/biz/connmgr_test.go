package biz_test

import (
	"testing"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

func TestConnManagerAddRemove(t *testing.T) {
	cm := biz.NewConnManager()
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
	cm.Add("u1", nil)
	if c := cm.OnlineCount(); c != 1 { t.Errorf("want 1, got %d", c) }
	_, ok := cm.GetConn("u1")
	if !ok { t.Error("u1 not found") }
	cm.Remove("u1")
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
}

func TestConnManagerGroupRouting(t *testing.T) {
	cm := biz.NewConnManager()
	cm.Add("u1", nil)
	cm.Add("u2", nil)
	cm.Add("u3", nil)
	cm.JoinGroup("g1", "u1")
	cm.JoinGroup("g1", "u2")
	cm.JoinGroup("g1", "u3")
	members := cm.GetGroupMembers("g1")
	if len(members) != 3 { t.Errorf("want 3, got %d", len(members)) }
	cm.LeaveGroup("g1", "u2")
	if len(cm.GetGroupMembers("g1")) != 2 { t.Error("u2 not removed") }
	cm.Remove("u3")
	if len(cm.GetGroupMembers("g1")) != 1 { t.Error("u3 not cleaned up") }
}

func TestConnManagerRoomRouting(t *testing.T) {
	cm := biz.NewConnManager()
	cm.Add("u1", nil)
	cm.Add("u2", nil)
	cm.JoinRoom("r1", "u1")
	cm.JoinRoom("r1", "u2")
	if len(cm.GetRoomMembers("r1")) != 2 { t.Errorf("want 2, got %d", len(cm.GetRoomMembers("r1"))) }
	cm.LeaveRoom("r1", "u2")
	if len(cm.GetRoomMembers("r1")) != 1 { t.Error("u2 not removed") }
}
