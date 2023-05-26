package crond

import (
	"context"
	"go-web-app/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	leaseId    clientv3.LeaseID
	isLocked   bool
	Jobname    string
	cancelFunc context.CancelFunc
}

func InitJobLock(jobname string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		Jobname: jobname,
	}
	return
}

func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrepResp *clientv3.LeaseGrantResponse
		cancelCtx     context.Context
		cancleFunc    context.CancelFunc
		leaseId       clientv3.LeaseID
		keepRespChan  <-chan *clientv3.LeaseKeepAliveResponse
		txn           clientv3.Txn
		lockKey       string
		txnResp       *clientv3.TxnResponse
	)
	if leaseGrepResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	cancelCtx, cancleFunc = context.WithCancel(context.TODO())
	leaseId = leaseGrepResp.ID
	if keepRespChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseId); err != nil {
		goto FAIL
	}

	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()
	txn = jobLock.kv.Txn(context.TODO())
	lockKey = models.JobLock + jobLock.Jobname
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	if !txnResp.Succeeded {
		err = models.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}
	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancleFunc
	jobLock.isLocked = true
	return
FAIL:
	cancleFunc()
	jobLock.lease.Revoke(context.TODO(), leaseId)

	return

}

func (jobLock *JobLock) UnLock() {
	if jobLock.isLocked {
		jobLock.cancelFunc()
		jobLock.lease.Revoke(context.TODO(), jobLock.leaseId)
	}

	return
}
