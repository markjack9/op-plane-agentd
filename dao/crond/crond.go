package crond

import (
	"context"
	"fmt"
	"go-web-app/dao"
	"go-web-app/models"
	"go-web-app/pkg/codeconversion"
	"go-web-app/pkg/unmarshal"
	"go-web-app/settings"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strconv"
	"time"
)

var (
	clinet     *clientv3.Client
	kv         clientv3.KV
	lease      clientv3.Lease
	GJobmgr    *models.JobMgr
	oldjob     *models.Job
	watcher    clientv3.Watcher
	GScheduler *Scheduler
)

type JobMgr struct {
	Kv     clientv3.KV
	Lease  clientv3.Lease
	Clinet *clientv3.Client
}

type Scheduler struct {
	jobEventChan     chan *unmarshal.JobEvent
	jobPlanTable     map[string]*models.JobSchedulePlan
	jobExcutingTable map[string]*models.JobExecutingInfo
	jobResultChan    chan *models.JobExecuteResult
}

func InitCrontab(cfg *settings.EtcdConfig) (err error) {
	config := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Millisecond,
		Username:    cfg.Username,
		Password:    cfg.Password,
	}
	if clinet, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	//获取KV和Lease的API子集
	kv = clientv3.NewKV(clinet)
	lease = clientv3.NewLease(clinet)
	watcher = clientv3.Watcher(clinet)
	//赋值单例
	GJobmgr = &models.JobMgr{
		Clinet:  clinet,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	go func() {
		err := InitScheduler()
		if err != nil {
			return
		}
	}()
	go func() {
		err := InitExecutor()
		if err != nil {
			return
		}
	}()
	if err = WatchJobs(GJobmgr); err != nil {
		return
	}
	if err = WatchKiller(GJobmgr); err != nil {
		return
	}

	return
}
func InitScheduler() (err error) {
	GScheduler = &Scheduler{
		jobEventChan:     make(chan *unmarshal.JobEvent, 1000),
		jobPlanTable:     make(map[string]*models.JobSchedulePlan),
		jobExcutingTable: make(map[string]*models.JobExecutingInfo),
		jobResultChan:    make(chan *models.JobExecuteResult, 1000),
	}
	go GScheduler.ScheduleLoop()
	return
}
func WatchKiller(jobmgr *models.JobMgr) (err error) {
	var (
		job        *models.Job
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName    string
		jobEnvent  *unmarshal.JobEvent
	)
	go func() {
		watchChan = jobmgr.Watcher.Watch(context.TODO(), models.JobKill, clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					jobName = unmarshal.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &models.Job{Name: jobName}
					jobEnvent = unmarshal.BUildJobEvent(models.JobKiller, job)
					//TODO:推一个删除时间给scheduler
					GScheduler.PushJobEvent(jobEnvent)
				case mvccpb.DELETE:

				}

			}
		}
	}()
	return
}
func WatchJobs(jobmgr *models.JobMgr) (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvpair             *mvccpb.KeyValue
		job                *models.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
		jobEnvent          *unmarshal.JobEvent
	)
	GScheduler = new(Scheduler)
	if getResp, err = jobmgr.Kv.Get(context.TODO(), models.JobDir, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, kvpair = range getResp.Kvs {
		if job, err = unmarshal.UnPackJob(kvpair.Value); err == nil {
			fmt.Println("同步job给scheduler调度协程", job)
			jobEnvent = unmarshal.BUildJobEvent(models.JobEventSave, job)
			GScheduler.PushJobEvent(jobEnvent)
		}

	}

	go func() {
		watchStartRevision = getResp.Header.Revision + 1
		watchChan = jobmgr.Watcher.Watch(context.TODO(), models.JobDir, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					if job, err = unmarshal.UnPackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					jobEnvent = unmarshal.BUildJobEvent(models.JobEventSave, job)
					fmt.Println("数据更新", *jobEnvent.Job)
				//TODO:反序列化job，推送给scheduler
				case mvccpb.DELETE:
					jobName = unmarshal.ExtractJobName(string(watchEvent.Kv.Key))
					job = &models.Job{Name: jobName}
					jobEnvent = unmarshal.BUildJobEvent(models.JobEventDelete, job)
					fmt.Println("数据删除", *jobEnvent.Job)

				}
				//TODO:推一个删除时间给scheduler
				GScheduler.PushJobEvent(jobEnvent)
			}
		}
	}()
	return
}
func (scheduler *Scheduler) PushJobEvent(jobevent *unmarshal.JobEvent) {
	scheduler.jobEventChan <- jobevent
}
func (scheduler *Scheduler) ScheduleLoop() {
	var (
		jobEvent      *unmarshal.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult     *models.JobExecuteResult
	)
	scheduleAfter = scheduler.TrySchedule()

	scheduleTimer = time.NewTimer(scheduleAfter)
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C:
		case jobResult = <-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}
		scheduleAfter = scheduler.TrySchedule()
		scheduleTimer.Reset(scheduleAfter)
	}
}
func (scheduler *Scheduler) handleJobEvent(jobEvent *unmarshal.JobEvent) {
	var (
		jobSchedulePlan *models.JobSchedulePlan
		jobExisted      bool
		err             error
		jobExecuting    bool
		jobExecuteInfo  *models.JobExecutingInfo
	)
	switch jobEvent.EventType {
	case models.JobEventSave:
		if jobSchedulePlan, err = unmarshal.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case models.JobEventDelete:
		fmt.Println("数据", jobEvent.Job)
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case models.JobKiller:
		if jobExecuteInfo, jobExecuting = scheduler.jobExcutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancleFunc()
		}
	}
}
func (scheduler *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	var (
		JobPlan  *models.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
		return
	}
	now = time.Now()
	for _, JobPlan = range scheduler.jobPlanTable {
		if JobPlan.NextTime.Before(now) || JobPlan.NextTime.Equal(now) {
			GScheduler.TryStartJob(JobPlan)
			JobPlan.NextTime = JobPlan.Expr.Next(now)
		}
		if nearTime == nil || JobPlan.NextTime.Before(*nearTime) {
			nearTime = &JobPlan.NextTime
		}
	}
	schedulerAfter = (*nearTime).Sub(now)
	return
}

