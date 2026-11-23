package mongodb

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/logger"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database
var available bool

// Available reports whether the database passed its startup probe.
func Available() bool { return available }

func init() {
	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Pass credentials separately so connection diagnostics can use the host safely.
	db := config.GlobalConfig.MongoDB
	host := net.JoinHostPort(db.IP, strconv.Itoa(db.Port))
	clientOptions := options.Client().SetHosts([]string{host})
	if db.Username != "" {
		clientOptions.SetAuth(options.Credential{Username: db.Username, Password: db.Password})
	}

	// 连接到 MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect %v to MongoDB: %v", host, err))
		return
	}

	// Other packages create collection handles during init, even when the server is down.
	Client = client
	DB = client.Database(db.Database)

	if err = client.Ping(ctx, nil); err != nil {
		logger.Error(fmt.Sprintf("failed to ping MongoDB: %v", err))
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_ = client.Disconnect(cleanupCtx)
		return
	}

	available = true
	logger.Info("MongoDB connected successfully")
	return
}

func Close() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := Client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect MongoDB: %v", err)
		}
	}
	return nil
}

// ================= 工具方法 =================

func Collection(name string) *mongo.Collection {
	return DB.Collection(name)
}

func defaultContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// ================= CRUD 封装 =================

// 插入一条
func InsertOne(col string, document interface{}) (*mongo.InsertOneResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).InsertOne(ctx, document)
}

// 插入多条
func InsertMany(col string, documents []interface{}) (*mongo.InsertManyResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).InsertMany(ctx, documents)
}

// 查询一条并解码
func FindOne(col string, filter interface{}, result interface{}) error {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).FindOne(ctx, filter).Decode(result)
}

// 查询多条
func FindMany(col string, filter interface{}, result interface{}, opts ...*options.FindOptions) error {
	ctx, cancel := defaultContext()
	defer cancel()

	cursor, err := Collection(col).Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, result)
}

func FindAll(collectionName string, query, selector, result interface{}) error {
	collection := Collection(collectionName)
	cur, err := collection.Find(context.Background(), query, options.Find().SetProjection(selector))
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	return cur.All(context.Background(), result)
}

// 更新一条
func UpdateOne(col string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).UpdateOne(ctx, filter, update)
}

// 更新多条
func UpdateMany(col string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).UpdateMany(ctx, filter, update)
}

// 删除一条
func DeleteOne(col string, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).DeleteOne(ctx, filter)
}

// 删除多条
func DeleteMany(col string, filter interface{}) (*mongo.DeleteResult, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).DeleteMany(ctx, filter)
}

// 统计数量
func Count(col string, filter interface{}) (int64, error) {
	ctx, cancel := defaultContext()
	defer cancel()

	return Collection(col).CountDocuments(ctx, filter)
}

// 分页查询
func FindPage(
	col string,
	filter interface{},
	result interface{},
	page int64,
	pageSize int64,
	sort bson.D,
) error {

	ctx, cancel := defaultContext()
	defer cancel()

	if page < 1 {
		page = 1
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(sort)

	cursor, err := Collection(col).Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, result)
}
