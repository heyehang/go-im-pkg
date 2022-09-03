package mongosdk

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"
)

var (
	cli *mongo.Client
)

func InitMongo(ctx context.Context, uri, userName, pwd string, poolSize int) {
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(uri),
		options.Client().SetAuth(options.Credential{
			Username: userName,
			Password: pwd,
		}),
		//使用 readConcern 需要配置replication.enableMajorityReadConcern选项
		options.Client().SetReadConcern(readconcern.Majority()),
		// 保证数据安全,写操作需要被复制到大多数节点上才算成功
		options.Client().SetWriteConcern(writeconcern.New(writeconcern.WMajority())),
		// 连接池大小
		options.Client().SetMaxPoolSize(uint64(poolSize)),
		options.Client().SetMinPoolSize(20),
		options.Client().SetMaxConnecting(99),
		//优先从 secondary 读取，没有 secondary 成员时，从 primary 读取
		options.Client().SetReadPreference(readpref.SecondaryPreferred()),
		options.Client().SetConnectTimeout(time.Second*30),
		options.Client().SetMaxConnIdleTime(0),
	)
	if err != nil {
		panic(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}
	cli = client
}

func GetClient() *mongo.Client {
	return cli
}

func Close(ctx context.Context) {
	_ = cli.Disconnect(ctx)
}
