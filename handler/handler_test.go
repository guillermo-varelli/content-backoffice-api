package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/workflowapi/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestServer(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("opening in-memory sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.Agent{}, &model.Workflow{}, &model.Step{}, &model.N{}); err != nil {
		t.Fatalf("auto migrating schema: %v", err)
	}

	r := gin.Default()
	RegisterAgentRoutes(r, db)
	RegisterWorkflowRoutes(r, db)
	RegisterStepRoutes(r, db)
	RegisterNRoutes(r, db)

	return r, db
}

func performJSONRequest(t *testing.T, r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encoding request body: %v", err)
		}
		buf.Write(payload)
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAgentCRUDFlow(t *testing.T) {
	router, _ := setupTestServer(t)

	createBody := model.Agent{Provider: "openai", Secret: "secret-1"}
	resp := performJSONRequest(t, router, http.MethodPost, "/agents", createBody)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.Code)
	}

	var created model.Agent
	if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created agent to have an ID")
	}

	listResp := performJSONRequest(t, router, http.MethodGet, "/agents?page=1&size=5", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResp.Code)
	}

	var list []model.Agent
	if err := json.Unmarshal(listResp.Body.Bytes(), &list); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(list))
	}

	updateBody := model.Agent{Provider: "azure-openai", Secret: "secret-2"}
	updatePath := fmt.Sprintf("/agents/%d", created.ID)
	updateResp := performJSONRequest(t, router, http.MethodPut, updatePath, updateBody)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateResp.Code)
	}

	afterUpdate := performJSONRequest(t, router, http.MethodGet, "/agents?page=1&size=5", nil)
	if afterUpdate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, afterUpdate.Code)
	}

	var updatedList []model.Agent
	if err := json.Unmarshal(afterUpdate.Body.Bytes(), &updatedList); err != nil {
		t.Fatalf("decoding updated list response: %v", err)
	}
	if len(updatedList) != 1 || updatedList[0].Provider != updateBody.Provider {
		t.Fatalf("expected provider %q after update, got %+v", updateBody.Provider, updatedList)
	}

	deletePath := fmt.Sprintf("/agents/%d", created.ID)
	deleteResp := performJSONRequest(t, router, http.MethodDelete, deletePath, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResp.Code)
	}

	finalList := performJSONRequest(t, router, http.MethodGet, "/agents", nil)
	if finalList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, finalList.Code)
	}

	var empty []model.Agent
	if err := json.Unmarshal(finalList.Body.Bytes(), &empty); err != nil {
		t.Fatalf("decoding final list response: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no agents after delete, got %d", len(empty))
	}
}

func TestNotificationCRUDFlow(t *testing.T) {
	router, _ := setupTestServer(t)

	now := time.Now().UTC()
	createBody := model.N{
		ExecutionID:      99,
		Title:            "Test notification",
		ShortDescription: "Short description",
		Message:          "Hello world",
		Status:           "pending",
		Type:             "info",
		SubType:          "email",
		Category:         "ops",
		SubCategory:      "deploy",
		ImageURL:         "http://example.com/image.png",
		ImagePrompt:      "prompt",
		Created:          now,
		LastUpdated:      now,
	}

	resp := performJSONRequest(t, router, http.MethodPost, "/n", createBody)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.Code)
	}

	var created model.N
	if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}
	if created.ID == 0 || created.Title != createBody.Title {
		t.Fatalf("expected created notification with title %q, got %+v", createBody.Title, created)
	}

	listResp := performJSONRequest(t, router, http.MethodGet, "/n?page=1&size=10", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResp.Code)
	}

	var list []model.N
	if err := json.Unmarshal(listResp.Body.Bytes(), &list); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(list))
	}

	updated := created
	updated.Status = "sent"
	updated.Message = "Updated message"
	updated.LastUpdated = now.Add(time.Minute)

	updatePath := fmt.Sprintf("/n/%d", created.ID)
	updateResp := performJSONRequest(t, router, http.MethodPut, updatePath, updated)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateResp.Code)
	}

	afterUpdate := performJSONRequest(t, router, http.MethodGet, "/n", nil)
	if afterUpdate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, afterUpdate.Code)
	}

	var updatedList []model.N
	if err := json.Unmarshal(afterUpdate.Body.Bytes(), &updatedList); err != nil {
		t.Fatalf("decoding updated list response: %v", err)
	}
	if len(updatedList) != 1 || updatedList[0].Status != updated.Status {
		t.Fatalf("expected status %q after update, got %+v", updated.Status, updatedList)
	}

	deletePath := fmt.Sprintf("/n/%d", created.ID)
	deleteResp := performJSONRequest(t, router, http.MethodDelete, deletePath, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResp.Code)
	}

	finalList := performJSONRequest(t, router, http.MethodGet, "/n", nil)
	if finalList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, finalList.Code)
	}

	var empty []model.N
	if err := json.Unmarshal(finalList.Body.Bytes(), &empty); err != nil {
		t.Fatalf("decoding final list response: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no notifications after delete, got %d", len(empty))
	}
}
