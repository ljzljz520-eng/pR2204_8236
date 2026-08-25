package store

import (
	"errors"
	"os"
	"sort"
	"strings"

	"examvault/internal/domain"
	bolt "go.etcd.io/bbolt"
)

var (
	recordsBucket     = []byte("records")
	eventsBucket      = []byte("audit_events")
	workflowsBucket   = []byte("workflows")
	attachmentsBucket = []byte("attachments")
	metaBucket        = []byte("meta")
)

type Store struct {
	db   *bolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage path is empty")
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{NoSync: true})
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, path: path}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range [][]byte{recordsBucket, eventsBucket, workflowsBucket, attachmentsBucket, metaBucket} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		if tx.Bucket(metaBucket).Get([]byte("sequence")) == nil {
			return tx.Bucket(metaBucket).Put([]byte("sequence"), []byte("0"))
		}
		return nil
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Path() string { return s.path }

func (s *Store) PutRecord(r domain.Record) error {
	if err := r.Validate(); err != nil {
		return err
	}
	data, err := r.Marshal()
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(recordsBucket).Put([]byte(r.ID), data) })
}

func (s *Store) GetRecord(id string) (domain.Record, error) {
	var r domain.Record
	err := s.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(recordsBucket).Get([]byte(id))
		if value == nil {
			return domain.ErrNotFound
		}
		var err error
		r, err = domain.DecodeRecord(append([]byte(nil), value...))
		return err
	})
	return r, err
}

func (s *Store) ListRecords() ([]domain.Record, error) {
	records := make([]domain.Record, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(recordsBucket).ForEach(func(_, value []byte) error {
			r, err := domain.DecodeRecord(append([]byte(nil), value...))
			if err != nil {
				return err
			}
			records = append(records, r)
			return nil
		})
	})
	domain.SortRecords(records)
	return records, err
}

func (s *Store) PutWorkflow(w domain.Workflow) error {
	data, err := w.Marshal()
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(workflowsBucket).Put([]byte(w.ID), data) })
}

func (s *Store) GetWorkflow(id string) (domain.Workflow, error) {
	var w domain.Workflow
	err := s.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(workflowsBucket).Get([]byte(id))
		if value == nil {
			return domain.ErrNotFound
		}
		var err error
		w, err = domain.DecodeWorkflow(append([]byte(nil), value...))
		return err
	})
	return w, err
}

func (s *Store) PutAttachment(a domain.Attachment) error {
	if err := a.Validate(); err != nil {
		return err
	}
	data, err := a.Marshal()
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(attachmentsBucket).Put([]byte(a.ID), data) })
}

func (s *Store) GetAttachment(id string) (domain.Attachment, error) {
	var a domain.Attachment
	err := s.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket(attachmentsBucket).Get([]byte(id))
		if value == nil {
			return domain.ErrNotFound
		}
		var err error
		a, err = domain.DecodeAttachment(append([]byte(nil), value...))
		return err
	})
	return a, err
}

func (s *Store) PutEvent(e domain.AuditEvent) error {
	data, err := e.Marshal()
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error { return tx.Bucket(eventsBucket).Put(eventKey(e.Sequence, e.ID), data) })
}

func (s *Store) NextSequence() (int, error) {
	next := 0
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBucket)
		raw := b.Get([]byte("sequence"))
		if len(raw) > 0 {
			for _, c := range raw {
				if c >= '0' && c <= '9' {
					next = next*10 + int(c-'0')
				}
			}
		}
		next++
		return b.Put([]byte("sequence"), []byte(itoa(next)))
	})
	return next, err
}

func (s *Store) ListEvents(recordID string) ([]domain.AuditEvent, error) {
	result := make([]domain.AuditEvent, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(eventsBucket).ForEach(func(_, value []byte) error {
			e, err := domain.DecodeAuditEvent(append([]byte(nil), value...))
			if err != nil {
				return err
			}
			if recordID == "" || e.RecordID == recordID {
				result = append(result, e)
			}
			return nil
		})
	})
	domain.SortEvents(result)
	return result, err
}

func (s *Store) Delete(path string) error {
	if path == "" {
		path = s.path
	}
	if err := s.Close(); err != nil {
		return err
	}
	return os.Remove(path)
}

func eventKey(sequence int, id string) []byte { return []byte(itoa(sequence) + ":" + id) }

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	neg := value < 0
	if neg {
		value = -value
	}
	out := make([]byte, 0, 12)
	for value > 0 {
		out = append(out, byte('0'+value%10))
		value /= 10
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if neg {
		out = append([]byte{'-'}, out...)
	}
	return string(out)
}

