package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	platformconfig "platform.local/platform/config"
)

// configuration and verifies the connection with a ping. Both the GORM handle
func ConnectWithPing(cfg platformconfig.DatabaseConfig) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql db: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	return db, sqlDB, nil
}

// ConnectRedis instantiates a Redis client and validates the connection with a
// ping.
func ConnectRedis(ctx context.Context, opts goredis.Options) (*goredis.Client, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	client := goredis.NewClient(&opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
