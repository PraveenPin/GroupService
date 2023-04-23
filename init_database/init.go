package init_database

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"log"
)

const (
	projectID = "trackingservice-383922"
	subID     = "TrackingSubcription"
	topic     = "swipe-track-record"
)

type App struct {
}

func (a *App) GetAppContext() context.Context {
	fmt.Println("Initialising App Context")
	return context.Background()
}

func (a *App) GetPubSubClient(ctx context.Context) *pubsub.Client {

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("init_database.NewClient:", err)
		return nil
	}
	fmt.Println("Pub/Sub client obtained")
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
