package main

import (
	"github.com/PraveenPin/GroupService/init_database"
	"github.com/PraveenPin/GroupService/routes"
)

func main() {
	app := &init_database.App{}
	session := app.StartAWSSession()
	dynamoDBSvc := app.GetDynamoDatabaseClient(session)

	dispatcher := routes.Dispatcher{}
	dispatcher.Init(dynamoDBSvc)

}