func mergeRecords(left, right []domain.Record) []domain.Record {
	merged := append(append([]domain.Record(nil), left...), right...)
	sort.Slice(merged, func(i, j int) bool { return merged[i].ID < merged[j].ID })
	return merged
}

func (s *Store) PutRecords(records []domain.Record) error {
	if err := domain.ValidateCollection(records); err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recordsBucket)
		for _, record := range records {
			data, err := record.Marshal()
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(record.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) PutAttachments(attachments []domain.Attachment) error {
	for _, attachment := range attachments {
		if err := attachment.Validate(); err != nil {
			return err
		}
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(attachmentsBucket)
		for _, attachment := range attachments {
			data, err := attachment.Marshal()
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(attachment.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) ListAttachments(recordID string) ([]domain.Attachment, error) {
	attachments := make([]domain.Attachment, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(attachmentsBucket).ForEach(func(_, value []byte) error {
			attachment, err := domain.DecodeAttachment(append([]byte(nil), value...))
			if err != nil {
				return err
			}
			if recordID == "" || attachment.RecordID == recordID {
				attachments = append(attachments, attachment)
			}
			return nil
		})
	})
	sort.Slice(attachments, func(i, j int) bool { return attachments[i].ID < attachments[j].ID })
	return attachments, err
}

func (s *Store) ListWorkflows(recordID string) ([]domain.Workflow, error) {
	workflows := make([]domain.Workflow, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(workflowsBucket).ForEach(func(_, value []byte) error {
			workflow, err := domain.DecodeWorkflow(append([]byte(nil), value...))
			if err != nil {
				return err
			}
			if recordID == "" || workflow.RecordID == recordID {
				workflows = append(workflows, workflow)
			}
			return nil
		})
	})
	sort.Slice(workflows, func(i, j int) bool { return workflows[i].ID < workflows[j].ID })
	return workflows, err
}

func (s *Store) DeleteRecord(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("record id is empty")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(recordsBucket).Delete([]byte(id)); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) RecordRevision(id string) (int, error) {
	record, err := s.GetRecord(id)
	if err != nil {
		return 0, err
	}
	return record.Revision, nil
}

func (s *Store) ReplaceRecord(id string, expectedRevision int, update func(*domain.Record) error) (domain.Record, error) {
	var result domain.Record
	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(recordsBucket)
		raw := bucket.Get([]byte(id))
		if raw == nil {
			return domain.ErrNotFound
		}
		record, err := domain.DecodeRecord(append([]byte(nil), raw...))
		if err != nil {
			return err
		}
		if expectedRevision > 0 && record.Revision != expectedRevision {
			return domain.ErrConflict
		}
		if err := update(&record); err != nil {
			return err
		}
		if err := domain.ValidateRecordForWrite(record); err != nil {
			return err
		}
		encoded, err := record.Marshal()
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(id), encoded); err != nil {
			return err
		}
		result = record
		return nil
	})
	return result, err
}

type Stats struct {
	Records     int
	Attachments int
	Workflows   int
	Events      int
}

func (s *Store) Stats() (Stats, error) {
	stats := Stats{}
	err := s.db.View(func(tx *bolt.Tx) error {
		stats.Records = tx.Bucket(recordsBucket).Stats().KeyN
		stats.Attachments = tx.Bucket(attachmentsBucket).Stats().KeyN
		stats.Workflows = tx.Bucket(workflowsBucket).Stats().KeyN
		stats.Events = tx.Bucket(eventsBucket).Stats().KeyN
		return nil
	})
	return stats, err
}

func (s *Store) Snapshot() (map[string][]byte, error) {
	snapshot := make(map[string][]byte)
	err := s.db.View(func(tx *bolt.Tx) error {
		for name, bucketName := range map[string][]byte{"records": recordsBucket, "attachments": attachmentsBucket, "workflows": workflowsBucket} {
			bucket := tx.Bucket(bucketName)
			if err := bucket.ForEach(func(key, value []byte) error {
				snapshot[name+":"+string(key)] = append([]byte(nil), value...)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	return snapshot, err
}

func (s *Store) Health() error {
	if s == nil || s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(recordsBucket) == nil || tx.Bucket(metaBucket) == nil {
			return errors.New("required buckets missing")
		}
		return nil
	})
}

func (s *Store) WithRecords(fn func([]domain.Record) error) error {
	records, err := s.ListRecords()
	if err != nil {
		return err
	}
	return fn(records)
}
