package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/haikali3/gymbara-backend/internal/database"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	pb "github.com/haikali3/gymbara-backend/proto/workout"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type workoutServer struct {
	pb.UnimplementedWorkoutServiceServer
}

func (s *workoutServer) GetWorkoutHistory(ctx context.Context, req *pb.WorkoutHistoryRequest) (*pb.WorkoutHistoryResponse, error) {
	// validate req
	if req.UserId == 0 || req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("missing required parameters")
	}

	// query user workout history
	query := `
    SELECT ued.exercise_id, e.name, ued.custom_reps, eud.custom_load, ued.submitted_at
    FROM UserExercisesDetails ued
    JOIN Exercises e ON ued.exercise_id = e.id
    JOIN UserWorkouts uw ON ued.user_workout_id = uw.id
    WHERE uw.user_id = $1 AND ued.submitted_at BETWEEN $2 AND $3
    ORDER BY ued.submitted_at ASC
  `

	rows, err := database.DB.Query(query, req.UserId, req.StartDate, req.EndDate)
	if err != nil {
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

	return &pb.WorkoutHistoryResponse{Records: records}, nil
}

func main() {
	//start grpc server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		utils.Logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWorkoutServiceServer(grpcServer, &workoutServer{})

	utils.Logger.Info("Workout gRPC server running on :50051")

	// Graceful shutdown handling
	// go func() {
	// 	if err := grpcServer.Serve(listener); err != nil {
	// 		utils.Logger.Fatal("Failed to serve", zap.Error(err))
	// 	}
	// }()

	//Wait for termination signal (Ctrl+C)
	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// <-sigChan
	// utils.Logger.Info("Shutting down gRPC server gracefully...")
	// grpcServer.GracefulStop()

	if err := grpcServer.Serve(listener); err != nil {
		utils.Logger.Fatal("Failed to server", zap.Error(err))
	}
}
