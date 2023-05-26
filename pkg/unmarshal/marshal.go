package unmarshal

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"go-web-app/models"
	"strings"
	"time"
)

func UnPackJob(value []byte) (ret *models.Job, err error) {

	var job *models.Job
	job = &models.Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, models.JobDir)
}
func ExtractKillerName(jobKey string) string {
	return strings.TrimPrefix(jobKey, models.JobKill)
}

type JobEvent struct {
	EventType int
	Job       *models.Job
}

func BUildJobEvent(evenType int, job *models.Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: evenType,
		Job:       job,
	}
}

func BuildJobSchedulePlan(job *models.Job) (jobSchedulePlan *models.JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &models.JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}
