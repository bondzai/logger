package model

import pb "github.com/bondzai/logger/proto"

type Task struct {
	ID           int         `bson:"task_id" json:"task_id"`
	Organization string      `bson:"organization" json:"organization"`
	ProjectID    int         `bson:"project_id" json:"project_id"`
	Type         pb.TaskType `bson:"type" json:"type"`
	Name         string      `bson:"task_name" json:"task_name"`
	Interval     int64       `bson:"interval" json:"interval"`
	CronExpr     []string    `bson:"task_cron_expression" json:"task_cron_expression"`
	Disabled     bool        `bson:"disabled" json:"disabled"`
	TimeStamp    string      `bson:"timestamp" json:"timestamp"`
}
