package biz_test

import (
	"testing"

	"github.com/murphy-hc/h-im/services/gateway/internal/data"
)

func require(t *testing.T, err error) { t.Helper(); if err != nil { t.Fatal(err) } }

func TestConnManagerAddRemove(t *testing.T) {
	cm := data.NewMemoryConnManager()
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
	require(t, cm.Add("u1", "d1", nil))
	if c := cm.OnlineCount(); c != 1 { t.Errorf("want 1, got %d", c) }
	conns, _ := cm.GetConns("u1")
	if len(conns) != 1 { t.Errorf("want 1, got %d", len(conns)) }
	require(t, cm.Remove("u1", "d1"))
	if c := cm.OnlineCount(); c != 0 { t.Errorf("want 0, got %d", c) }
}

func TestConnManagerMultiDevice(t *testing.T) {
	cm := data.NewMemoryConnManager()
	require(t, cm.Add("u1", "phone", nil))
	require(t, cm.Add("u1", "laptop", nil))
	if c := cm.OnlineCount(); c != 2 { t.Errorf("want 2, got %d", c) }
	conns, _ := cm.GetConns("u1")
	if len(conns) != 2 { t.Errorf("want 2, got %d", len(conns)) }
	require(t, cm.Remove("u1", "phone"))
	conns, _ = cm.GetConns("u1")
	if len(conns) != 1 { t.Errorf("want 1, got %d", len(conns)) }
}

func TestConnManagerKickUser(t *testing.T) {
	cm := data.NewMemoryConnManager()
	require(t, cm.Add("u1", "d1", nil))
	require(t, cm.Add("u1", "d2", nil))
	kicked, err := cm.KickUser("u1")
	require(t, err)
	if len(kicked) != 2 { t.Errorf("want 2, got %d", len(kicked)) }
	if cm.OnlineCount() != 0 { t.Errorf("want 0, got %d", cm.OnlineCount()) }
}

func TestConnManagerGroupRouting(t *testing.T) {
	cm := data.NewMemoryConnManager()
	require(t, cm.Add("u1", "d1", nil))
	require(t, cm.Add("u2", "d1", nil))
	require(t, cm.Add("u3", "d1", nil))
	require(t, cm.JoinGroup("g1", "u1"))
	require(t, cm.JoinGroup("g1", "u2"))
	require(t, cm.JoinGroup("g1", "u3"))
	members, _ := cm.GetGroupMembers("g1")
	if len(members) != 3 { t.Errorf("want 3, got %d", len(members)) }
	require(t, cm.LeaveGroup("g1", "u2"))
	members, _ = cm.GetGroupMembers("g1")
	if len(members) != 2 { t.Error("u2 not removed") }
	require(t, cm.Remove("u3", "d1"))
	members, _ = cm.GetGroupMembers("g1")
	if len(members) != 2 { t.Error("group membership should persist after disconnect") }
}

func TestConnManagerRoomRouting(t *testing.T) {
	cm := data.NewMemoryConnManager()
	require(t, cm.Add("u1", "d1", nil))
	require(t, cm.Add("u2", "d1", nil))
	require(t, cm.JoinRoom("r1", "u1"))
	require(t, cm.JoinRoom("r1", "u2"))
	members, _ := cm.GetRoomMembers("r1")
	if len(members) != 2 { t.Errorf("want 2, got %d", len(members)) }
	require(t, cm.LeaveRoom("r1", "u2"))
	members, _ = cm.GetRoomMembers("r1")
	if len(members) != 1 { t.Error("u2 not removed") }
}
