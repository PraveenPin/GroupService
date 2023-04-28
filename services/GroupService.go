package services

import (
	"context"
	"fmt"
	"github.com/PraveenPin/GroupService/groupModels"
	"github.com/PraveenPin/GroupService/repo"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid/v4"
	"log"
)

type GroupService struct {
	dynamodbSVC *dynamodb.DynamoDB
	ctx         context.Context
	redisClient *redis.Client
	groupRepo   *repo.GroupRepository
}

func NewGroupService(dynamodbSVC *dynamodb.DynamoDB, ctx context.Context, redisClient *redis.Client, groupRepo *repo.GroupRepository) *GroupService {
	return &GroupService{dynamodbSVC: dynamodbSVC, ctx: ctx, redisClient: redisClient, groupRepo: groupRepo}
}

func (g *GroupService) DynamodbSVC() *dynamodb.DynamoDB {
	return g.dynamodbSVC
}

func (g *GroupService) SetDynamodbSVC(dynamodbSVC *dynamodb.DynamoDB) {
	g.dynamodbSVC = dynamodbSVC
}

func (g *GroupService) Ctx() context.Context {
	return g.ctx
}

func (g *GroupService) SetCtx(ctx context.Context) {
	g.ctx = ctx
}

func (g *GroupService) RedisClient() *redis.Client {
	return g.redisClient
}

func (g *GroupService) SetRedisClient(redisClient *redis.Client) {
	g.redisClient = redisClient
}

func (g *GroupService) GroupRepo() *repo.GroupRepository {
	return g.groupRepo
}

func (g *GroupService) SetGroupRepo(groupRepo *repo.GroupRepository) {
	g.groupRepo = groupRepo
}

func (g *GroupService) CreateGroupService(newGroup groupModels.Group) (bool, error) {

	newGroup.GroupID = shortuuid.New()
	log.Println("Group object is :", newGroup)
	if len(newGroup.LeaderBoard) == 0 {
		newLeaderBoardItem := groupModels.LeaderBoardItem{newGroup.CreatedBy, 0.0}
		newGroup.LeaderBoard = append(newGroup.LeaderBoard, newLeaderBoardItem)
	}

	_, create_redis_err := g.createLeaderBoardInRedis(newGroup)

	if create_redis_err != nil {
		log.Fatal("Error create a group table in redis", create_redis_err)
		return false, create_redis_err
	}

	_, create_err := g.groupRepo.Create(newGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v creating group with", create_err, newGroup)
		return false, create_err
	}

	log.Println("New group created:", newGroup)
	return true, nil
}

func (g *GroupService) JoinGroupService(joinGroup groupModels.JoinGroupModel) (bool, error) {
	_, joinError := g.AddUserToLeaderBoard(joinGroup)

	if joinError != nil {
		log.Fatal("Error %v adding user to group in redis", joinError, joinGroup)
		return false, joinError
	}
	_, create_err := g.groupRepo.AddUserToGroup(joinGroup, g.DynamodbSVC())

	if create_err != nil {
		log.Fatal("Error %v adding user to group in dynamodb", create_err, joinGroup)
		return false, create_err
	}
	log.Println("Added User to group:", joinGroup)

	return true, nil
}

func (g *GroupService) LeaveGroupService(leaveGroup groupModels.JoinGroupModel) (bool, error) {
	_, add_err := g.RemoveUserFromLeaderBoard(leaveGroup)
	if add_err != nil {
		log.Fatal("Error %v removing user from group in redis", add_err, leaveGroup)
		return false, add_err
	}

	_, err := g.groupRepo.RemoveUserFromGroup(leaveGroup, g.DynamodbSVC())

	if err != nil {
		log.Fatal("Error %v removing user from group in dynamodb", err, leaveGroup)
		return false, err
	}
	log.Println("Removed User from group:", leaveGroup)
	return false, nil
}

func (g *GroupService) GetGroupService(leaveGroup groupModels.JoinGroupModel) groupModels.Group {

	return g.groupRepo.GetOne(leaveGroup.GroupID, g.DynamodbSVC())
}

func (g *GroupService) createLeaderBoardInRedis(newGroup groupModels.Group) (bool, error) {
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

func (g *GroupService) AddUserToLeaderBoard(joinGroup groupModels.JoinGroupModel) (bool, error) {
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

func (g *GroupService) RemoveUserFromLeaderBoard(joinGroup groupModels.JoinGroupModel) (bool, error) {
	// Add a user with score to a group
	err := g.redisClient.ZRem(g.Ctx(), joinGroup.GroupID, joinGroup.Username).Err()
	if err != nil {
		return false, err
	}

	fmt.Println("User deleted from group successfully!")
	return true, nil
}

func (g *GroupService) DeleteGroup(joinGroup groupModels.JoinGroupModel) (bool, error) {
	// Remove a group
	err := g.redisClient.Del(g.Ctx(), joinGroup.GroupID).Err()
	if err != nil {
		return false, err
	}

	fmt.Println("Group removed successfully!")
	return true, nil
}
