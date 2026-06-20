package biz_test

import (
	"testing"

	"github.com/murphy-hc/h-im/services/gateway/internal/data"
)

func TestConnManagerAddRemove(t *testing.T) {
	cm := data.NewMemoryConnManager()
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
	cm.Add("u1", "d1", nil)
	if c := cm.OnlineCount(); c != 1 { t.Errorf("want 1, got %d", c) }
	if c := len(cm.GetConns("u1")); c != 1 { t.Errorf("want 1, got %d", c) }
	cm.Remove("u1", "d1")
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
}

func TestConnManagerMultiDevice(t *testing.T) {
	cm := data.NewMemoryConnManager()
	cm.Add("u1", "phone", nil)
	cm.Add("u1", "laptop", nil)
	if c := cm.OnlineCount(); c != 2 { t.Errorf("want 2, got %d", c) }
	if c := len(cm.GetConns("u1")); c != 2 { t.Errorf("want 2, got %d", c) }
	cm.Remove("u1", "phone")
	if c := len(cm.GetConns("u1")); c != 1 { t.Errorf("want 1, got %d", c) }
}

func TestConnManagerKickUser(t *testing.T) {
	cm := data.NewMemoryConnManager()
	cm.Add("u1", "d1", nil)
	cm.Add("u1", "d2", nil)
	kicked := cm.KickUser("u1")
	if len(kicked) != 2 { t.Errorf("want 2, got %d", len(kicked)) }
	if cm.OnlineCount() != 0 { t.Errorf("want 0, got %d", cm.OnlineCount()) }
}

func TestConnManagerGroupRouting(t *testing.T) {
	cm := data.NewMemoryConnManager()
	cm.Add("u1", "d1", nil)
	cm.Add("u2", "d1", nil)
	cm.Add("u3", "d1", nil)
	cm.JoinGroup("g1", "u1")
	cm.JoinGroup("g1", "u2")
	cm.JoinGroup("g1", "u3")
	if len(cm.GetGroupMembers("g1")) != 3 { t.Errorf("want 3, got %d", len(cm.GetGroupMembers("g1"))) }
	cm.LeaveGroup("g1", "u2")
	if len(cm.GetGroupMembers("g1")) != 2 { t.Error("u2 not removed") }
	cm.Remove("u3", "d1")
	if len(cm.GetGroupMembers("g1")) != 2 { t.Error("group membership should persist after disconnect") }
}

func TestConnManagerRoomRouting(t *testing.T) {
	cm := data.NewMemoryConnManager()
	cm.Add("u1", "d1", nil)
	cm.Add("u2", "d1", nil)
	cm.JoinRoom("r1", "u1")
	cm.JoinRoom("r1", "u2")
	if len(cm.GetRoomMembers("r1")) != 2 { t.Errorf("want 2, got %d", len(cm.GetRoomMembers("r1"))) }
	cm.LeaveRoom("r1", "u2")
	if len(cm.GetRoomMembers("r1")) != 1 { t.Error("u2 not removed") }
}
