package controllers

import (
	"context"
	"encoding/json"
	"github.com/PraveenPin/GroupService/groupModels"
	"github.com/PraveenPin/GroupService/repo"
	"github.com/PraveenPin/GroupService/services"
	"github.com/PraveenPin/SwipeMeter/utils"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
)

var response *utils.Response

type GroupControllerInterface interface {
	DynamodbSVC() *dynamodb.DynamoDB
	SetDynamodbSVC(dynamodbSVC *dynamodb.DynamoDB)
}

type GroupController struct {
	groupService *services.GroupService
}

func NewGroupController(dynamodbSVC *dynamodb.DynamoDB, ctx context.Context, redisClient *redis.Client, grpcClient services.UserServiceClient) *GroupController {
	groupRepo := &repo.GroupRepository{}
	groupService := services.NewGroupService(dynamodbSVC, ctx, redisClient, groupRepo, grpcClient)
	return &GroupController{groupService}
}

func (g *GroupController) CreateGroupController(w http.ResponseWriter, r *http.Request) {

	log.Println("Create Group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	newGroup := groupModels.Group{}
	err := decoder.Decode(&newGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}

	_, err = g.groupService.CreateGroupService(newGroup)

	if err != nil {
		response.Format(w, r, true, 418, err)
		return
	}

	response.Format(w, r, false, 201, newGroup)
}

func (g *GroupController) JoinGroupController(w http.ResponseWriter, r *http.Request) {

	log.Println("Add a user to group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	joinGroup := groupModels.JoinGroupModel{}
	err := decoder.Decode(&joinGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}

	_, joinError := g.groupService.JoinGroupService(joinGroup)
	if joinError != nil {
		log.Fatal("Error %v adding user to group in redis", joinError, joinGroup)
		response.Format(w, r, true, 418, joinError)
		return
	}

	response.Format(w, r, false, 201, joinGroup)
	return
}

func (g *GroupController) LeaveGroupController(w http.ResponseWriter, r *http.Request) {
	log.Println("Remove a user from a group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	leaveGroup := groupModels.JoinGroupModel{}
	err := decoder.Decode(&leaveGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}

	_, r_err := g.groupService.LeaveGroupService(leaveGroup)
	if r_err != nil {
		response.Format(w, r, true, 418, r_err)
		return

	}

	response.Format(w, r, false, 201, leaveGroup)
	return
}

func (g *GroupController) GetGroup(w http.ResponseWriter, r *http.Request) {
	log.Println("Remove a user from a group Request: ", r)
	decoder := json.NewDecoder(r.Body)

	leaveGroup := groupModels.JoinGroupModel{}
	err := decoder.Decode(&leaveGroup)
	if err != nil {
		response.Format(w, r, true, 417, err)
		return
	}
	returnVal := g.groupService.GetGroupService(leaveGroup)
	response.Format(w, r, false, 200, returnVal)
	return
}

func (g *GroupController) UpdateScoresController(username string, score float64) {
	groups := g.groupService.GetAllUserGroupsServiceAndUpdateTotalScore(username, float32(score))
	// update the score in each group board
	for _, groupId := range groups {
		success := g.groupService.UpdateScoreForUserInAGroup(username, groupId, float64(score))
		if success != true {
			log.Fatal("Group Score for user:", username, " cannot be updated for group:", groupId)
		}
	}
	return
}
