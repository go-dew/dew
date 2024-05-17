package dew

import "testing"

func TestTree(t *testing.T) {
	tr := &node{}
	createUser := &handler{}
	createUser2 := &handler{}
	createAccount := &handler{}
	findUser := &handler{}
	findUserByUserID := &handler{}
	findUserByUserName := &handler{}

	tr.insert(ACTION, "CreateUser", createUser)
	tr.insert(ACTION, "CreateUser2", createUser2)
	tr.insert(ACTION, "CreateAccount", createAccount)
	tr.insert(QUERY, "FindUser", findUser)
	tr.insert(QUERY, "FindUserByUserID", findUserByUserID)
	tr.insert(QUERY, "FindUserByUserName", findUserByUserName)

	tests := []struct {
		o OpType
		k string
		h *handler
	}{
		{ACTION, "CreateUser", createUser},
		{ACTION, "CreateUser2", createUser2},
		{QUERY, "CreateUser", nil}, // not found
		{ACTION, "CreateAccount", createAccount},
		{QUERY, "FindUser", findUser},
		{QUERY, "FindUserByUserID", findUserByUserID},
		{QUERY, "FindUserByUserName", findUserByUserName},
	}

	for _, tt := range tests {
		n := tr.findRoute(tt.o, tt.k)
		if n == nil && tt.h != nil {
			t.Fatalf("expected %s, got nil", tt.k)
		}
		if !(n == nil && (tt.h == nil)) && n.handler.handler != tt.h {
			t.Errorf("exected %p, got %p", tt.h, n.handler.handler)
		}
	}
}

func TestTree_Panic_DuplicateHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic")
		}
	}()

	tr := &node{}
	tr.insert(ACTION, "CreateUser", &handler{})
	tr.insert(ACTION, "CreateUser", &handler{})
}

func TestTree_Panic_ReplacingMissingChild(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic")
		}
	}()
	tr := &node{}
	tr.replaceChild(byte('a'), &node{})
}
