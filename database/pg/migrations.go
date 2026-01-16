package pg

import (
	"errors"
	"io/fs"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Migrate(db *Database, sourceFS fs.FS, dir string) error {
	if sourceFS == nil {
		return errors.New("sourceFS required")
	}

	goose.SetDialect("postgres")
	goose.SetBaseFS(sourceFS)

	sqlDB := stdlib.OpenDBFromPool(db.Pool)
	defer sqlDB.Close()

	db.log.Info("running database migrations from", "dir", dir)
	
	err := goose.Up(sqlDB, dir)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			db.log.Info("no new migrations to apply")
			return nil
		}
		return err
	}

	db.log.Info("migrations completed successfully")
	return nil
}
