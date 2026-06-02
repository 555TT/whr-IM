package model

import (
	"os"
	"strings"
	"testing"
)

func TestInitSQLMatchesCurrentUserFriendAndFriendRequestModels(t *testing.T) {
	content, err := os.ReadFile("../../../init.sql")
	if err != nil {
		t.Fatalf("read init.sql: %v", err)
	}

	sql := string(content)
	usersTable := tableSQL(t, sql, "users")
	friendRequestsTable := tableSQL(t, sql, "friend_requests")
	friendsTable := tableSQL(t, sql, "friends")

	if strings.Contains(usersTable, "created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'") {
		t.Fatal("expected users table in init.sql to match current model without created_at")
	}
	if strings.Contains(usersTable, "updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'") {
		t.Fatal("expected users table in init.sql to match current model without updated_at")
	}

	if strings.Contains(friendRequestsTable, "created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'") {
		t.Fatal("expected friend_requests table in init.sql to match current model without created_at")
	}
	if strings.Contains(friendRequestsTable, "handled_at DATETIME NULL DEFAULT NULL COMMENT '处理时间'") {
		t.Fatal("expected friend_requests table in init.sql to match current model without handled_at")
	}
	if !strings.Contains(friendRequestsTable, "status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '申请状态'") {
		t.Fatal("expected friend_requests.status to be VARCHAR(20) with default pending to match current model")
	}

	if strings.Contains(friendsTable, "created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'") {
		t.Fatal("expected friends table in init.sql to match current model without created_at")
	}
}

func tableSQL(t *testing.T, sql string, tableName string) string {
	t.Helper()

	startMarker := "CREATE TABLE " + tableName + " ("
	start := strings.Index(sql, startMarker)
	if start == -1 {
		t.Fatalf("table %s not found in init.sql", tableName)
	}

	rest := sql[start:]
	end := strings.Index(rest, ") ENGINE=")
	if end == -1 {
		t.Fatalf("table %s has no closing ENGINE clause in init.sql", tableName)
	}

	return rest[:end]
}