func (scheduler *Scheduler) TryStartJob(jobPlan *models.JobSchedulePlan) {
	var (
		jobExecuteInfo *models.JobExecutingInfo
		jobExecuting   bool
	)
	if jobExecuteInfo, jobExecuting = scheduler.jobExcutingTable[jobPlan.Job.Name]; jobExecuting {
		fmt.Println("任务在进行中")
		return
	}
	jobExecuteInfo = BuildJobExecuteInfo(jobPlan)
	scheduler.jobExcutingTable[jobPlan.Job.Name] = jobExecuteInfo
	fmt.Println("执行任务", jobExecuteInfo)
	Gexecutor.ExecuteJob(jobExecuteInfo)

}

func BuildJobExecuteInfo(jobSchedulePlan *models.JobSchedulePlan) (jobExecuteInfo *models.JobExecutingInfo) {
	jobExecuteInfo = &models.JobExecutingInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.CancleCtx, jobExecuteInfo.CancleFunc = context.WithCancel(context.TODO())
	return
}
func (scheduler *Scheduler) handleJobResult(jobResult *models.JobExecuteResult) {
	delete(scheduler.jobExcutingTable, jobResult.ExecuteInfo.Job.Name)
	fmt.Println("清除已执行的任务", jobResult.ExecuteInfo.Job.Name)

	if jobResult.Err != models.ERR_LOCK_ALREADY_REQUIRED {
		Joblog := &models.CrontabJob{
			JobId:        0,
			JobCronExpr:  jobResult.ExecuteInfo.Job.CronExpr,
			JobName:      jobResult.ExecuteInfo.Job.Name,
			JobShell:     jobResult.ExecuteInfo.Job.Command,
			JobStatus:    0,
			JobStartTime: strconv.FormatInt(jobResult.StartTime.UnixNano()/1000/1000, 10),
			JobStopTime:  strconv.FormatInt(jobResult.EndTime.UnixNano()/1000/1000, 10),
			JobInfo:      codeconversion.ConvertByte2String(jobResult.Output, "GB18030"),
			JobRunning:   strconv.FormatInt(jobResult.ExecuteInfo.PlanTime.UnixNano()/1000/1000, 10),
		}

		if jobResult.Err != nil {
			Joblog.JobErr = jobResult.Err.Error()
		} else {
			Joblog.JobErr = ""
		}

		go func() {
			err := PushLogToServer(Joblog)
			if err != nil {
				return
			}
		}()

	}
}

func (jobmgr *JobMgr) CreateJobLock(jobname string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobname, jobmgr.Kv, jobmgr.Lease)
	return
}
func PushLogToServer(Joblog *models.CrontabJob) (err error) {
	parame := models.ParameCrontab{
		ParameOption: "taskjoblog",
		CrontabJob: models.CrontabJob{
			JobId:        Joblog.JobId,
			JobCronExpr:  Joblog.JobCronExpr,
			JobName:      Joblog.JobName,
			JobShell:     Joblog.JobShell,
			JobStatus:    Joblog.JobStatus,
			JobStartTime: Joblog.JobStartTime,
			JobStopTime:  Joblog.JobStopTime,
			JobInfo:      Joblog.JobInfo,
			JobRunning:   Joblog.JobRunning,
			JobErr:       Joblog.JobErr,
		},
	}
	err = dao.LogPost(parame, settings.Conf.ServerConfig, "crontab")
	if err != nil {
		return
	}
	return
}
