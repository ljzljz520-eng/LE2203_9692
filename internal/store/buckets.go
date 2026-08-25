package store

import "go.etcd.io/bbolt"

var (
	bucketRecords     = []byte("records")
	bucketAudit       = []byte("audit_events")
	bucketWorkflows   = []byte("workflows")
	bucketAttachments = []byte("attachments")
)

func initializeBuckets(tx *bbolt.Tx) error {
	for _, name := range [][]byte{bucketRecords, bucketAudit, bucketWorkflows, bucketAttachments} {
		if _, err := tx.CreateBucketIfNotExists(name); err != nil {
			return err
		}
	}
	return nil
}
