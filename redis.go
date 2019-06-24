package backTrace

import (
	"github.com/go-redis/redis"
)

var taskQueueName string

func CreateRedisClient() (*redis.Client, error) {
	host := gConf.String("redis::host") + ":" + gConf.String("redis::port")
	password := gConf.String("redis::password")
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password, // no password set
		DB:       0,        // use default DB
	})
	_, err := client.Ping().Result()
	return client, err
}

func init() {
	taskQueueName = gConf.String("redis::taskQueue")
}
