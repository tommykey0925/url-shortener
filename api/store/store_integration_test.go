//go:build integration

package store

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const testTableName = "url-shortener-test"

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("ap-northeast-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "dummy")),
	)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String("http://localhost:8000")
	})

	// Create table (ignore error if already exists)
	_, _ = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(testTableName),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("code"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("code"), AttributeType: types.ScalarAttributeTypeS},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	// Clean up existing items
	s := NewWithClient(client, testTableName)
	items, _ := s.List(ctx)
	for _, item := range items {
		_ = s.Delete(ctx, item.Code)
	}

	return s
}

func TestIntegration_PutAndGet(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	u, err := s.Put(ctx, "test001", "https://example.com", "safe")
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if u.Code != "test001" {
		t.Errorf("expected code 'test001', got %q", u.Code)
	}
	if u.Original != "https://example.com" {
		t.Errorf("expected original URL, got %q", u.Original)
	}
	if u.SafeStatus != "safe" {
		t.Errorf("expected safe status, got %q", u.SafeStatus)
	}

	got, err := s.Get(ctx, "test001")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Original != "https://example.com" {
		t.Errorf("expected original URL from Get, got %q", got.Original)
	}
	if got.SafeStatus != "safe" {
		t.Errorf("expected safe status from Get, got %q", got.SafeStatus)
	}
}

func TestIntegration_GetNotFound(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestIntegration_List(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	s.Put(ctx, "list001", "https://a.com", "safe")
	s.Put(ctx, "list002", "https://b.com", "safe")
	s.Put(ctx, "list003", "https://c.com", "unsafe")

	urls, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(urls) != 3 {
		t.Errorf("expected 3 urls, got %d", len(urls))
	}
}

func TestIntegration_Delete(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	s.Put(ctx, "del001", "https://example.com", "safe")

	if err := s.Delete(ctx, "del001"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err := s.Get(ctx, "del001")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestIntegration_UpdateSafeStatus(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	s.Put(ctx, "status001", "https://example.com", "safe")

	if err := s.UpdateSafeStatus(ctx, "status001", "unsafe"); err != nil {
		t.Fatalf("UpdateSafeStatus failed: %v", err)
	}

	got, err := s.Get(ctx, "status001")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if got.SafeStatus != "unsafe" {
		t.Errorf("expected 'unsafe', got %q", got.SafeStatus)
	}
}

func TestIntegration_IncrementClicks(t *testing.T) {
	s := setupTestStore(t)
	ctx := context.Background()

	s.Put(ctx, "click001", "https://example.com", "safe")

	for i := 0; i < 3; i++ {
		if err := s.IncrementClicks(ctx, "click001"); err != nil {
			t.Fatalf("IncrementClicks failed on iteration %d: %v", i, err)
		}
	}

	got, err := s.Get(ctx, "click001")
	if err != nil {
		t.Fatalf("Get after clicks failed: %v", err)
	}
	if got.Clicks != 3 {
		t.Errorf("expected 3 clicks, got %d", got.Clicks)
	}
}
