package main

import (
	"github.com/PraveenPin/GroupService/init_database"
	"github.com/PraveenPin/GroupService/routes"
)

func main() {
	app := &init_database.App{}
	ctx := app.GetAppContext()
	session := app.StartAWSSession()
	dynamoDBSvc := app.GetDynamoDatabaseClient(session)
	redisClient := app.GetRedisClient(ctx)
	pubsubClient := app.GetPubSubClient(ctx)

	//redis database utils
	redisDB := &init_database.RedisDatabase{}
	redisDB.GetLeaderBoardByGroup(ctx, redisClient, "JTcWuWP3Wkug22d3XM5KS5")
	//redisDB.ClearDB(ctx, redisClient)
	//redisDB.GetUserWithScoresInGroup(ctx, redisClient, "JTcWuWP3Wkug22d3XM5KS5")
	//redisDB.GetScoreByAUserInAGroup(ctx, redisClient, "JTcWuWP3Wkug22d3XM5KS5", "praveenpin-1")
	//redisDB.AddScoreToAUserInAGroup(ctx, redisClient, "JTcWuWP3Wkug22d3XM5KS5", "praveenpin-1", 9)

	dispatcher := routes.Dispatcher{}

	dispatcher.Init(dynamoDBSvc, redisClient, ctx, pubsubClient)

}
