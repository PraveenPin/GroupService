package init_database

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
)

type RedisDatabase struct {
}

func (rd *RedisDatabase) GetScoreByAUserInAGroup(ctx context.Context, client *redis.Client, groupName string, userName string) {
	score := client.ZRank(ctx, groupName, userName)
	fmt.Println("User, group, score :", userName, groupName, score.Val())
	return
}
func (rd *RedisDatabase) AddScoreToAUserInAGroup(ctx context.Context, client *redis.Client, groupName string, userName string, scoreToAdd float64) {

	res, err := client.ZIncrBy(ctx, groupName, scoreToAdd, userName).Result()

	if err != nil {
		log.Fatalf("Error adding user score", err)
	}

	log.Println("Score added to user in a group successfully!", res)
	score := client.ZRank(ctx, groupName, userName)
	fmt.Println("User, group, score :", userName, groupName, score.Val())
	return
}

func (rd *RedisDatabase) GetUserWithScoresInGroup(ctx context.Context, client *redis.Client, groupName string) {
	// Get all users in a specific group from the database
	userScores, err := client.ZRangeWithScores(ctx, groupName, 0, -1).Result()
	if err != nil {
		log.Fatalf("Error obtaining users with scores ", err)
	}

	// Print all users with scores in the group to the console
	fmt.Printf("Users in group %s with scores:\n", groupName)
	for _, userScore := range userScores {
		fmt.Printf("%s: %f\n", userScore.Member, userScore.Score)
	}
}

func (rd *RedisDatabase) GetEverything(ctx context.Context, client *redis.Client) {
	var cursor uint64
	var keys []string
	var err error

	for {
		// Scan all keys matching the pattern "*"
		keys, cursor, err = client.Scan(ctx, cursor, "*", 1000).Result()
		if err != nil {
			panic(err)
		}

		// Print each key with its data type
		for _, key := range keys {
			fmt.Println("value is: ", key)
			keyType, err := client.Type(ctx, key).Result()
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s: %s\n", key, keyType)
			members, err := client.ZRangeWithScores(ctx, "somegrop", 0, -1).Result()
			if err != nil {
				panic(err)
			}

			// Print the group key and its associated user IDs and scores
			fmt.Printf("%s:\n", key)
			for _, member := range members {
				fmt.Printf("  %s: %v\n", member.Member, member.Score)
			}
		}

		// Exit loop when cursor is 0
		if cursor == 0 {
			break
		}
	}
}

func (rd *RedisDatabase) ClearDB(ctx context.Context, client *redis.Client) {
	// flush the database
	_, err := client.FlushDB(ctx).Result()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Database flushed successfully")
	}

	// close the Redis client
	//err = client.Close()
	//if err != nil {
	//	fmt.Println(err)
	//}
}

func (rd *RedisDatabase) GetLeaderBoardByGroup(ctx context.Context, client *redis.Client, groupId string) {
	// retrieve the top 10 scores from the leaderboard
	scores, err := client.ZRevRangeWithScores(ctx, groupId, 0, 19).Result()
	if err != nil {
		panic(err)
	}

	// print the leaderboard scores
	for _, score := range scores {
		fmt.Printf("%s: %f\n", score.Member, score.Score)
	}
}
