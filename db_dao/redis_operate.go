package db_dao

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"net"
)

type RedisConfig struct {
	address string
	port uint16
	dbName string
}

func (conf *RedisConfig)InitRedisConfig(address string, port uint16, dbName string) {
	conf.address = address
	conf.port = port
	conf.dbName = dbName
}

func (conf *RedisConfig)OpenDatabase() (conn interface{}, err error) {
	var address string
	fmt.Sprintf(address, "%s:%d", conf.address, conf.port)
	redisConn, err := redis.Dial("tcp", "127.0.0.1:6379")
	// 若连接出错，则打印错误信息，返回
	if err != nil {
		fmt.Println(err)
		fmt.Println("redis connect error")
		return nil, err
	}
	redisConn.Do("SELECT", conf.dbName)
	return redisConn, nil
}

func (conf *RedisConfig)CloseDatabase(conn interface{}) int {
	redisConn, ok := conn.(net.Conn)
	if ok {
		redisConn.Close()
	} else {
		return -1
	}
	return 0
}