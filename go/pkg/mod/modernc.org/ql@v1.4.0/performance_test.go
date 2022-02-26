package ql_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	_ "modernc.org/ql/driver" // driver
)

func TestPerformance(t *testing.T) {
	if os.Getenv("PERF") == "" {
		t.Skip("set PERF=1 to run this test")
	}

	var (
		seed        = envInt(t, "SEED", 123)           // rng seed
		insertCount = envInt(t, "INSERT_COUNT", 10000) // INSERT overall
		batchSize   = envInt(t, "BATCH_SIZE", 128)     // INSERT per transaction
		userCount   = envInt(t, "USER_COUNT", 8)       // unique object users
		idCount     = envInt(t, "ID_COUNT", 64)        // unique object IDs
		kindCount   = envInt(t, "KIND_COUNT", 4)       // unique object kinds
	)

	rand.Seed(int64(seed))

	d, err := newDatabase(t)
	if err != nil {
		t.Fatal(err)
	}

	keys := make([]key, insertCount)
	for i := range keys {
		keys[i] = key{
			user: fmt.Sprintf("user_%d", rand.Intn(userCount)),
			id:   fmt.Sprintf("id_%d", rand.Intn(idCount)),
			kind: fmt.Sprintf("kind_%d", rand.Intn(kindCount)),
		}
	}

	var (
		begin       = time.Now()
		ctx         = context.Background()
		objects     = map[key]object{}
		batch       = map[key]object{}
		sequence    = uint64(1)
		readCount   = uint64(0)
		writeCount  = uint64(0)
		deleteCount = uint64(0)
	)

	for _, k := range keys {
		now := time.Now()

		o := objects[k]
		{
			o.sequence = sequence
			sequence++

			o.accessCount.Valid = true
			o.accessCount.Int64 = o.accessCount.Int64 + 1
			o.lastAccessAt.Valid = true
			o.lastAccessAt.Time = now

			r := rand.Float64()
			switch {
			case r < 0.20:
				readCount++
				o.readCount.Valid = true
				o.readCount.Int64 = o.readCount.Int64 + 1
				o.lastReadAt.Valid = true
				o.lastReadAt.Time = now

			case r < 0.95:
				writeCount++
				o.size.Valid = true
				o.size.Int64 = 100 + int64(rand.Intn(100))
				o.writeCount.Valid = true
				o.writeCount.Int64 = o.writeCount.Int64 + 1
				o.lastWriteAt.Valid = true
				o.lastWriteAt.Time = now

			default:
				deleteCount++
				o.size.Valid = false
				o.size.Int64 = 0
				o.deleteCount.Valid = true
				o.deleteCount.Int64 = o.deleteCount.Int64 + 1
				o.lastDeleteAt.Valid = true
				o.lastDeleteAt.Time = now
			}
		}
		objects[k] = o
		batch[k] = o

		if len(batch) >= batchSize {
			if err := d.insertBatch(ctx, batch); err != nil {
				t.Fatal(err)
			}
			batch = map[key]object{}
		}
	}

	if len(batch) >= 0 {
		if err := d.insertBatch(ctx, batch); err != nil {
			t.Fatal(err)
		}
	}

	t.Logf("%d INSERT(s) in %s", insertCount, time.Since(begin))
	t.Logf("%d reads, %d writes, %d deletes", readCount, writeCount, deleteCount)
	t.Logf("final object count %d, row count %d", len(objects), d.selectCount(ctx))
}

//
//
//

type key struct {
	user string
	id   string
	kind string
}

type object struct {
	sequence     uint64
	size         sql.NullInt64
	accessCount  sql.NullInt64
	lastAccessAt sql.NullTime
	readCount    sql.NullInt64
	lastReadAt   sql.NullTime
	writeCount   sql.NullInt64
	lastWriteAt  sql.NullTime
	deleteCount  sql.NullInt64
	lastDeleteAt sql.NullTime
}

//
//
//

type database struct {
	t   *testing.T
	db  *sql.DB
	ins *sql.Stmt
	upd *sql.Stmt
	sel *sql.Stmt
}

