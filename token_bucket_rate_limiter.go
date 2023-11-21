package tokenbucketratelimiter

import (
	"context"
	"errors"
	"net/http"

	"github.com/awmpietro/token-bucket-rate-limiter/repository"
	"github.com/awmpietro/token-bucket-rate-limiter/util"

	"github.com/redis/go-redis/v9"
)

type Limiter interface {
	RateLimiterMiddleware(next http.Handler) http.Handler
}

type limiter struct {
	rp             repository.ClientRepository
	maxTokens      uint
	secondsBetween float64
}

type TokensKey string

const (
	Tokens TokensKey = "tokens"
)

func NewLimiter(max uint, sb float64, redisClient *redis.Client) Limiter {
	var redisCl *redis.Client
	if redisClient != nil {
		redisCl = redisClient
	}

	rp := repository.NewClientRepository(redisCl)
	return &limiter{
		rp:             rp,
		maxTokens:      max,
		secondsBetween: sb,
	}
}

func (l *limiter) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, Tokens, 0)
		ip, err := util.GetIp(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		client, err := l.rp.GetClient(ctx, ip)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if client == nil {
			client, err = l.rp.InsertClient(ip, l.maxTokens)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		} else {
			l.rp.UpdateBucket(client, ip, l.secondsBetween, l.maxTokens)
			if client.Tokens == 0 {
				http.Error(w, errors.New("quota exceeded").Error(), http.StatusTooManyRequests)
				return
			}
		}
		if err := l.rp.DecreaseBucket(client, ip); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		ctx = context.WithValue(ctx, Tokens, client.Tokens)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
