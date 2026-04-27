//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/tommykey-apps/url-shortener-api/store"
)

const testTableName = "url-shortener-monitor-test"

type mockSB struct {
	unsafeURLs map[string]bool
}

func (m *mockSB) Check(targetURL string) (bool, string, error) {
	if m.unsafeURLs[targetURL] {
		return false, "blocked: MALWARE", nil
	}
	return true, "safe", nil
}

func setupMonitorTestStore(t *testing.T) *store.Store {
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

	s := store.NewWithClient(client, testTableName, testTableName+"-stats")
	items, _ := s.List(ctx)
	for _, item := range items {
		_ = s.Delete(ctx, item.Code)
	}

	return s
}

func TestIntegration_RunMonitor_FlagsUnsafeURLs(t *testing.T) {
	s := setupMonitorTestStore(t)
	ctx := context.Background()

	s.Put(ctx, "safe01", "https://example.com", "safe")
	s.Put(ctx, "safe02", "https://google.com", "safe")
	s.Put(ctx, "bad01", "https://malware.example.com", "safe")

	sb := &mockSB{
		unsafeURLs: map[string]bool{
			"https://malware.example.com": true,
		},
	}

	if err := RunMonitor(ctx, s, sb); err != nil {
		t.Fatalf("RunMonitor failed: %v", err)
	}

	// safe URLs should remain safe
	u1, _ := s.Get(ctx, "safe01")
	if u1.SafeStatus != "safe" {
		t.Errorf("safe01 should remain safe, got %q", u1.SafeStatus)
	}
	u2, _ := s.Get(ctx, "safe02")
	if u2.SafeStatus != "safe" {
		t.Errorf("safe02 should remain safe, got %q", u2.SafeStatus)
	}

	// malware URL should be flagged unsafe
	u3, _ := s.Get(ctx, "bad01")
	if u3.SafeStatus != "unsafe" {
		t.Errorf("bad01 should be unsafe, got %q", u3.SafeStatus)
	}
}

func TestIntegration_RunMonitor_SkipsAlreadyUnsafe(t *testing.T) {
	s := setupMonitorTestStore(t)
	ctx := context.Background()

	// Already marked as unsafe — should not trigger an update
	s.Put(ctx, "already01", "https://malware.example.com", "unsafe")

	sb := &mockSB{
		unsafeURLs: map[string]bool{
			"https://malware.example.com": true,
		},
	}

	if err := RunMonitor(ctx, s, sb); err != nil {
		t.Fatalf("RunMonitor failed: %v", err)
	}

	u, _ := s.Get(ctx, "already01")
	if u.SafeStatus != "unsafe" {
		t.Errorf("already01 should remain unsafe, got %q", u.SafeStatus)
	}
}

func TestIntegration_RunMonitor_AllSafe(t *testing.T) {
	s := setupMonitorTestStore(t)
	ctx := context.Background()

	s.Put(ctx, "ok01", "https://example.com", "safe")
	s.Put(ctx, "ok02", "https://google.com", "safe")

	sb := &mockSB{unsafeURLs: map[string]bool{}}

	if err := RunMonitor(ctx, s, sb); err != nil {
		t.Fatalf("RunMonitor failed: %v", err)
	}

	u1, _ := s.Get(ctx, "ok01")
	if u1.SafeStatus != "safe" {
		t.Errorf("ok01 should remain safe, got %q", u1.SafeStatus)
	}
	u2, _ := s.Get(ctx, "ok02")
	if u2.SafeStatus != "safe" {
		t.Errorf("ok02 should remain safe, got %q", u2.SafeStatus)
	}
}
