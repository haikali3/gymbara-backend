package main

import (
	"context"
	"time"

	pb "github.com/haikali3/gymbara-backend/pkg/proto"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize logger
	utils.InitializeLogger()
	defer func() {
		if err := utils.SyncLogger(); err != nil {
			utils.Logger.Error("Failed to sync logger", zap.Error(err))
		}
	}() // flush logger on exit

	// Dial the server on port 50051
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.Logger.Fatal("Failed to connect to server", zap.Error(err))
	}
	defer func() {
		if err := conn.Close(); err != nil {
			utils.Logger.Error("Failed to close connection", zap.Error(err))
		}
	}()

	client := pb.NewWorkoutServiceClient(conn)

	// Define the request parameters
	req := &pb.WorkoutHistoryRequest{
		UserId:    1,
		StartDate: "2025-01-01",
		EndDate:   "2025-12-01",
	}

	// Set a context timeout for the RPC
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call GetWorkoutHistory
	res, err := client.GetWorkoutHistory(ctx, req)
	if err != nil {
		utils.Logger.Error("Error fetching workout history", zap.Error(err))
		return
	}

	if res == nil {
		utils.Logger.Warn("Received nil response from server")
		return
	}

	// Log the received workout history records
	utils.Logger.Info("Workout History:")
	for _, record := range res.Records {
		utils.Logger.Info("Workout Record",
			zap.String("exercise_name", record.ExerciseName),
			zap.Int("reps", int(record.CustomReps)),
			zap.Int("load", int(record.CustomLoad)),
			zap.String("date", record.SubmittedAt.AsTime().Format("2006-01-02")),
		)
	}
}
