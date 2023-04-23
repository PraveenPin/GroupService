package repo

import (
	"errors"
	"fmt"
	"github.com/PraveenPin/GroupService/groupModels"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
)

const (
	authTableName  = "Authentication"
	groupTableName = "Groups"
)

type GroupRepoInterface interface {
	Create(group groupModels.Group, dynamoDBSvc *dynamodb.DynamoDB) (bool, error)
	Leave()
	Exit()
}

type GroupRepository struct{}

func (g *GroupRepository) GetOne(groupId string, dynamoDBSvc *dynamodb.DynamoDB) groupModels.Group {
	emptyGroup := groupModels.Group{}
	log.Println("Find the group with Id :", groupId)
	result, err := dynamoDBSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(groupTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"GroupID": {
				S: aws.String(groupId),
			},
		},
	})
	if err != nil {
		log.Fatalf("Got error calling get group: %s", err)
		return emptyGroup
	}

	if result.Item == nil {
		log.Fatalf("Could not find group with Id'" + groupId + "'")
		return emptyGroup
	}

	group := groupModels.Group{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &group)
	if err != nil {
		log.Fatalf("Failed to unmarshal Record, %v", err)
		return emptyGroup
	}

	log.Println("Group details")

	return group
}

func (g *GroupRepository) Create(group groupModels.Group, dynamoDBSvc *dynamodb.DynamoDB) (bool, error) {

	if len(group.LeaderBoard) == 0 {
		newLeaderBoardItem := groupModels.LeaderBoardItem{group.CreatedBy, 0.0}
		group.LeaderBoard = append(group.LeaderBoard, newLeaderBoardItem)
	}

	av, err := dynamodbattribute.MarshalMap(group)
	if err != nil {
		log.Fatalf("Got error marshalling new group item: %s", err)
		return false, nil
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(groupTableName),
	}

	_, err = dynamoDBSvc.PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
		return false, nil
	}
	log.Println("New Group %s inserted for user: %s", group.GroupName, group.CreatedBy)

	return true, nil
}

func (g *GroupRepository) AddUserToGroup(joinGroupObj groupModels.JoinGroupModel, dynamoDBSvc *dynamodb.DynamoDB) (bool, error) {
	log.Println("Find the group with Id :", joinGroupObj.GroupID)
	result, err := dynamoDBSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(groupTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"GroupID": {
				S: aws.String(joinGroupObj.GroupID),
			},
		},
	})
	if err != nil {
		msg := fmt.Sprintf("Got error calling get group: %s", err)
		return false, errors.New(msg)
	}

	if result.Item == nil {
		msg := "Could not find group with Id'" + joinGroupObj.GroupID + "'"
		return false, errors.New(msg)
	}

	group := groupModels.Group{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &group)
	if err != nil {
		msg := fmt.Sprintf("Failed to unmarshal Record, %v", err)
		return false, errors.New(msg)
	}

	//search for same user name in the group
	for _, v := range group.LeaderBoard {
		if v.Username == joinGroupObj.Username {
			errMsg := fmt.Sprintf("User already added to the group")
			return false, errors.New(errMsg)
		}
	}

	//number of users in group should not exceed USER_GROUP_LIMIT = 20
	if len(group.LeaderBoard) == groupModels.USER_GROUP_LIMIT {
		errMsg := fmt.Sprintf("Max User in a Group Limit Reached")
		return false, errors.New(errMsg)
	}

	log.Println("Previous list of users in group", group.LeaderBoard)
	group.LeaderBoard = append(group.LeaderBoard, groupModels.LeaderBoardItem{joinGroupObj.Username, 0.0})
	log.Println("New list of users in group", group.LeaderBoard)

	av, err := dynamodbattribute.MarshalMap(group)
	if err != nil {
		msg := fmt.Sprintf("Got error marshalling new group item: %s", err)
		return false, errors.New(msg)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(groupTableName),
	}

	_, err = dynamoDBSvc.PutItem(input)
	if err != nil {
		msg := fmt.Sprintf("Got error calling PutItem: %s", err)
		return false, errors.New(msg)
	}
	log.Println("New user:", joinGroupObj.Username, "added to the group ", group.GroupName)

	return true, nil
}

func (g *GroupRepository) RemoveUserFromGroup(leaveGroupObj groupModels.JoinGroupModel, dynamoDBSvc *dynamodb.DynamoDB) (bool, error) {
	log.Println("Find the group with Id for removing user :", leaveGroupObj.GroupID)

	result, err := dynamoDBSvc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(groupTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"GroupID": {
				S: aws.String(leaveGroupObj.GroupID),
			},
		},
	})
	if err != nil {
		msg := fmt.Sprintf("Got error calling get group: %s", err)
		return false, errors.New(msg)
	}

	if result.Item == nil {
		msg := "Could not find group with Id'" + leaveGroupObj.GroupID + "'"
		return false, errors.New(msg)
	}

	group := groupModels.Group{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &group)
	if err != nil {
		msg := fmt.Sprintf("Failed to unmarshal Record, %v", err)
		return false, errors.New(msg)
	}

	//search for same user name in the group
	found := -1
	for index, v := range group.LeaderBoard {
		if v.Username == leaveGroupObj.Username {
			found = index
			break
		}
	}

	if found == -1 {
		errMsg := fmt.Sprintf("User does not exist in the group")
		return false, errors.New(errMsg)
	} else {
		group.LeaderBoard[found] = group.LeaderBoard[len(group.LeaderBoard)-1]
		group.LeaderBoard = group.LeaderBoard[:len(group.LeaderBoard)-1]
	}

	//number of users in group is zero, delete group
	if len(group.LeaderBoard) == 0 {
		//call detele function
		return g.DeleteGroup(group, dynamoDBSvc)
	}

	av, err := dynamodbattribute.MarshalMap(group)
	if err != nil {
		msg := fmt.Sprintf("Got error marshalling new group item: %s", err)
		return false, errors.New(msg)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(groupTableName),
	}

	_, err = dynamoDBSvc.PutItem(input)
	if err != nil {
		msg := fmt.Sprintf("Got error calling PutItem: %s", err)
		return false, errors.New(msg)
	}
	log.Println("Removed user:", leaveGroupObj.Username, "added to the group ", group.GroupName)

	return true, nil
}

func (g *GroupRepository) DeleteGroup(group groupModels.Group, dynamoDBSvc *dynamodb.DynamoDB) (bool, error) {

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"GroupID": {
				S: aws.String(group.GroupID),
			},
		},
		TableName: aws.String(groupTableName),
	}

	_, err := dynamoDBSvc.DeleteItem(input)
	if err != nil {
		msg := fmt.Sprintf("Got error calling DeleteItem: %s", err)
		return false, errors.New(msg)
	}

	fmt.Println("Deleted '" + group.GroupName + "from table " + groupTableName)
	return true, nil
}
