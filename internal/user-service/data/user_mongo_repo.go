package data

import (
	"context"
	"fmt"

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

// UserMongoRepository 用户仓库的 MongoDB 实现
// 实现 UserRepository 接口,提供基于 MongoDB 的数据持久化
type UserMongoRepository struct {
	client     *db.MongoClient
	collection *mongo.Collection
}

// NewUserMongoRepository 创建新的 MongoDB 用户仓库
// 使用依赖注入,接收 MongoDB 客户端作为参数
func NewUserMongoRepository(client *db.MongoClient) UserRepository {
	return &UserMongoRepository{
		client:     client,
		collection: client.GetCollection(CollectionUsers),
	}
}

// Create 创建用户
func (r *UserMongoRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID 根据ID获取用户
func (r *UserMongoRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserMongoRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// Update 更新用户
func (r *UserMongoRepository) Update(ctx context.Context, user *domain.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

// Delete 删除用户
func (r *UserMongoRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

// List 列出用户
func (r *UserMongoRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)
	
	var users []*domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}
	
	return users, nil
}
