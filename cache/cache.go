package cache

import (
	"github.com/gomodule/redigo/redis"
)

type Cache struct {
	conf Config

	pool *redis.Pool
}

type Config struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	MaxConns int    `mapstructure:"max_conns"`
}

func (c Config) New() (*Cache, error) {
	pool := &redis.Pool{
		// Other pool configuration not shown in this example.
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", c.Addr)
			if err != nil {
				return nil, err
			}
			if c.Password != "" {
				if _, err := conn.Do("AUTH", c.Password); err != nil {
					conn.Close()
					return nil, err
				}
			}

			if _, err := conn.Do("SELECT", c.DB); err != nil {
				conn.Close()
				return nil, err
			}
			return conn, nil
		},
	}
	cn := pool.Get()
	if cn.Err() != nil {
		return nil, cn.Err()
	}
	defer cn.Close()

	cache := &Cache{
		pool: pool,
	}
	return cache, nil
}

func (c *Cache) Get(key string) (val string, err error) {
	conn := c.pool.Get()
	if conn.Err() != nil {
		return "", conn.Err()
	}
	defer conn.Close()

	val, err = redis.String(conn.Do("GET", key))
	return
}

func (c *Cache) TTL(key string, ttl int) (err error) {
	conn := c.pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err = conn.Do("Expire", key, ttl)
	return
}

func (c *Cache) GetInt(key string) (val int, err error) {
	conn := c.pool.Get()
	if conn.Err() != nil {
		return 0, conn.Err()
	}
	defer conn.Close()

	val, err = redis.Int(conn.Do("GET", key))
	return
}

func (c *Cache) Setx(key string, val string, expire int) (err error) {
	conn := c.pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()
	if expire > 0 {
		_, err = conn.Do("SETEX", key, expire, val)
	} else {
		_, err = conn.Do("SET", key, val)
	}
	return
}

func (c *Cache) Incr(key string) (v int, err error) {
	conn := c.pool.Get()
	if conn.Err() != nil {
		return 0, conn.Err()
	}
	defer conn.Close()

	v, err = redis.Int(conn.Do("incr", key))
	return
}

func (c *Cache) Del(key string) error {
	conn := c.pool.Get()
	if conn.Err() != nil {
		return conn.Err()
	}
	defer conn.Close()

	_, err := conn.Do("del", key)
	return err
}

func (c *Cache) Keys(keyPattern string) (v []string, err error) {
	conn := c.pool.Get()
	if conn.Err() != nil {
		return nil, conn.Err()
	}
	defer conn.Close()

	v, err = redis.Strings(conn.Do("keys", keyPattern))
	return
}

func (c *Cache) Close() {
	c.pool.Close()
	return
}
