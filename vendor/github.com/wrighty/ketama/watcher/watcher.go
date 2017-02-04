package watcher

import (
	"log"
	"sync"

	"gopkg.in/fsnotify.v1"
)

//Watcher will watch a given file on behalf of a Reloadable and notify it, via Reload(), if it changes
type Watcher struct {
	watching bool
	filename string
	close    chan struct{}
	mu       sync.RWMutex
	r        Reloadable
	wg       sync.WaitGroup
}

//Reloadable is the interface that any consumers of Watcher need to implement
type Reloadable interface {
	Reload() bool
}

//Make returns a Watcher that will monitor filename and notify Reloadable if it changes
func Make(filename string, r Reloadable) *Watcher {
	w := &Watcher{}
	w.close = make(chan struct{})
	w.r = r
	w.filename = filename

	w.wg.Add(1)
	go w.watchFile()
	w.wg.Wait()
	return w
}

func (w *Watcher) watchFile() {
	w.watchWithFsnotify()
}

func (w *Watcher) watchWithFsnotify() {
	watcher, err := fsnotify.NewWatcher()
	err = watcher.Add(w.filename)
	if err != nil {
		log.Fatal(err)
	}
	w.watching = true
	defer func() { w.watching = false }()
	w.wg.Done()
	for {
		select {
		case <-w.close:
			return
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				ok := w.r.Reload()
				if !ok {
					return
				}
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				return
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
