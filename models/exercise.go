package models

type Exercise struct {
	ID           int    `json:"id"`
	ExerciseName string `json:"name"`
	WarmupSets   int    `json:"warmup_sets"`
	WorkSets     int    `json:"work_sets"`
	Reps         string `json:"reps"`
	Load         int    `json:"load"`
	RPE          string `json:"rpe"`
	RestTime     string `json:"rest_time"`
}
