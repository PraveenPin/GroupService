package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PraveenPin/GroupService/controllers"
	"github.com/PraveenPin/GroupService/groupModels"
	"github.com/PraveenPin/GroupService/init_database"
	"github.com/PraveenPin/GroupService/services"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/form3tech-oss/jwt-go"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const PORT = ":8083"

type Dispatcher struct {
}

func HomeEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world :)")
}

func (r *Dispatcher) StartSubscriber(pubsubClient *pubsub.Client, ctx context.Context, groupController *controllers.GroupController) {
	log.Println("Starting pub/sub subscriber")
	sub := pubsubClient.Subscription(init_database.SubID)

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

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(fmt.Sprintf("https://%s/.well-known/jwks.json", init_database.AUTH0_DOMAIN))

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = groupModels.Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k, _ := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return cert, err
	}

	return cert, nil
}

func (r *Dispatcher) Init(db *dynamodb.DynamoDB, rdc *redis.Client, ctx context.Context, pubsubClient *pubsub.Client, grpcClient services.UserServiceClient) {
	groupController := controllers.NewGroupController(db, ctx, rdc, grpcClient)
	//Start Subscriber Service
	go r.StartSubscriber(pubsubClient, ctx, groupController)

	log.Println("Initialize the router")
	router := mux.NewRouter()

	log.Println("Securing all endpoints with jwt middleware")
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			aud := init_database.AUTH0_AUDIENCE
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New("Invalid audience.")
			}
			// Verify 'iss' claim
			iss := fmt.Sprintf("https://%s/", init_database.AUTH0_DOMAIN)
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})
	router.StrictSlash(true)
	router.HandleFunc("/", HomeEndpoint).Methods("GET")

	// Group Resource
	groupRoutes := router.PathPrefix("/group").Subrouter()
	groupRoutes.Handle("/createGroup", jwtMiddleware.Handler(http.HandlerFunc(groupController.CreateGroupController))).Methods("POST")
	groupRoutes.Handle("/joinGroup", jwtMiddleware.Handler(http.HandlerFunc(groupController.JoinGroupController))).Methods("POST")
	groupRoutes.Handle("/leaveGroup", jwtMiddleware.Handler(http.HandlerFunc(groupController.LeaveGroupController))).Methods("POST")
	//groupRoutes.HandleFunc("/openGroup", groupController.getGroup).Methods("POST")

	//Testing purposes
	groupRoutes.Handle("/getGroup", jwtMiddleware.Handler(http.HandlerFunc(groupController.GetGroup))).Methods("GET")

	corsWrapper := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Origin", "Accept", "*"},
	})

	log.Println("Add the listener to port ", PORT)

	//serve
	http.ListenAndServe(PORT, corsWrapper.Handler(router))
}
