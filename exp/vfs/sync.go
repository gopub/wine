package vfs

import (
	"fmt"
	"os"

	"github.com/gopub/errors"
)

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

func (fs *FileSystem) Sync(source *FileSystem) (log <-chan *SyncLog, done <-chan error) {
	logC := make(chan *SyncLog, 256)
	doneC := make(chan error, 2)
	go fs.startSync(source, logC, doneC)
	return logC, doneC
}

func (fs *FileSystem) startSync(source *FileSystem, logC chan<- *SyncLog, doneC chan<- error) {
	if err := fs.syncKeyChain(source); err != nil {
		doneC <- err
		return
	}

	root := (*FileSystem)(fs).Root()
	for _, sub := range source.root.Files {
		logger.Debugf("Start %s", sub.Path())
		if err := fs.syncFile(source, sub, root, logC); err != nil {
			logger.Errorf("Cannot sync %s: %v", sub.Name(), err)
		} else {
			logger.Debugf("Synced %s", sub.Path())
		}
	}
	logger.Debug("Completed")
	doneC <- nil
}

func (fs *FileSystem) syncKeyChain(source *FileSystem) error {
	src, err := source.KeyChain().load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("load source: %w", err)
	}

	kc := fs.KeyChain()
	dst, err := kc.load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("load destin: %w", err)
		}
	}

	for k, v := range src {
		if cur, ok := dst[k]; ok && cur.ModifiedAt > v.ModifiedAt {
			continue
		}
		dst[k] = v
	}

	err = kc.save(dst)
	return errors.Wrapf(err, "save")
}

func (fs *FileSystem) syncFile(source *FileSystem, f *FileInfo, toDir *FileInfo, logC chan<- *SyncLog) error {
	log := &SyncLog{
		Source: f,
	}
	if f.IsDir() {
		fi, err := fs.Wrapper().Stat(f.UUID())
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("stat %s: %w", f.UUID(), err)
			}

			fi, err = fs.Wrapper().Mkdir(toDir.UUID(), f.Name())
			if err != nil {
				return fmt.Errorf("mkdir %s: %w", f.Name(), err)
			}
			fi.SetUUID(f.UUID())
			fi.SetPermission(f.Permission)
			fi.SetCreatedAt(f.CreatedAt())
			log.Action = SyncActionAdded
		} else {
			log.Action = SyncActionOverrided
		}
		for _, sub := range f.Files {
			if err = fs.syncFile(source, sub, fi, logC); err != nil {
				fileLog := &SyncLog{
					Source: sub,
					Action: SyncActionFailed,
				}
				logger.Errorf("Cannot sync file %s: %w", sub.Name(), err)
				logC <- fileLog
			}
		}
		logC <- log
		return nil
	}

	fi, err := fs.Wrapper().Stat(f.UUID())
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", f.UUID(), err)
		}
		log.Action = SyncActionAdded
	} else {
		if fi.ModifiedAt() >= f.ModifiedAt() {
			log.Action = SyncActionSkipped
			return nil
		}
		log.Action = SyncActionOverrided
	}

	df, err := fs.Wrapper().Create(toDir.UUID(), f.Name())
	if err != nil {
		return fmt.Errorf("mkdir %s: %w", f.Path(), err)
	}
	defer df.Close()

	fi = df.info
	fi.SetUUID(f.UUID())
	fi.SetPermission(f.Permission)
	fi.SetCreatedAt(f.CreatedAt())
	fi.SetMIMEType(f.MIMEType())
	fi.SetDuration(f.Duration())
	fi.SetLocation(f.Location())

	logger.Debug("sync file", fi.Name())

	for _, page := range f.Pages {
		data, err := source.storage.Get(page)
		if err != nil {
			return fmt.Errorf("get %s page %s: %w", fi.Path(), page, err)
		}
		err = source.DecryptPage(data)
		if err != nil {
			return fmt.Errorf("decrypt page %s: %w", page, err)
		}
		_, err = df.Write(data)
		if err != nil {
			return fmt.Errorf("write page %s: %w", page, err)
		}
	}

	if f.Thumbnail != "" {
		data, err := source.Wrapper().ReadThumbnail(f.UUID())
		if err != nil {
			return fmt.Errorf("read thumbnail %s: %w", f.Thumbnail, err)
		}
		err = df.WriteThumbnail(data)
		if err != nil {
			return fmt.Errorf("write thumbnail %s: %w", f.Thumbnail, err)
		}
		fi.Thumbnail = f.Thumbnail
	}

	logC <- log
	return nil
}
