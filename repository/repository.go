package repository

import (
	"context"
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type ClientRepository interface {
	InsertClient(ip string, maxTokens uint) (*Client, error)
	GetClient(ctx context.Context, ip string) (*Client, error)
	DecreaseBucket(cl *Client, ip string) error
	IncreaseBucket(cl *Client, ip string, qtd uint) error
	UpdateBucket(client *Client, ip string, secondsBetween float64, maxTokens uint) error
}

type clientRepository struct {
	RedisCl *redis.Client
}

func NewClientRepository(cl *redis.Client) ClientRepository {
	return &clientRepository{
		RedisCl: cl,
	}
}

type Client struct {
	Tokens uint   `json:"tokens"`
	Ts     string `json:"ts"`
}

func (cl Client) MarshalBinary() ([]byte, error) {
	return json.Marshal(cl)
}

func parseClient(val string) (*Client, error) {
	client := Client{}
	if err := json.Unmarshal([]byte(val), &client); err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *clientRepository) InsertClient(ip string, maxTokens uint) (*Client, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	cl, err := json.Marshal(&Client{Tokens: maxTokens, Ts: ts})
	if err != nil {
		return nil, err
	}

	err = r.RedisCl.Set(context.Background(), ip, cl, 0).Err()
	if err != nil {
		return nil, err
	}
	client, err := r.GetClient(context.Background(), ip)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (r *clientRepository) GetClient(ctx context.Context, ip string) (*Client, error) {
	val, err := r.RedisCl.Get(ctx, ip).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		} else {
			return nil, err
		}
	}
	client, err := parseClient(val)
	if err != nil {
		return nil, err
	}
	return client, nil

}

func (r *clientRepository) DecreaseBucket(cl *Client, ip string) error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	cl.Tokens = cl.Tokens - 1
	cl.Ts = ts
	err := r.RedisCl.Set(context.Background(), ip, cl, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *clientRepository) IncreaseBucket(cl *Client, ip string, qtd uint) error {
	cl.Tokens = cl.Tokens + qtd
	err := r.RedisCl.Set(context.Background(), ip, cl, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *clientRepository) UpdateBucket(client *Client, ip string, secondsBetween float64, maxTokens uint) error {
	tsInt, err := strconv.ParseInt(client.Ts, 10, 64)
	if err != nil {
		return err
	}
	t1 := time.Unix(tsInt, 0)
	currentTime := time.Now().Unix()
	t2 := time.Unix(currentTime, 0)
	diff := t2.Sub(t1).Seconds()
	roundDiff := math.Round(diff / secondsBetween)
	if roundDiff > 0 {
		if roundDiff >= float64(maxTokens) {
			r.IncreaseBucket(client, ip, maxTokens)
		} else {
			r.IncreaseBucket(client, ip, uint(roundDiff))
		}
	}
	return nil
}
