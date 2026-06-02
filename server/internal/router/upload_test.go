package router

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
)

func TestAuthenticatedUserCanUploadImage(t *testing.T) {
	r := newTestRouter(t)
	token := registerAndLogin(t, r, "uploader")

	body, contentType := buildUploadBody(t, "cover.png", "image/png", []byte("fake-image-bytes"))
	req := httptest.NewRequest(http.MethodPost, "/api/uploads/images", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected upload status 201, got %d with body %s", w.Code, w.Body.String())
	}

	var resp struct {
		ObjectKey string `json:"objectKey"`
		URL       string `json:"url"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid upload response json, got error: %v", err)
	}
	if resp.ObjectKey == "" {
		t.Fatal("expected objectKey to be returned")
	}
	if resp.URL == "" {
		t.Fatal("expected url to be returned")
	}
}

func TestUploadRejectsNonImageFile(t *testing.T) {
	r := newTestRouter(t)
	token := registerAndLogin(t, r, "notimage")

	body, contentType := buildUploadBody(t, "note.txt", "text/plain", []byte("hello"))
	req := httptest.NewRequest(http.MethodPost, "/api/uploads/images", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected upload status 400, got %d with body %s", w.Code, w.Body.String())
	}
}

func buildUploadBody(t *testing.T, filename string, partContentType string, content []byte) (bytes.Buffer, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	header.Set("Content-Type", partContentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}
