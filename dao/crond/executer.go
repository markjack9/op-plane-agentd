package crond

import (
	"go-web-app/models"
	"math/rand"
	"os/exec"
	"time"
)

type Executor struct {
}

var (
	Gexecutor *Executor
)

func (excutor *Executor) ExecuteJob(info *models.JobExecutingInfo) {
	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *models.JobExecuteResult
			jobLock *JobLock
		)
		result = &models.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}
		jobLock = InitJobLock(info.Job.Name, GJobmgr.Kv, GJobmgr.Lease)
		result.StartTime = time.Now()
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		err = jobLock.TryLock()
		defer jobLock.UnLock()
		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			result.StartTime = time.Now()
			cmd = exec.CommandContext(info.CancleCtx, "D:\\unix\\bin\\bash.exe", "-c", info.Job.Command)
			output, err = cmd.CombinedOutput()
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err

		}
		GScheduler.PushJobResult(result)
	}()
}
func InitExecutor() (err error) {
	Gexecutor = &Executor{}
	return
}

func (scheduler *Scheduler) PushJobResult(jobResult *models.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
