package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

type RedisCfg struct {
	Addr         string
	Username     string
	Password     string
	Db           int
	PoolSize     int
	MinIdleConns int
	PoolTimeout  time.Duration
	UseTLS       bool
}

const pingTimeout = 5 * time.Second

func NewRedis(rc RedisCfg) (*Redis, error) {

	err := validateAddr(rc.Addr)
	if err != nil {
		return nil, err
	}

	opts := &redis.Options{
		Addr:     rc.Addr,
		Username: rc.Username,
		Password: rc.Password,
		DB:       rc.Db,
	}

	if rc.PoolSize > 0 {
		opts.PoolSize = rc.PoolSize
	}
	if rc.MinIdleConns > 0 {
		opts.MinIdleConns = rc.MinIdleConns
	}
	if rc.PoolTimeout > 0 {
		opts.PoolTimeout = rc.PoolTimeout
	}

	if rc.UseTLS {
		opts.TLSConfig = &tls.Config{InsecureSkipVerify: false}
	}

	rdb := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &Redis{client: rdb, ctx: context.Background()}, nil
}

// test connection
func (r *Redis) Ping() error {
	ctx, cancel := context.WithTimeout(r.ctx, pingTimeout)
	defer cancel()
	_, err := r.client.Ping(ctx).Result()
	return err
}

func validateAddr(dsn string) error {
	if strings.HasPrefix(dsn, "redis://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return fmt.Errorf("invalid Redis URL: %w", err)
		}
		if u.Host == "" {
			return errors.New("Redis DSN must include host")
		}
	} else {
		host, port, err := net.SplitHostPort(dsn)
		if err != nil || host == "" || port == "" {
			return fmt.Errorf("invalid Redis address: %w", err)
		}
	}
	return nil
}

// test actual usability of redis
func (r *Redis) IsReady() error {
	testKey := fmt.Sprintf("%s_%s", r.Options().ClientName, "redis_health_check")
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	if err := r.client.Set(ctx, testKey, "ok", 0).Err(); err != nil {
		return err
	}
	val, err := r.client.Get(ctx, testKey).Result()
	if err != nil {
		return err
	}
	if val != "ok" {
		return fmt.Errorf("unexpected Redis test value: %s", val)
	}
	return nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Options() *redis.Options {
	return r.client.Options()
}

func (r *Redis) WithContext(ctx context.Context) *Redis {
	return &Redis{
		client: r.client,
		ctx:    ctx,
	}
}

func (r *Redis) Client() *redis.Client {
	return r.client
}
