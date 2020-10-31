package main

import (
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"strconv"
	"syscall"
	"os/exec"
)

func main() {
	switch os.Args[1] {
		case "run":
			run()
		case "child":
			child()
		default:
			panic("no specified args found")
	}
}

func run() {
	fmt.Printf("Running host process %v as pid %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS|syscall.CLONE_NEWPID|syscall.CLONE_NEWNS|syscall.CLONE_NEWUSER,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
		UidMappings: []syscall.SysProcIDMap {
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
	}

	must(cmd.Run())
}

func child() {
	fmt.Printf("Running child process %v as pid %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("container")))
	must(syscall.Chroot("./alpinefs"))
	must(syscall.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))

	must(cmd.Run())

	must(syscall.Unmount("proc", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")

	must(os.Mkdir(filepath.Join(pids, "ubuntu"), 0755))
	must(ioutil.WriteFile(filepath.Join(pids, "ubuntu/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place
	must(ioutil.WriteFile(filepath.Join(pids, "ubuntu/notify_on_release"), []byte("1"), 0700))
	must(ioutil.WriteFile(filepath.Join(pids, "ubuntu/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
