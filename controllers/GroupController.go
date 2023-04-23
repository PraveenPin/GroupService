package controllers

import (
	"encoding/json"
	"github.com/PraveenPin/GroupService/groupModels"
	"github.com/PraveenPin/GroupService/repo"
	"github.com/PraveenPin/SwipeMeter/utils"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"log"
	"net/http"
)

var response *utils.Response

type GroupControllerInterface interface {
	DynamodbSVC() *dynamodb.DynamoDB
	SetDynamodbSVC(dynamodbSVC *dynamodb.DynamoDB)
}

type GroupController struct {
	dynamodbSVC *dynamodb.DynamoDB
}

func (g *GroupController) DynamodbSVC() *dynamodb.DynamoDB {
	return g.dynamodbSVC
}

func (g *GroupController) SetDynamodbSVC(dynamodbSVC *dynamodb.DynamoDB) {
	g.dynamodbSVC = dynamodbSVC
}

func (g *GroupController) CreateGroup(w http.ResponseWriter, r *http.Request) {
	groupRepo := &repo.GroupRepository{}

	log.Println("Create Group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	newGroup := groupModels.Group{}
	err := decoder.Decode(&newGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}

	newGroup.GroupID = uuid.New().String()
	log.Println("Group object is :", newGroup)

	created, create_err := groupRepo.Create(newGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v creating group with", create_err, newGroup)
		response.Format(w, r, true, 418, create_err)
		return
	}

	if created {
		log.Println("New group created:", newGroup)
		response.Format(w, r, false, 201, newGroup)
		return
	}

	return
}

func (g *GroupController) JoinGroup(w http.ResponseWriter, r *http.Request) {
	groupRepo := &repo.GroupRepository{}

	log.Println("Add a user to group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	joinGroup := groupModels.JoinGroupModel{}
	err := decoder.Decode(&joinGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}

	created, create_err := groupRepo.AddUserToGroup(joinGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v adding user to group", create_err, joinGroup)
		response.Format(w, r, true, 418, create_err)
		return
	}

	if created {
		log.Println("Added User to group:", joinGroup)
		response.Format(w, r, false, 201, joinGroup)
		return
	}

	return
}

func (g *GroupController) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	groupRepo := &repo.GroupRepository{}

	log.Println("Remove a user from a group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	leaveGroup := groupModels.JoinGroupModel{}
	err := decoder.Decode(&leaveGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}

	created, create_err := groupRepo.RemoveUserFromGroup(leaveGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v removing user from group", create_err, leaveGroup)
		response.Format(w, r, true, 418, create_err)
		return
	}

	if created {
		log.Println("Removed User from group:", leaveGroup)
		response.Format(w, r, false, 201, leaveGroup)
		return
	}

	return
}

func (g *GroupController) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupRepo := &repo.GroupRepository{}

	log.Println("Remove a user from a group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	leaveGroup := groupModels.JoinGroupModel{}
	err := decoder.Decode(&leaveGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}

	returnVal := groupRepo.GetOne(leaveGroup.GroupID, g.DynamodbSVC())
	response.Format(w, r, false, 200, returnVal)
	return
}
