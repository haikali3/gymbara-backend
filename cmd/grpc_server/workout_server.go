package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/haikali3/gymbara-backend/internal/database"
	pb "github.com/haikali3/gymbara-backend/pkg/proto"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// workoutServer implements pb.WorkoutServiceServer.
type workoutServer struct {
	pb.UnimplementedWorkoutServiceServer
}

// GetWorkoutHistory queries the workout history for a user between two dates.
func (s *workoutServer) GetWorkoutHistory(ctx context.Context, req *pb.WorkoutHistoryRequest) (*pb.WorkoutHistoryResponse, error) {
	utils.Logger.Info("Received request for workout history", zap.Int32("user_id", req.UserId))

	if req.UserId == 0 || req.StartDate == "" || req.EndDate == "" {
		utils.Logger.Error("Invalid request parameters", zap.Any("request", req))
		return nil, fmt.Errorf("missing required parameters")
	}

	// Log before executing the query
	utils.Logger.Info("Preparing to execute query", zap.Int32("userId", req.UserId))

	// Query user workout history
	query := `
    SELECT ued.exercise_id, e.name, ued.custom_reps, ued.custom_load, ued.submitted_at
    FROM UserExercisesDetails ued
    JOIN Exercises e ON ued.exercise_id = e.id
    JOIN UserWorkouts uw ON ued.user_workout_id = uw.id
    WHERE uw.user_id = $1 AND ued.submitted_at BETWEEN $2 AND $3
    ORDER BY ued.submitted_at ASC
  `
	utils.Logger.Info("Executing query", zap.String("query", query))

	rows, err := database.DB.Query(query, req.UserId, req.StartDate, req.EndDate)
	if err != nil {
		utils.Logger.Error("Database query error", zap.Error(err))
		return nil, fmt.Errorf("database query error: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.Logger.Fatal("Error closing rows", zap.Error(err))
		}
	}()

	var records []*pb.WorkoutRecord
	for rows.Next() {
		var record pb.WorkoutRecord
		var submittedAt sql.NullTime

		err := rows.Scan(&record.ExerciseId, &record.ExerciseName, &record.CustomReps, &record.CustomLoad, &submittedAt)
		if err != nil {
			utils.Logger.Error("Error scanning row", zap.Error(err))
			return nil, err
		}

		if submittedAt.Valid {
			record.SubmittedAt = timestamppb.New(submittedAt.Time)
		}
		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		utils.Logger.Error("Row iteration error", zap.Error(err))
		return nil, err
	}

	utils.Logger.Info("Successfully fetched workout history", zap.Int("record_count", len(records)))
	return &pb.WorkoutHistoryResponse{Records: records}, nil
}

func initDB() {
	connStr := "user=postgres dbname=postgres sslmode=disable" // Update as needed
	var err error
	database.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		utils.Logger.Fatal("Failed to open database", zap.Error(err))
	}
	if err = database.DB.Ping(); err != nil {
		utils.Logger.Fatal("Failed to ping database", zap.Error(err))
	}
	utils.Logger.Info("Database connection established")
}

func main() {
	// Initialize logger
	utils.InitializeLogger()
	defer utils.SyncLogger() // flush logger on exit

	//TODO: use the one from db.go?
	initDB()

	// Listen on port 50051
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		utils.Logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterWorkoutServiceServer(grpcServer, &workoutServer{})

	utils.Logger.Info("Workout gRPC server running on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		utils.Logger.Fatal("Failed to serve", zap.Error(err))
	}
}
