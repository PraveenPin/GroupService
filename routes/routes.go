package routes

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/PraveenPin/GroupService/controllers"
	"github.com/PraveenPin/GroupService/init_database"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const PORT = ":8083"
const GRPC_PORT = ":8085"
const GRPC_TARGET = "localhost:9000"

type Dispatcher struct {
}

func HomeEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world :)")
}

func (r *Dispatcher) StartSubscriber(pubsubClient *pubsub.Client, ctx context.Context, groupController *controllers.GroupController) {
	log.Println("Starting pub/sub subscriber")
	sub := pubsubClient.Subscription(init_database.SubID)
	log.Println("Adtrr pub/sub subscriber")

	err := sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Println("Got message: ", string(msg.Data), msg.Attributes)
		msg.Ack()

		message := strings.Split(string(msg.Data), "|")
		score, err := strconv.ParseFloat(message[2], 64)

		if err != nil {
			log.Println(err)
		}
		groupController.UpdateScoresController(message[0], score)

	})
	if err != nil {
		log.Fatalf("sub.Receive:", err)
		return
	}
	return
}

func (r *Dispatcher) Init(db *dynamodb.DynamoDB, rdc *redis.Client, ctx context.Context, pubsubClient *pubsub.Client) {
	groupController := controllers.NewGroupController(db, ctx, rdc)
	//Start Subscriber Service
	go r.StartSubscriber(pubsubClient, ctx, groupController)

	log.Println("Initialize the router")
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/", HomeEndpoint).Methods("GET")

	// Group Resource
	groupRoutes := router.PathPrefix("/group").Subrouter()
	groupRoutes.HandleFunc("/createGroup", groupController.CreateGroupController).Methods("POST")
	groupRoutes.HandleFunc("/joinGroup", groupController.JoinGroupController).Methods("POST")
	groupRoutes.HandleFunc("/leaveGroup", groupController.LeaveGroupController).Methods("POST")
	//groupRoutes.HandleFunc("/openGroup", groupController.getGroup).Methods("POST")

	//Testing purposes
	groupRoutes.HandleFunc("/getGroup", groupController.GetGroup).Methods("POST")

	// bind the routes
	http.Handle("/", router)

	log.Println("Add the listener to port ", PORT)

	//serve
	http.ListenAndServe(PORT, nil)
}
