package rating

import (
	"context"
	"database/sql"
	"fmt"
)

// RatingRepository handles all database operations for ratings.
type RatingRepository struct {
	DB *sql.DB
}

// CategoryScoresQuery executes the category scores query and returns the raw rows.
func (r *RatingRepository) CategoryScoresQuery(ctx context.Context, startDate string, endDate string, dateExpr string, groupBy string) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		SELECT
			rc.name AS category,
			%s AS period,
			ROUND(100.0 * SUM(r.rating * rc.weight) / NULLIF(SUM(5.0 * rc.weight), 0), 2) AS score,
			COUNT(*) as ratings_count
		FROM ratings r
		JOIN rating_categories rc ON r.rating_category_id = rc.id
		WHERE r.created_at BETWEEN ? AND ?
		GROUP BY %s
		ORDER BY period asc;
	`, dateExpr, groupBy)
	return r.DB.QueryContext(ctx, query, startDate, endDate)
}

// TicketScoresQuery executes the ticket scores query and returns the raw rows.
func (r *RatingRepository) TicketScoresQuery(ctx context.Context, startDate string, endDate string) (*sql.Rows, error) {
	query := `
		SELECT
			r.ticket_id,
			rc.name AS category,
			ROUND(100.0 * SUM(r.rating * rc.weight) / NULLIF(SUM(5.0 * rc.weight), 0), 2) AS score
		FROM ratings r
		JOIN rating_categories rc ON r.rating_category_id = rc.id
		WHERE r.created_at BETWEEN ? AND ?
		GROUP BY r.ticket_id, rc.name
		ORDER BY r.ticket_id, rc.name;
	`
	return r.DB.QueryContext(ctx, query, startDate, endDate)
}

// OverallScoreQuery executes the overall score query and returns the result row.
func (r *RatingRepository) OverallScoreQuery(ctx context.Context, startDate string, endDate string) *sql.Row {
	query := `
		SELECT ROUND(
			100.0 * SUM(r.rating * rc.weight) / NULLIF(SUM(5.0 * rc.weight), 0),
			2
		) AS score
		FROM ratings r
		JOIN rating_categories rc ON r.rating_category_id = rc.id
		WHERE r.created_at BETWEEN ? AND ?;
	`
	return r.DB.QueryRowContext(ctx, query, startDate, endDate)
}

// ScoreChangeQuery executes the score change query for a given range and returns the result row.
func (r *RatingRepository) ScoreChangeQuery(ctx context.Context, start string, end string) *sql.Row {
	query := `
		SELECT ROUND(
			100.0 * SUM(r.rating * rc.weight) / NULLIF(SUM(5.0 * rc.weight), 0),
			2
		) AS score
		FROM ratings r
		JOIN rating_categories rc ON r.rating_category_id = rc.id
		WHERE DATE(r.created_at) BETWEEN ? AND ?;
	`
	return r.DB.QueryRowContext(ctx, query, start, end)
}
