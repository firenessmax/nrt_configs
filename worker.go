package nrt_configs

import (
	"cloud.google.com/go/firestore"
	"context"
	"log"
)

type DocsHandler func([]*firestore.DocumentSnapshot) error

type ErrorHandler func(err error)

type RTWorker struct {
	qsi      *firestore.QuerySnapshotIterator
	ctx      context.Context
	onChange DocsHandler
	onError  ErrorHandler
	logMode  bool
}

func NewRTWorker(ctx context.Context, qsi *firestore.QuerySnapshotIterator) *RTWorker {
	return &RTWorker{
		qsi: qsi,
		ctx: ctx,
	}
}

func (w *RTWorker) OnChange(handler DocsHandler) {
	w.onChange = handler
}

func (w *RTWorker) OnError(handler ErrorHandler) {
	w.onError = handler
}
func (w *RTWorker) LogMode(enabled bool) {
	w.logMode = enabled
}

func (w *RTWorker) Listen() {
	for {
		select {
		case <-w.ctx.Done():
			w.qsi.Stop()
			return
		default:
			snap, err := w.qsi.Next()
			if w.handleError(err) {
				continue
			}
			docs, err := snap.Documents.GetAll()
			if w.handleError(err) {
				continue
			}
			err = w.onChange(docs)
			if w.handleError(err) {
				continue
			}
		}
	}
}

func (w *RTWorker) handleError(err error) bool {
	if err == nil {
		return false
	}
	if w.logMode {
		log.Printf("[RTConfig:loop] %s", err)
	}
	w.onError(err)
	return true
}
