package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PraveenPin/GroupService/groupModels"
	"github.com/PraveenPin/GroupService/repo"
	"github.com/PraveenPin/SwipeMeter/utils"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid/v4"
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
	ctx         context.Context
	redisClient *redis.Client
}

func (g *GroupController) RedisClient() *redis.Client {
	return g.redisClient
}

func (g *GroupController) SetRedisClient(redisClient *redis.Client) {
	g.redisClient = redisClient
}

func (g *GroupController) Ctx() context.Context {
	return g.ctx
}

func (g *GroupController) SetCtx(ctx context.Context) {
	g.ctx = ctx
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

	newGroup.GroupID = shortuuid.New()
	log.Println("Group object is :", newGroup)
	if len(newGroup.LeaderBoard) == 0 {
		newLeaderBoardItem := groupModels.LeaderBoardItem{newGroup.CreatedBy, 0.0}
		newGroup.LeaderBoard = append(newGroup.LeaderBoard, newLeaderBoardItem)
	}

	_, create_redis_err := g.createLeaderBoardInRedis(newGroup)

	if create_redis_err != nil {
		log.Fatal("Error create a group table in redis", err)
		response.Format(w, r, true, 418, create_redis_err)
		return
	}

	_, create_err := groupRepo.Create(newGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v creating group with", create_err, newGroup)
		response.Format(w, r, true, 418, create_err)
		return
	}

	log.Println("New group created:", newGroup)
	response.Format(w, r, false, 201, newGroup)
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
	_, joinError := g.AddUserToLeaderBoard(joinGroup)
	if joinError != nil {
		log.Fatal("Error %v adding user to group in redis", joinError, joinGroup)
		response.Format(w, r, true, 418, joinError)
		return
	}
	created, create_err := groupRepo.AddUserToGroup(joinGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v adding user to group in dynamodb", create_err, joinGroup)
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

	_, add_err := g.RemoveUserFromLeaderBoard(leaveGroup)
	if add_err != nil {
		log.Fatal("Error %v removing user from group in redis", add_err, leaveGroup)
		response.Format(w, r, true, 418, add_err)
		return

	}

	created, create_err := groupRepo.RemoveUserFromGroup(leaveGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v removing user from group in dynamodb", create_err, leaveGroup)
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

func (g *GroupController) createLeaderBoardInRedis(newGroup groupModels.Group) (bool, error) {
	pipe := g.redisClient.TxPipeline()

	fmt.Println("Created a redis pipeline")
	//pipe.SAdd(g.Ctx(), "groups", newGroup.GroupID)
	//fmt.Println("Added create a group to redis pipeline")
	scores := map[string]float32{}
	for _, v := range newGroup.LeaderBoard {
		//pipe.SAdd(g.Ctx(), newGroup.GroupID, v.Username)
		pipe.ZAdd(g.Ctx(), newGroup.GroupID, &redis.Z{
			Score:  float64(v.Score),
			Member: v.Username,
		})
		scores[v.Username] = v.Score
	}
	fmt.Println("Added users to group to redis pipeline")

	_, err := pipe.Exec(g.Ctx())
	if err != nil {
		log.Fatalf("Error is: ", err)
	}

	return true, nil
}

func (g *GroupController) AddUserToLeaderBoard(joinGroup groupModels.JoinGroupModel) (bool, error) {
	// Add a user with score to a group
	err := g.redisClient.ZAdd(g.Ctx(), joinGroup.GroupID, &redis.Z{
		Score:  float64(0.0),
		Member: joinGroup.Username,
	}).Err()

	if err != nil {
		return false, err
	}

	log.Println("User added to group successfully!")
	return true, nil
}

func (g *GroupController) RemoveUserFromLeaderBoard(joinGroup groupModels.JoinGroupModel) (bool, error) {
	// Add a user with score to a group
	err := g.redisClient.ZRem(g.Ctx(), joinGroup.GroupID, joinGroup.Username).Err()
	if err != nil {
		return false, err
	}

	fmt.Println("User deleted from group successfully!")
	return true, nil
}

func (g *GroupController) DeleteGroup(joinGroup groupModels.JoinGroupModel) (bool, error) {
	// Remove a group
	err := g.redisClient.Del(g.Ctx(), joinGroup.GroupID).Err()
	if err != nil {
		return false, err
	}

	fmt.Println("Group removed successfully!")
	return true, nil
}
