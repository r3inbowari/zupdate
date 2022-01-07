package zupdate

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Mode int
type Type string

const (
	DISABLE Mode = 1 >> 1
	DEV     Mode = 1 << 0
	REL     Mode = 1 << 1
	ALI     Mode = 1 << 2
)

type Update struct {
	file       map[string]File
	remoteFile map[string]File
	pending    []File
	sync.RWMutex
	UpdateOptions
	batch Batch
}

type UpdateOptions struct {
	EntryName   string
	Mode        Mode
	Log         *logrus.Logger
	CheckSource string
	Callback    func()
	EntryArgs   []string
}

func (up *Update) IncludeFile(id string, f File) {
	up.Lock()
	defer up.Unlock()
	up.file[id] = f
}

func InitUpdater(opt UpdateOptions) *Update {
	if runtime.GOOS == "windows" {
		_ = os.Remove(".install.bat")
	}
	var update Update
	update.UpdateOptions = opt
	update.file = make(map[string]File)
	if opt.Log == nil {
		update.Log = logrus.New()
	}
	if opt.CheckSource == "" {
		update.Log.Warn("update check source not set")
	}
	return &update
}

func (up *Update) SetCallback(callback func()) {
	up.Lock()
	defer up.Unlock()
	up.Callback = callback
}

func (up *Update) Update(file *File) (bool, error) {

	// check local file
	err := file.Check()
	if err == nil {
		// already new
		return file.Reload, nil
	}

	up.Log.WithFields(logrus.Fields{"name": file.Name}).Info(fmt.Sprintf("[UP] updating... -> %s", file.String()))

	// download
	err = file.DownloadExec()
	if err != nil {
		up.Log.WithFields(logrus.Fields{"name": file.Name, "err": err.Error()}).Error("[UP] download failed")
		return false, err
	}

	// verify
	err = file.Verify()
	if err != nil {
		up.Log.WithFields(logrus.Fields{"name": file.Name, "err": err.Error()}).Error("[UP] file digest mismatch")
		return false, err
	}

	// install
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// we can rename the all file with open with other process on linux/unix
		err := os.Rename(file.Name+"_tmp", file.Name)
		if err != nil {
			up.Log.WithFields(logrus.Fields{"name": file.Name, "err": err.Error()}).Error("[UP] file rename failed")
			return false, err
		}
	} else if runtime.GOOS == "windows" && !file.Reload {
		// we can rename all files, but excluded files which will reload
		err = os.Rename(file.Name+"_tmp", file.Name)
		if err != nil {
			up.Log.WithFields(logrus.Fields{"name": file.Name, "err": err.Error()}).Error("[UP] can not rename this file")
			return false, err
		}
	} else if runtime.GOOS == "windows" && file.Reload {
		// we try to rename excluded file on batch when using windows,
		// but this process must be defined already on batch
		up.batch.Add(file.Name)
	}
	return file.Reload, err
}

func (up *Update) Check() {
	err := up.pullUpdateList()
	if err != nil {
		return
	}
	for id, file := range up.file {
		if v, ok := up.remoteFile[id]; ok {
			if f := file.Compare(&v); f != nil {
				up.Log.WithField("id", id).Info(fmt.Sprintf("[UP] found patch: %s -> %s", file.String(), v.String()))
				if runtime.GOOS == "windows" && f.Type == "exec" {
					f.Name += ".exe"
				}
				up.pending = append(up.pending, *f)
			}
		}
	}
}

func (up *Update) CheckAndUpdate() {
	needReload := false
	up.Check()
	up.batch = NewBatch(".install.bat")
	for _, file := range up.pending {
		nReload, err := up.Update(&file)
		if err != nil {
			continue
		}
		if nReload {
			needReload = true
		}
	}

	// finally
	if needReload {
		up.Reload()
	}
}

func (up *Update) Reload() {
	systemType := runtime.GOOS
	up.Log.WithFields(logrus.Fields{"os": runtime.GOOS, "arch": runtime.GOARCH}).Info("[UP] reloading")
	if up.Callback != nil {
		up.Lock()
		up.Callback()
		up.Unlock()
	}
	if systemType == "linux" || systemType == "darwin" {
		// release current app resource.
		_ = reload(up.EntryName, up.EntryArgs...)
		time.Sleep(time.Second)
		os.Exit(1010)
	} else if systemType == "windows" {
		up.batch.Entry(up.EntryName, up.EntryArgs...)
		// it will kill current process,
		// release resource is no need.
		_ = up.batch.Exec()
		time.Sleep(time.Minute)
	}
}

func (up *Update) CheckAndUpdateWithGap(cronSpec ...string) {
	spec := "0 0 12 * * ?"
	if len(cronSpec) > 0 {
		spec = cronSpec[0]
	}
	if up.Mode != DISABLE {
		up.CheckAndUpdate()
		cm := cron.New()
		_ = cm.AddFunc(spec, func() {
			time.Sleep(time.Second)
			up.CheckAndUpdate()
		})
		cm.Start()
	}
}

func (up *Update) pullUpdateList() error {
	res, err := http.Get(up.CheckSource)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &up.remoteFile)
	return err
}

func reload(path string, runArgs ...string) error {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		exec.Command("chmod", "777", path)
		path = "./" + path
	}
	// init 接管
	argsCommand := ""
	for _, arg := range runArgs {
		argsCommand += arg + " "
	}
	cmd := exec.Command(path, strings.TrimRight(argsCommand, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
