package cache

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/k0kubun/pp"
)

func TestCache(t *testing.T) {
	conf := Config{
		Addr: "127.0.0.1:6379",
	}

	cc, err := conf.New()
	if err != nil {
		t.Fatal(err)
	}

	pp.Println(cc.Get("k1"))
	pp.Println(cc.Setx("k1", "v1", 0))
	pp.Println(cc.Get("k1"))

	pp.Println(cc.Setx("k2", "v2", 120))
	pp.Println(cc.Get("k2"))

	pp.Println(cc.Keys("*"))

	pp.Println(cc.Incr("n1"))
	pp.Println(cc.Incr("n1"))
	pp.Println(cc.GetInt("n1"))
	cc.Del("n1")
	pp.Println(cc.GetInt("n1"))
	r, err := cc.GetInt("n1")
	pp.Println(r, err, err == redis.ErrNil)
}
