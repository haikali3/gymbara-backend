package main

import (
	"net"

	pb "github.com/haikali3/gymbara-backend/pkg/proto"
	"github.com/haikali3/gymbara-backend/pkg/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type workoutServer struct {
	pb.UnimplementedWorkoutServiceServer
}

func main() {
	// init logger
	utils.InitializeLogger()
	defer utils.SyncLogger() //logger flush on exit

	// conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	utils.Logger.Fatal("Failed to start server", zap.Error(err))
	// }
	// defer conn.Close()

	// client := pb.NewWorkoutServiceClient(conn)

	// // define req param
	// req := &pb.WorkoutHistoryRequest{
	// 	UserId:    1,
	// 	StartDate: "2025-01-01",
	// 	EndDate:   "2025-12-01",
	// }

	// // fetch workout history with context timeout
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// res, err := client.GetWorkoutHistory(ctx, req)
	// if err != nil {
	// 	utils.Logger.Error("Error fetching workout history", zap.Error(err))
	// }

	// if res == nil {
	// 	utils.Logger.Warn("Received nil response from server")
	// 	return
	// }

	// utils.Logger.Info("Workout History:")
	// for _, record := range res.Records {
	// 	utils.Logger.Info("Workout Record",
	// 		zap.String("exercise_name", record.ExerciseName),
	// 		zap.Int("reps", int(record.CustomReps)),
	// 		zap.Int("load", int(record.CustomLoad)),
	// 		zap.String("date", record.SubmittedAt.AsTime().Format("2006-01-02")),
	// 	)
	// }

	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		utils.Logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWorkoutServiceServer(grpcServer, &workoutServer{})

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	utils.Logger.Info("Workout gRPC server running on :50051")

	if err := grpcServer.Serve(listener); err != nil {
		utils.Logger.Fatal("Failed to serve", zap.Error(err))
	}

}
