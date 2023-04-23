package routes

import (
	"fmt"
	"github.com/PraveenPin/GroupService/controllers"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const PORT = ":8082"

type Dispatcher struct {
}

func HomeEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world :)")
}

func (r *Dispatcher) Init(db *dynamodb.DynamoDB) {
	log.Println("Initialize the router")
	router := mux.NewRouter()

	groupController := &controllers.GroupController{}
	groupController.SetDynamodbSVC(db)

	router.StrictSlash(true)
	router.HandleFunc("/", HomeEndpoint).Methods("GET")

	// Group Resource
	groupRoutes := router.PathPrefix("/group").Subrouter()
	groupRoutes.HandleFunc("/createGroup", groupController.CreateGroup).Methods("POST")
	groupRoutes.HandleFunc("/joinGroup", groupController.JoinGroup).Methods("POST")
	groupRoutes.HandleFunc("/leaveGroup", groupController.LeaveGroup).Methods("POST")
	//groupRoutes.HandleFunc("/openGroup", groupController.getGroup).Methods("POST")

	//Testing purposes
	groupRoutes.HandleFunc("/getGroup", groupController.GetGroup).Methods("POST")

	// bind the routes
	http.Handle("/", router)

	log.Println("Add the listener to port ", PORT)

	//serve
	http.ListenAndServe(PORT, nil)
}

func profile(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("test"))
}
