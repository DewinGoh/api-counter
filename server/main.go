package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type (
	RedisContext struct {
		echo.Context
		rdb *redis.Client
		ctx context.Context
	}
)

func (rctx *RedisContext) addCounter(c echo.Context) error {
	_, err := rctx.rdb.Incr(rctx.ctx, c.FormValue("testID")).Result()
	if err != nil {
		panic(err)
	}
	return c.JSON(http.StatusOK, "API call registered")
}

func (rctx *RedisContext) getCounter(c echo.Context) error {
	val, err := rctx.rdb.Get(rctx.ctx, c.FormValue("testID")).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(c.FormValue("testID"), val)
	return c.JSON(http.StatusOK, val)
}

func main() {
	redisAddr := flag.String("redisAddr", "", "URL of redis")
	redisDB := flag.Int("redisDB", 0, "Redis DB number") // use default DB
	port := flag.String("port", "8080", "Port on which service starts")
	flag.Parse()

	// create Server
	e := echo.New()

	// Middleware
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// // initialize redis client
	rctx := &RedisContext{
		rdb: redis.NewClient(&redis.Options{
			Addr: *redisAddr,
			DB:   *redisDB,
		}),
		ctx: context.Background()}

	// Routes
	e.POST("/counter", rctx.addCounter)
	e.GET("/counter", rctx.getCounter)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", *port)))
}
