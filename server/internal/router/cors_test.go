package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 复现退出群聊的真实根因：浏览器对 DELETE 先发 OPTIONS 预检，
// 如果 Access-Control-Allow-Methods 不包含 DELETE，前端会在到达业务 handler 前被 CORS 拦截。
func TestCORSPreflightAllowsDeleteForLeaveGroup(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/groups/3/members/me", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodDelete)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status 204, got %d", w.Code)
	}
	allowMethods := w.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Fatal("expected Access-Control-Allow-Methods header to be set")
	}
	if !strings.Contains(allowMethods, http.MethodDelete) {
		t.Fatalf("expected Access-Control-Allow-Methods to include DELETE, got %q", allowMethods)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow-origin echo, got %q", got)
	}
}