func newDatabase(t *testing.T) (_ *database, err error) {
	db, err := sql.Open("ql-mem", randomDSN())
	if err != nil {
		return nil, fmt.Errorf("during Open: %w", err)
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("during Begin: %w", err)
	}

	if _, err := tx.Exec(`
		CREATE TABLE objects (
			user           string NOT NULL,
			id             string NOT NULL,
			kind           string NOT NULL,
			insert_at      time   NOT NULL DEFAULT timeIn(now(), "UTC"),
			operations     int64  NOT NULL DEFAULT 0,
			sequence       int64  NOT NULL DEFAULT 0,
			size           int64,
			access_count   int64,
			last_access_at time,
			read_count     int64,
			last_read_at   time,
			write_count    int64,
			last_write_at  time,
			delete_count   int64,
			last_delete_at time,
		);
	`); err != nil {
		return nil, fmt.Errorf("during CREATE TABLE: %w", err)
	}

	for _, s := range []string{
		"CREATE UNIQUE INDEX index_key      ON objects (user, id, kind);",
		"CREATE        INDEX index_access   ON objects (last_access_at);",
		"CREATE        INDEX index_size     ON objects (size);",
		"CREATE        INDEX index_write    ON objects (last_write_at);",
		"CREATE        INDEX index_delete   ON objects (last_delete_at);",
	} {
		if _, err := tx.Exec(s); err != nil {
			return nil, fmt.Errorf("during CREATE INDEX: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("during Commit: %w", err)
	}

	ins, err := db.Prepare(`
		INSERT INTO objects
		IF NOT EXISTS
			(user, id, kind)
		VALUES
			($1, $2, $3);
	`)
	if err != nil {
		return nil, fmt.Errorf("during PREPARE for INSERT: %w", err)
	}

	upd, err := db.Prepare(`
		UPDATE objects
		SET
			operations    = 1 + operations,
			sequence      = $1,
			size          = $2,
			access_count  = $3,
			last_access_at= $4,
			read_count    = $5,
			last_read_at  = $6,
			write_count   = $7,
			last_write_at = $8,
			delete_count  = $9,
			last_delete_at= $10,
		WHERE
			    user == $11
			AND id   == $12
			AND kind == $13
			AND $1 > sequence;
	`)
	if err != nil {
		return nil, fmt.Errorf("during PREPARE for UPDATE: %w", err)
	}

	sel, err := db.Prepare(`
		SELECT
			user,
			id,
			kind,
			insert_at,
			operations,
			sequence,
			size,
			access_count,
			last_access_at,
			read_count,
			last_read_at,
			write_count,
			last_write_at,
			delete_count,
			last_delete_at,
		FROM objects
		WHERE user == $1 AND id == $2 AND kind == $3;
	`)
	if err != nil {
		return nil, fmt.Errorf("during PREPARE for SELECT: %w", err)
	}

	return &database{
		t:   t,
		db:  db,
		ins: ins,
		upd: upd,
		sel: sel,
	}, nil
}

func (d *database) selectCount(ctx context.Context) int {
	var n int
	d.db.QueryRowContext(ctx, `SELECT count() FROM objects;`).Scan(&n)
	return n
}

func (d *database) insertBatch(ctx context.Context, batch map[key]object) (err error) {
	if len(batch) <= 0 {
		return nil
	}

	defer func(begin time.Time) {
		var (
			rowCount     = d.selectCount(ctx)
			perBatch     = perUnit(time.Since(begin), 1).Truncate(time.Microsecond)
			perRecord    = perUnit(perBatch, len(batch)).Truncate(time.Microsecond)
			perRecordRow = perUnit(perRecord, rowCount)
		)
		d.t.Helper()
		d.t.Logf(
			"batch size %4d · row count %5d · per batch %10s · per record %10s · per record-row %10s · error %v",
			len(batch), rowCount, perBatch, perRecord, perRecordRow, err,
		)
	}(time.Now())

	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			err = fmt.Errorf("%w (rollback err=%v)", err, tx.Rollback())
		}
	}()

	var (
		ins = tx.Stmt(d.ins)
		upd = tx.Stmt(d.upd)
	)
	for k, o := range batch {
		if _, err := ins.Exec(
			k.user,
			k.id,
			k.kind,
		); err != nil {
			return fmt.Errorf("INSERT: %w", err)
		}
		if _, err := upd.Exec(
			o.sequence,     // $1
			o.size,         // $2
			o.accessCount,  // $3
			o.lastAccessAt, // $4
			o.readCount,    // $5
			o.lastReadAt,   // $6
			o.writeCount,   // $7
			o.lastWriteAt,  // $8
			o.deleteCount,  // $9
			o.lastDeleteAt, // $10
			k.user,         // $11
			k.id,           // $12
			k.kind,         // $13
		); err != nil {
			return fmt.Errorf("UPDATE: %w", err)
		}
	}

	return nil
}

//
//
//

var unique uint64

func randomDSN() string {
	cnt := atomic.AddUint64(&unique, 1)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, 32)
	n, _ := rng.Read(buf)
	return fmt.Sprintf("memory://%d-%x", cnt, buf[:n])
}

func perUnit(d time.Duration, n int) time.Duration {
	if n == 0 {
		n = 1
	}
	return d / time.Duration(n)
}

func envInt(t *testing.T, key string, def int) int {
	t.Helper()
	i, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		i = def
	}
	t.Logf("%s=%d", key, i)
	return i
}
