package helpers

import (
	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"log"
	"sync"
	"time"
)

const (
	FREE    = "free"
	BLOCKED = "block"
)

var pool *pools.ResourcePool

func InitRedisPool() {
	pool = pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial("tcp", ":6379")
		if err != nil {
			log.Println(err)
		}

		return ResourceConn{c}, err
	}, 30, 31, time.Minute)
}

// ResourceConn adapts a Redigo connection to a Vitess Resource.
type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}

func GetConnection() ResourceConn {
	connection, err := pool.Get()
	if err != nil {
		log.Println(err)
	}

	return connection.(ResourceConn)
}

func ReturnConnection(c ResourceConn) {
	pool.Put(c)
}

func BlockSeat(seatkey string, syncWait *sync.WaitGroup) {
	defer syncWait.Done()

	c := GetConnection()
	defer pool.Put(c)

	if exists, _ := redis.Int(c.Do("EXISTS", seatkey)); exists == 0 {
		log.Println("No seat with key: ", seatkey)
		return
	}

	c.Do("WATCH", seatkey)
	status, serr := redis.String(c.Do("GET", seatkey))
	if serr != nil {
		log.Println("Serr: ", seatkey, serr)
	}

	if status == FREE {
		c.Do("MULTI")
		c.Do("SET", seatkey, FREE)
		_, err := c.Do("EXEC")
		if err != nil {
			log.Println(err)
		}
	} else {
		c.Do("UNWATCH")
	}
}
