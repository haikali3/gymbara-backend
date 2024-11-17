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

type WorkoutSection struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Route string `json:"route"`
}

type ExerciseDetails struct {
	Name       string `json:"name"`
	WarmupSets int    `json:"warmup_sets"`
	WorkSets   int    `json:"working_sets"`
	Reps       string `json:"reps"`
	Load       int    `json:"load"`
	RPE        string `json:"rpe"`
	RestTime   string `json:"rest_time"`
}
