// main.go
package main

import (
	"context"
	"database/sql"
	"log"
	"net"

	pb "github.com/aa-ahmed-aa/zendesk_grpc_service/proto"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedRatingServiceServer
	db *sql.DB
}

func (s *server) OverallScore(ctx context.Context, req *pb.DateRangeRequest) (*pb.ScoreResponse, error) {
	query := `
		SELECT ROUND(
			100.0 * SUM(r.rating * rc.weight) / NULLIF(SUM(5.0 * rc.weight), 0),
			2
		) AS score
		FROM ratings r
		JOIN rating_categories rc ON r.rating_category_id = rc.id
		WHERE r.created_at BETWEEN ? AND ?;
	`
	var score float64
	err := s.db.QueryRow(query, req.StartDate, req.EndDate).Scan(&score)
	if err != nil {
		return nil, err
	}
	return &pb.ScoreResponse{Score: score}, nil
}

func (s *server) ScoreChange(ctx context.Context, req *pb.ScoreChangeRequest) (*pb.ScoreChangeResponse, error) {
	query := `
		SELECT ROUND(
			100.0 * SUM(r.rating * rc.weight) / NULLIF(SUM(5.0 * rc.weight), 0),
			2
		) AS score
		FROM ratings r
		JOIN rating_categories rc ON r.rating_category_id = rc.id
		WHERE r.created_at BETWEEN ? AND ?;
	`

	var current, previous float64
	if err := s.db.QueryRow(query, req.CurrentStart, req.CurrentEnd).Scan(&current); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow(query, req.PreviousStart, req.PreviousEnd).Scan(&previous); err != nil {
		return nil, err
	}

	change := current - previous
	return &pb.ScoreChangeResponse{
		CurrentScore:  current,
		PreviousScore: previous,
		Change:        change,
	}, nil
}

func (s *server) CategoryScores(ctx context.Context, req *pb.DateRangeRequest) (*pb.CategoryScoresResponse, error) {
	query := `
		SELECT
			rc.name AS category,
			strftime('%Y-%m-%d', r.created_at) AS date,
			ROUND(100.0 * SUM(r.rating * rc.weight) / NULLIF(SUM(5.0 * rc.weight), 0), 2) AS score,
			COUNT(*) as ratings_count
		FROM ratings r
		JOIN rating_categories rc ON r.rating_category_id = rc.id
		WHERE r.created_at BETWEEN ? AND ?
		GROUP BY category, date
		ORDER BY category, date;
	`

	rows, err := s.db.Query(query, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type categoryData struct {
		scores       map[string]float64
		ratingsCount int32
		totalScore   float64
		daysCount    int32
	}

	categoryMap := make(map[string]*categoryData)

	for rows.Next() {
		var category, date string
		var score float64
		var count int32
		if err := rows.Scan(&category, &date, &score, &count); err != nil {
			return nil, err
		}
		if _, exists := categoryMap[category]; !exists {
			categoryMap[category] = &categoryData{scores: make(map[string]float64)}
		}
		categoryMap[category].scores[date] = score
		categoryMap[category].ratingsCount += count
		categoryMap[category].totalScore += score
		categoryMap[category].daysCount++
	}

	var response pb.CategoryScoresResponse
	for cat, data := range categoryMap {
		avgScore := 0.0
		if data.daysCount > 0 {
			avgScore = data.totalScore / float64(data.daysCount)
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

func (s *server) TicketScores(ctx context.Context, req *pb.DateRangeRequest) (*pb.TicketScoresResponse, error) {
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

	rows, err := s.db.Query(query, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ticketMap := make(map[string]map[string]float64)

	for rows.Next() {
		var ticketID, category string
		var score float64
		if err := rows.Scan(&ticketID, &category, &score); err != nil {
			return nil, err
		}
		if _, exists := ticketMap[ticketID]; !exists {
			ticketMap[ticketID] = make(map[string]float64)
		}
		ticketMap[ticketID][category] = score
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

func main() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRatingServiceServer(s, &server{db: db})
	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
