package zupdate

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Batch struct {
	file *os.File
	w    *bufio.Writer
	path string
}

func NewBatch(path string) Batch {
	var retBat Batch
	var err error
	retBat.path = path
	retBat.file, err = os.OpenFile(path+".bat", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("failed")
		os.Exit(1004)
	}
	retBat.w = bufio.NewWriter(retBat.file)
	retBat.w.WriteString(fmt.Sprintf("taskkill /f /pid %d\n", os.Getpid()))
	return retBat
}

func (bat *Batch) Add(name string) {
	bat.w.WriteString(fmt.Sprintf("del %s.exe\n", name))
	bat.w.WriteString(fmt.Sprintf("rename %s.exe_tmp %s.exe\n", name, name))
}

func (bat *Batch) Entry(title, name string, runArgs ...string) {
	argsCommand := ""
	for _, arg := range runArgs {
		argsCommand += arg + " "
	}
	bat.w.WriteString(fmt.Sprintf("start \"%s\" %s.exe %s\n", title, name, strings.TrimRight(argsCommand, " ")))
	bat.w.WriteString("exit\n")
	bat.w.Flush()
	bat.file.Close()
}

func (bat *Batch) Exec() error {
	return exec.Command("cmd.exe", "/c", "start "+bat.path+".bat").Start()
}

func ExecByName(path string) error {
	return exec.Command("cmd.exe", "/c", "start "+path+".bat").Start()
}
