# token bucket rate limiter

a simple rate limiter using the token bucket strategy to be used as a http middleware in a Golang API.

## Installation

To use this package, install it using go get:

```bash
go get github.com/awmpietro/token-bucket-rate-limiter
```

## Usage:

First, import the package:

```go
import "github.com/awmpietro/token-bucket-rate-limiter"
```

Create an instance of NewLimiter and apply as a middleware to a handler:

```go
/**
/* The params for NewLimiter are:
/* max uint - max number of tokens each IP receives
/* sb uint - how many seconds between each request should refill 1 token to the user's bucket
/* redisClient *redis.Client - an instance of github.com/redis/go-redis/v9 redis client.
*/
// main.go
redisCl = config.NewRedisClient("localhost", 6379, "", 0)
defer redisCl.Close()
mux := http.NewServeMux()
rl := tokenbucketratelimiter.NewLimiter(10, 1, redisCl)
mux.Handle("/token-bucket", rl.RateLimiterMiddleware(http.HandlerFunc(tokenBucket))) //tokenBucket is a handler

log.Fatal(http.ListenAndServe(":8080", mux))
```

In the provided code each IP address will be assigned a max of 10 tokens. Each request consumes 1 token, and 1 token is refilled every 1 second until max of 10.

## TODO

Unit tests

## Contributing

Feel free to open issues or PRs if you find any problems or have suggestions!
