package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/ironsmile/logger"
)

func main() {
	setUpDefaultLogger()

	if len(os.Args) < 3 {
		logger.Fatalf("Usage: %s FILE-TO-WATCH COMMAND-OR-EXECUTABLE", os.Args[0])
	}

	watched := os.Args[1]

	cmd := command{
		cmd:  os.Args[2],
		args: os.Args[3:],
	}

	st, err := os.Stat(watched)
	if err != nil {
		logger.Fatalf("Error statting file %s: %s", watched, err)
	}

	watchingDir := st.IsDir()
	watchedFull, err := filepath.Abs(watched)
	if err != nil {
		logger.Fatalf("Cannot resolve watched full path: %s", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatalf("Creating watch failed: %s", err)
	}
	defer watcher.Close()

	watchedDir := watchedFull
	if !watchingDir {
		watchedDir = filepath.Dir(watched)
	}

	if err := watcher.Add(watchedDir); err != nil {
		logger.Fatalf("Watching %s failed: %s", watchedDir, err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			var (
				isWrite  = event.Op&fsnotify.Write == fsnotify.Write
				isCreate = event.Op&fsnotify.Create == fsnotify.Create
			)

			if !isWrite && !isCreate {
				continue
			}

			if !watchingDir && event.Name != watchedFull {
				continue
			}

			runCommand(cmd)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Errorf("Watch error: %s", err)
		}
	}
}

func setUpDefaultLogger() {
	lg := logger.Default()
	lg.Errorer.SetFlags(lg.Errorer.Flags() ^ (log.Ldate | log.Ltime))
	lg.Logger.SetFlags(lg.Errorer.Flags())
	lg.Debugger.SetFlags(lg.Errorer.Flags())
	lg.Logger.SetPrefix("")
}

type command struct {
	cmd  string
	args []string
}

func runCommand(info command) {
	cmd := exec.Command(info.cmd, info.args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Errorf("Getting command stdout: %s", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Errorf("Getting command stderr: %s", err)
		return
	}

	if err := cmd.Start(); err != nil {
		logger.Errorf("Failed to start %s: %s", info.cmd, err)
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	if err := cmd.Wait(); err != nil {
		stdout.Close()
		stderr.Close()

		logger.Errorf("Program finished with errors: %s", err)
	}
}
