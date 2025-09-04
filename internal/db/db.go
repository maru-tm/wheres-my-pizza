package db

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"restaurant-system/internal/config"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("PostgreSQL не отвечает: %w", err)
	}

	log.Println("✅ Подключение к PostgreSQL успешно")
	return &DB{Pool: pool}, nil
}

func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

func (d *DB) RunMigrations(ctx context.Context, migrationsDir string) error {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("не удалось прочитать директорию миграций: %w", err)
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}

	sort.Strings(sqlFiles)

	for _, fname := range sqlFiles {
		path := filepath.Join(migrationsDir, fname)

		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("ошибка чтения файла %s: %w", path, err)
		}

		query := string(sqlBytes)
		if strings.TrimSpace(query) == "" {
			continue
		}

		fmt.Printf("▶ Выполняется миграция: %s\n", fname)

		_, err = d.Pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("ошибка при выполнении миграции %s: %w", fname, err)
		}
	}

	fmt.Println("✅ Все миграции успешно применены")
	return nil
}
