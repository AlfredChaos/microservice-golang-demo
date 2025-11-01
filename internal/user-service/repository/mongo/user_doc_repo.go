package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/alfredchaos/demo/internal/user-service/domain"
	"github.com/alfredchaos/demo/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// CollectionUsers 用户集合名称
	CollectionUsers = "users"
)

type UserMongoDocumentRepository struct {
	client     *db.MongoClient
	collection *mongo.Collection
}

// NewUserMongoDocumentRepository 创建新的 MongoDB 用户文档仓库
func NewUserMongoDocumentRepository(client *db.MongoClient) *UserMongoDocumentRepository {
	return &UserMongoDocumentRepository{
		client:     client,
		collection: client.GetCollection(CollectionUsers),
	}
}

// SaveDocument 保存用户文档（JSON 格式）
func (r *UserMongoDocumentRepository) SaveDocument(ctx context.Context, userID string, document map[string]interface{}) error {
	document["_id"] = userID

	// 自动添加/更新时间戳
	now := time.Now()
	if _, exists := document["created_at"]; !exists {
		document["created_at"] = now
	}
	document["updated_at"] = now

	// Upsert 操作
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": document}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save document: %w", err)
	}

	return nil
}

// GetDocument 根据ID获取用户文档（JSON 格式）
func (r *UserMongoDocumentRepository) GetDocument(ctx context.Context, userID string) (map[string]interface{}, error) {
	var document map[string]interface{}

	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&document)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return document, nil
}

// DeleteDocument 删除用户文档
func (r *UserMongoDocumentRepository) DeleteDocument(ctx context.Context, userID string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if result.DeletedCount == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// FindDocuments 根据查询条件查找文档
func (r *UserMongoDocumentRepository) FindDocuments(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]map[string]interface{}, error) {
	// 构建查询选项
	opts := options.Find()
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	// 按创建时间倒序排序
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	// 执行查询
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	// 解码结果
	var documents []map[string]interface{}
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode documents: %w", err)
	}

	return documents, nil
}

// UpdateDocumentFields 更新文档的部分字段
func (r *UserMongoDocumentRepository) UpdateDocumentFields(ctx context.Context, userID string, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": fields}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document fields: %w", err)
	}

	if result.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
