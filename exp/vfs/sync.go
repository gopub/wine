package vfs

type fileSystemSync FileSystem

type SyncAction string

const (
	SyncActionAdded     SyncAction = "Added"
	SyncActionOverrided SyncAction = "Overrided"
	SyncActionSkipped   SyncAction = "Skipped"
	SyncActionFailed    SyncAction = "Failed"
)

type SyncLog struct {
	Source      *FileInfo
	Destination *FileInfo
	Action      SyncAction
}

func (s *fileSystemSync) Sync(fs *FileSystem) (log <-chan *SyncLog, done <-chan error) {
	logC := make(chan *SyncLog, 256)
	doneC := make(chan error, 1)
	go s.start(fs, logC, doneC)
	return logC, doneC
}

func (s *fileSystemSync) start(fs *FileSystem, log chan<- *SyncLog, done chan<- error) {
	// TODO:

}
