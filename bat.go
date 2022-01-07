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
	name string
}

func NewBatch(name string) *Batch {
	var retBat Batch
	var err error
	retBat.name = name
	retBat.file, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("failed")
		os.Exit(1004)
	}
	retBat.w = bufio.NewWriter(retBat.file)
	retBat.w.WriteString(fmt.Sprintf("taskkill /f /pid %d\n", os.Getpid()))
	return &retBat
}

func (bat *Batch) Add(name string) {
	// notice: param name in here already has ext name(.exe)
	bat.w.WriteString(fmt.Sprintf("del %s\n", name))
	bat.w.WriteString(fmt.Sprintf("rename %s_tmp %s\n", name, name))
}

func (bat *Batch) Entry(name string, runArgs ...string) {
	argsCommand := ""
	for _, arg := range runArgs {
		argsCommand += arg + " "
	}
	bat.w.WriteString(fmt.Sprintf("start \"%s\" %s.exe %s\n", name, name, strings.TrimRight(argsCommand, " ")))
	bat.w.WriteString("exit\n")
	bat.w.Flush()
	bat.file.Close()
}

func (bat *Batch) Exec() error {
	return exec.Command("cmd.exe", "/c", "start "+bat.name).Start()
}

func ExecByName(name string) error {
	return exec.Command("cmd.exe", "/c", "start "+name).Start()
}
