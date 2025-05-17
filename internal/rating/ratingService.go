package rating

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	pb "zendesk_grpc_service/proto/ratingService/v1"
)

// RatingService implements pb.RatingServiceServer and contains business logic.
type RatingService struct {
	pb.UnimplementedRatingServiceServer
	Repo *RatingRepository
}

// CategoryScores returns scores for each category, grouped by day or week depending on the date range.
func (s *RatingService) CategoryScores(ctx context.Context, req *pb.DateRangeRequest) (*pb.CategoryScoresResponse, error) {
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}
	duration := end.Sub(start)

	var dateExpr, groupBy string
	if duration.Hours() > 24*31 {
		dateExpr = "strftime('%W', r.created_at)"
		groupBy = "category, period"
	} else {
		dateExpr = "strftime('%Y-%m-%d', r.created_at)"
		groupBy = "category, period"
	}

	rows, err := s.Repo.CategoryScoresQuery(ctx, req.StartDate, req.EndDate, dateExpr, groupBy)
	if err != nil {
		return nil, fmt.Errorf("query category scores: %w", err)
	}
	defer rows.Close()

	type categoryData struct {
		scores       map[string]float64
		ratingsCount int32
		totalScore   float64
		periodsCount int32
	}
	categoryMap := make(map[string]*categoryData)

	for rows.Next() {
		var category, period string
		var scoreNull sql.NullFloat64
		var count int32
		if err := rows.Scan(&category, &period, &scoreNull, &count); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		score := 0.0
		if scoreNull.Valid {
			score = scoreNull.Float64
		}
		if _, exists := categoryMap[category]; !exists {
			categoryMap[category] = &categoryData{scores: make(map[string]float64)}
		}
		categoryMap[category].scores[period] = score
		categoryMap[category].ratingsCount += count
		categoryMap[category].totalScore += score
		categoryMap[category].periodsCount++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category rows: %w", err)
	}

	var response pb.CategoryScoresResponse
	for cat, data := range categoryMap {
		avgScore := 0.0
		if data.periodsCount > 0 {
			avgScore = data.totalScore / float64(data.periodsCount)
		}
		response.Categories = append(response.Categories, &pb.CategoryScore{
			Category:     cat,
			ScoresByDate: data.scores,
			AverageScore: avgScore,
			RatingsCount: data.ratingsCount,
		})
	}
	return &response, nil
}

// TicketScores returns scores for each ticket and category in the given date range.
func (s *RatingService) TicketScores(ctx context.Context, req *pb.DateRangeRequest) (*pb.TicketScoresResponse, error) {
	rows, err := s.Repo.TicketScoresQuery(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("query ticket scores: %w", err)
	}
	defer rows.Close()

	ticketMap := make(map[string]map[string]float64)
	for rows.Next() {
		var ticketID, category string
		var scoreNull sql.NullFloat64
		if err := rows.Scan(&ticketID, &category, &scoreNull); err != nil {
			return nil, fmt.Errorf("scan ticket row: %w", err)
		}
		score := 0.0
		if scoreNull.Valid {
			score = scoreNull.Float64
		}
		if _, exists := ticketMap[ticketID]; !exists {
			ticketMap[ticketID] = make(map[string]float64)
		}
		ticketMap[ticketID][category] = score
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticket rows: %w", err)
	}

	var response pb.TicketScoresResponse
	for ticketID, categories := range ticketMap {
		response.Tickets = append(response.Tickets, &pb.TicketScore{
			TicketId:       ticketID,
			CategoryScores: categories,
		})
	}
	return &response, nil
}

// OverallScore returns the overall score for the given date range.
func (s *RatingService) OverallScore(ctx context.Context, req *pb.DateRangeRequest) (*pb.ScoreResponse, error) {
	row := s.Repo.OverallScoreQuery(ctx, req.StartDate, req.EndDate)
	var score sql.NullFloat64
	if err := row.Scan(&score); err != nil {
		return nil, fmt.Errorf("query overall score: %w", err)
	}
	result := 0.0
	if score.Valid {
		result = score.Float64
	}
	return &pb.ScoreResponse{Score: result}, nil
}

// ScoreChange returns the change in score between two periods.
func (s *RatingService) ScoreChange(ctx context.Context, req *pb.ScoreChangeRequest) (*pb.ScoreChangeResponse, error) {
	currentRow := s.Repo.ScoreChangeQuery(ctx, req.CurrentStart, req.CurrentEnd)
	previousRow := s.Repo.ScoreChangeQuery(ctx, req.PreviousStart, req.PreviousEnd)

	var currentNull, previousNull sql.NullFloat64
	if err := currentRow.Scan(&currentNull); err != nil {
		return nil, fmt.Errorf("query current score: %w", err)
	}
	if err := previousRow.Scan(&previousNull); err != nil {
		return nil, fmt.Errorf("query previous score: %w", err)
	}

	current := 0.0
	if currentNull.Valid {
		current = currentNull.Float64
	}
	previous := 0.0
	if previousNull.Valid {
		previous = previousNull.Float64
	}

	change := current - previous
	return &pb.ScoreChangeResponse{
		CurrentScore:  current,
		PreviousScore: previous,
		Change:        change,
	}, nil
}
