package init_database

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-redis/redis/v8"
	"log"
)

const (
	projectID     = "trackingservice-383922"
	SubID         = "TrackingSubscription"
	topic         = "swipe-track-record"
	redisPassowrd = "rFhBRo645gvbwRSSTimJfrVNhuUhhbOG"
	redisHostName = "redis-11838.c245.us-east-1-3.ec2.cloud.redislabs.com:11838"
)

type App struct {
}

func (a *App) GetRedisClient(ctx context.Context) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisHostName,
		Password: redisPassowrd,
	})
	log.Println("Testing connection to redis", client)
	// Test the connection
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	log.Println("Connection to redis database successful", pong)

	return client
}

func (a *App) GetAppContext() context.Context {
	log.Println("Initialising App Context")
	return context.Background()
}

func (a *App) GetPubSubClient(ctx context.Context) *pubsub.Client {

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("init_database.NewClient:", err)
		return nil
	}
	fmt.Println("Pub/Sub client obtained")
	//defer client.Close()
	return client
}

func (a *App) CloseClient(client *pubsub.Client) {
	fmt.Println("Will close the pub/sub client shortly")
	defer client.Close()
}

func (app *App) StartAWSSession() *session.Session {
	log.Println("Initiating aws session to create tables")
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return sess
}

func (app *App) GetDynamoDatabaseClient(session *session.Session) *dynamodb.DynamoDB {
	svc := dynamodb.New(session)
	log.Println("Dynamodb client connector obtained")
	return svc
}
