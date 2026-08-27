package store

import (
	"context"
	"fmt"
	"time"

	"github.com/finance/seed/internal/sources"
)

// SaveBars upserts a slice of bars into the ohlcv table.
func (d *DB) SaveBars(ctx context.Context, bars []sources.Bar) (int, error) {
	if len(bars) == 0 {
		return 0, nil
	}

	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ohlcv (symbol, source, interval, ts, open, high, low, close, volume, adj_close)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (symbol, source, interval, ts) DO UPDATE SET
			open      = excluded.open,
			high      = excluded.high,
			low       = excluded.low,
			close     = excluded.close,
			volume    = excluded.volume,
			adj_close = excluded.adj_close
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	saved := 0
	for _, b := range bars {
		if _, err := stmt.ExecContext(ctx,
			b.Symbol, b.Source, b.Interval, b.TS,
			b.Open, b.High, b.Low, b.Close, b.Volume, b.AdjClose,
		); err != nil {
			return saved, fmt.Errorf("insert bar %s@%s: %w", b.Symbol, b.TS, err)
		}
		saved++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return saved, nil
}

// QueryBars retrieves bars for a symbol/interval in [from, to] from all sources,
// ordered by timestamp ascending.
func (d *DB) QueryBars(ctx context.Context, symbol, interval string, from, to time.Time) ([]sources.Bar, error) {
	rows, err := d.conn.QueryContext(ctx, `
		SELECT symbol, source, interval, ts, open, high, low, close, volume, adj_close
		FROM ohlcv
		WHERE symbol = ? AND interval = ? AND ts >= ? AND ts <= ?
		ORDER BY ts ASC
	`, symbol, interval, from, to)
	if err != nil {
		return nil, fmt.Errorf("query bars: %w", err)
	}
	defer rows.Close()

	var bars []sources.Bar
	for rows.Next() {
		var b sources.Bar
		var adjClose *float64
		if err := rows.Scan(
			&b.Symbol, &b.Source, &b.Interval, &b.TS,
			&b.Open, &b.High, &b.Low, &b.Close, &b.Volume, &adjClose,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		if adjClose != nil {
			b.AdjClose = *adjClose
		}
		bars = append(bars, b)
	}
	return bars, rows.Err()
}

// LatestBar returns the most recent bar for a symbol/interval across all sources.
func (d *DB) LatestBar(ctx context.Context, symbol, interval string) (*sources.Bar, error) {
	row := d.conn.QueryRowContext(ctx, `
		SELECT symbol, source, interval, ts, open, high, low, close, volume, adj_close
		FROM ohlcv
		WHERE symbol = ? AND interval = ?
		ORDER BY ts DESC
		LIMIT 1
	`, symbol, interval)

	var b sources.Bar
	var adjClose *float64
	err := row.Scan(
		&b.Symbol, &b.Source, &b.Interval, &b.TS,
		&b.Open, &b.High, &b.Low, &b.Close, &b.Volume, &adjClose,
	)
	if err != nil {
		return nil, err
	}
	if adjClose != nil {
		b.AdjClose = *adjClose
	}
	return &b, nil
}
