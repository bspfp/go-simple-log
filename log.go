package gosimplelog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	LogDirPerm  = 0755
	LogFilePerm = 0644
)

var onceInit sync.Once

func InitLogFile(logDir, logFilename string, maxFiles int) (retval io.WriteCloser, reterr error) {
	onceInit.Do(func() {
		l := &logger{
			dir:      logDir,
			filename: logFilename,
			maxFiles: maxFiles,
		}

		if err := l.makeDir(); err != nil {
			reterr = err
			return
		}

		if l.needRotateOnStartup() {
			if err := l.rotateFile(); err != nil {
				reterr = err
				return
			}
		}

		if err := l.openFile(); err != nil {
			reterr = err
			return
		}

		log.SetOutput(io.MultiWriter(os.Stdout, l))
		retval = l
	})
	return
}

type logger struct {
	mu       sync.Mutex
	dir      string
	filename string
	maxFiles int
	openTime time.Time
	file     *os.File
}

func (l *logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

func (l *logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.maxFiles > 1 {
		now := time.Now().Unix() / 60 / 60 / 24
		open := l.openTime.Unix() / 60 / 60 / 24
		if now != open {
			if err := l.close(); err != nil {
				log.Println("close error:", err)
			}
			if err := l.rotateFile(); err != nil {
				log.Println("rotate error:", err)
			}
			if err := l.openFile(); err != nil {
				log.Println("open error:", err)
			}
		}
	}

	if l.file == nil {
		return len(p), nil // 쓴 척!
	}

	return l.file.Write(p)
}

func (l *logger) close() error {
	if l.file == nil {
		return nil
	}
	old := l.file
	l.file = nil
	return old.Close()
}

func (l *logger) makeDir() error {
	return os.MkdirAll(l.dir, LogDirPerm)
}

func (l *logger) needRotateOnStartup() bool {
	if l.maxFiles < 2 {
		return false
	}
	fs, err := os.Stat(path.Join(l.dir, l.filename))
	if err != nil {
		return false
	}
	return fs.ModTime().Unix()/60/60/24 != time.Now().Unix()/60/60/24
}

func (l *logger) rotateFile() error {
	if l.maxFiles < 2 {
		return nil
	}

	ext := path.Ext(l.filename)
	name := strings.TrimSuffix(l.filename, ext)

	filepath := path.Join(l.dir, fmt.Sprintf("%s.%d%s", name, l.maxFiles-1, ext))
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return err
	}

	for i := l.maxFiles - 2; i > 0; i-- {
		oldpath := path.Join(l.dir, fmt.Sprintf("%s.%d%s", name, i, ext))
		newpath := path.Join(l.dir, fmt.Sprintf("%s.%d%s", name, i+1, ext))
		if err := os.Rename(oldpath, newpath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	oldpath := path.Join(l.dir, l.filename)
	newpath := path.Join(l.dir, fmt.Sprintf("%s.%d%s", name, 1, ext))
	if err := os.Rename(oldpath, newpath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (l *logger) openFile() error {
	if l.file != nil {
		if err := l.close(); err != nil {
			log.Println("failed to close log file:", err)
		}
	}

	newfile, err := os.OpenFile(path.Join(l.dir, l.filename), os.O_CREATE|os.O_APPEND|os.O_WRONLY, LogFilePerm)
	if err != nil {
		return err
	}

	l.file = newfile
	l.openTime = time.Now()
	return nil
}

var _ io.WriteCloser = (*logger)(nil)
