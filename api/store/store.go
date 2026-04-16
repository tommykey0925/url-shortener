package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/tommykey-apps/url-shortener-api/model"
)

var ErrNotFound = errors.New("url not found")

type Store struct {
	client    *dynamodb.Client
	tableName string
}

func New() *Store {
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "url-shortener"
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	return &Store{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}
}

func NewWithClient(client *dynamodb.Client, tableName string) *Store {
	return &Store{
		client:    client,
		tableName: tableName,
	}
}

func (s *Store) Put(ctx context.Context, code, originalURL, safeStatus string) (*model.URL, error) {
	u := &model.URL{
		Code:       code,
		Original:   originalURL,
		CreatedAt:  time.Now().UTC(),
		Clicks:     0,
		SafeStatus: safeStatus,
	}

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("put item: %w", err)
	}

	return u, nil
}

func (s *Store) Get(ctx context.Context, code string) (*model.URL, error) {
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"code": &types.AttributeValueMemberS{Value: code},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get item: %w", err)
	}
	if out.Item == nil {
		return nil, ErrNotFound
	}

	var u model.URL
	if err := attributevalue.UnmarshalMap(out.Item, &u); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &u, nil
}

func (s *Store) IncrementClicks(ctx context.Context, code string) error {
	today := time.Now().UTC().Format("2006-01-02")

	// Update total clicks
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"code": &types.AttributeValueMemberS{Value: code},
		},
		UpdateExpression: aws.String("SET clicks = clicks + :inc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	if err != nil {
		return err
	}

	// Record daily click in stats table
	statsTable := s.tableName + "-stats"
	_, err = s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &statsTable,
		Key: map[string]types.AttributeValue{
			"code": &types.AttributeValueMemberS{Value: code},
			"date": &types.AttributeValueMemberS{Value: today},
		},
		UpdateExpression: aws.String("SET clicks = if_not_exists(clicks, :zero) + :inc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inc":  &types.AttributeValueMemberN{Value: "1"},
			":zero": &types.AttributeValueMemberN{Value: "0"},
		},
	})
	return err
}

func (s *Store) GetClickStats(ctx context.Context, code string, days int) ([]model.DailyClicks, error) {
	statsTable := s.tableName + "-stats"
	startDate := time.Now().UTC().AddDate(0, 0, -days).Format("2006-01-02")

	out, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &statsTable,
		KeyConditionExpression: aws.String("code = :code AND #d >= :start"),
		ExpressionAttributeNames: map[string]string{
			"#d": "date",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":code":  &types.AttributeValueMemberS{Value: code},
			":start": &types.AttributeValueMemberS{Value: startDate},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query stats: %w", err)
	}

	var stats []model.DailyClicks
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	return stats, nil
}

func (s *Store) UpdateSafeStatus(ctx context.Context, code, status string) error {
	_, err := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"code": &types.AttributeValueMemberS{Value: code},
		},
		UpdateExpression: aws.String("SET safe_status = :s"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": &types.AttributeValueMemberS{Value: status},
		},
	})
	return err
}

func (s *Store) Delete(ctx context.Context, code string) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"code": &types.AttributeValueMemberS{Value: code},
		},
	})
	return err
}

func (s *Store) List(ctx context.Context) ([]model.URL, error) {
	out, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: &s.tableName,
	})
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	var urls []model.URL
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &urls); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return urls, nil
}
