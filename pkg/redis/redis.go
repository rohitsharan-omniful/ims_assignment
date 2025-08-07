package redis

import "github.com/omniful/go_commons/redis"

type Redis struct {
	*redis.Client
}

var redisInstance *Redis

func GetClient() *Redis {
	return redisInstance
}

func SetClient(client *redis.Client) {
	redisInstance = &Redis{client}
}
