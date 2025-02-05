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

	// conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.Logger.Fatal("Failed to start server", zap.Error(err))
	}
	defer conn.Close()

	client := pb.NewWorkoutServiceClient(conn)

	// define req param
	req := &pb.WorkoutHistoryRequest{
		UserId:    1,
		StartDate: "2024-01-01",
		EndDate:   "2024-02-01",
	}

	// fetch workout history with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.GetWorkoutHistory(ctx, req)
	if err != nil {
		utils.Logger.Fatal("Error fetching workout history", zap.Error(err))
	}

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
