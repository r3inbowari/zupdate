package zupdate

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

const (
	TypeExec = "exec"
	TypeFile = "file"
)

type File struct {
	Name           string            `json:"name"`
	Type           Type              `json:"type"`
	Major          int64             `json:"major"`
	Minor          int64             `json:"minor"`
	Patch          int64             `json:"patch"`
	Digest         map[string]string `json:"digests"`
	DownloadSource string            `json:"source"`
	Reload         bool              `json:"reload"`
	Desc           string            `json:"desc"`
	Path           string            `json:"path"` // relative path
}

func (f *File) String() string {
	return fmt.Sprintf("%s(v%d.%d.%d)", f.Name, f.Major, f.Minor, f.Patch)
}

func (f *File) Compare(remoteFile *File) *File {
	remote := remoteFile.Major<<24 + remoteFile.Minor<<12 + remoteFile.Patch<<0
	current := f.Major<<24 + f.Minor<<12 + f.Patch<<0

	if current < remote {
		return remoteFile
	}
	return nil
}

func (f *File) Check() error {
	resolvePath := fmt.Sprintf(".%s/%s", f.Path, f.Name)
	return f.checkDigest(resolvePath)
}

func (f *File) Verify() error {
	resolvePath := fmt.Sprintf(".%s/%s_tmp", f.Path, f.Name)
	return f.checkDigest(resolvePath)
}

func (f *File) checkDigest(resolvePath string) error {
	fp, err := os.Open(resolvePath)
	if err != nil {
		return err
	}
	defer fp.Close()
	md5f := md5.New()
	_, err = io.Copy(md5f, fp)
	if err != nil {
		return err
	}

	expectedDigest := ""
	ok := false
	if f.Type == TypeFile {
		expectedDigest, ok = f.Digest[string(f.Type)]
	} else if f.Type == TypeExec {
		expectedDigest, ok = f.Digest[fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)]
	}
	if !ok {
		return errors.New("file digest not exist")
	}

	if ok = expectedDigest != hex.EncodeToString(md5f.Sum([]byte(""))); ok {
		return errors.New("file digest mismatch")
	}
	return err
}

func (f *File) DownloadExec() error {
	var bar *ProgressBar
	dUrl := ""
	if f.Type == TypeFile {
		dUrl = f.DownloadSource
	} else if f.Type == TypeExec {
		dUrl = fmt.Sprintf(f.DownloadSource, runtime.GOOS, runtime.GOARCH)
	}
	err := download(dUrl, f.Name+"_tmp", func(fileLength int64) {
		bar = NewProgressBar(fileLength)
	}, func(length, downLen int64) {
		bar.Play(downLen)
	})
	defer func() {
		if bar != nil {
			bar.Finish()
		}
	}()
	return err
}

func download(url, name string, sizeReport func(fileLength int64), fb func(length, downLen int64)) error {
	var (
		fileSize int64
		buf      = make([]byte, 32*1024)
		written  int64
	)
	client := new(http.Client)
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	fileSize, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		fmt.Println(err)
	}

	sizeReport(fileSize)

	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	if resp.Body == nil {
		return errors.New("empty body")
	}
	defer resp.Body.Close()
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := file.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
		fb(fileSize, written)
	}
	return err
}
