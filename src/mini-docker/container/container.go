package container

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

/**
 * 这里的代码是由容器内这个进程执行的
 * 这里首先mount proc 文件系统
 * + 我们去掉原本函数的参数
 */
func RunContainerInitProcess() error {
	cmdArr := readUserCommands()
	if cmdArr == nil || len(cmdArr) < 1 {
		return fmt.Errorf("Run container get user command error, Commands is nil")
	}

	setupMount()

	// 默认的syscall.exec系统调用需要传入命令的全路径，这里exec.LookPath操作是帮我们
	// 在环境变量PATH里找出命令的全路径，使得我们可以在执行run命令是输入bash也可成功，而非必须/bin/bash
	path, err := exec.LookPath(cmdArr[0])
	if err != nil {
		log.Printf("Exec loop path error %v", err)
		return err
	}

	log.Printf("Find path %s", path)

	// 这里的syscall.Exec比较重要，按照我们目前的代码看来，容器执行后执行的第一个进程并不是我们指定的进程
	// 那我们通过ps查看到的pid=1的进程不会是我们指定的进程，通过syscall.Exec 这个方法就可以解决这个问题
	// 它会覆盖当前进程的镜像，数据，和堆栈信息，包括PID，使得我们可以替换掉本身的init进程。
	if err := syscall.Exec(path, cmdArr[0:], os.Environ()); err != nil {
		log.Printf(err.Error())
	}
	return nil
}

func setupMount() {

	pwd, err := os.Getwd()
	if (err != nil) {
		fmt.Printf("[Error] Get current location error %v", err)
		return
	}
	fmt.Printf("Current location is %s", pwd)
	pivotRoot(pwd)

	// mount proc
	// MS_NOEXEC 在本文件系统下不允许运行其他程序
	// MS_NOSUID 不允许set uid 或 set gid
	// MS_NODEV  这个是自linux 2.4以来，所有mount都会默认设置的一个参数
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")

}

func readUserCommands() []string {
	//uintptr(3)指的是index=3的文件描述符，这个就是前面传递进来的管道
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		//log.Errorf("init read pipe error %v", err)
		log.Printf("[ERROR] init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

/**
 * /proc/self/exe 指代的是当前进程本身，通过这种方式对创建出来的进程进行初始化
 * 使得新进程在 namespace 隔离中，同时如果指定tty=true的话，将当前进程的输入输出转到新进程
 * 当前command对象启动后会调用initCommand命令，并传入args参数
 *
 * + 这里去掉了原来的command参数，因为已经不需要传进来参数了
 * + 同时函数增加返回值 writePipe 对象
 */
func NewParentProcess(tty bool) (*exec.Cmd, *os.File ) {
	readPipe, writePipe, err := NewPipe()

	if (err != nil) {
		log.Printf("New pipe error %v", err)
		return nil, nil
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 如果用户指定了 -it 参数, 就把当前进程的输入输出导入到标准输入输出上
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// 这里的作用是将创建的管道传给子进程
	// cmd.ExtraFiles的意思是带着额外的文件句柄去创建进程
	// 默认的进程会带着 标准输入，标准输出，标准错误这3个文件描述符
	// 这里我们需要把第四个文件描述符传给子进程，通过此来传递命令行参数
	// 可以通过/proc/self/fd查看当前文件描述符
	cmd.ExtraFiles = []*os.File{readPipe}
	//cmd.Dir = "/root/busybox"
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	NewWorkSpace(rootURL, mntURL)
	cmd.Dir = mntURL

	return cmd, writePipe
}

// 生成匿名管道
func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

func pivotRoot(root string) error {
	/**
	  为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}
	// 创建 rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// pivot_root 到新的rootfs, 现在老的 old_root 是挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}


//创建AUFS文件系统
func NewWorkSpace(rootURL string, mntURL string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)
}

func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exist, err := PathExists(busyboxURL)
	if err != nil {
		log.Printf("Fail to judge whether dir %s exists. %v", busyboxURL, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			log.Printf("Mkdir dir %s error. %v", busyboxURL, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			log.Printf("Untar dir %s error %v", busyboxURL, err)
		}
	}
}

func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		log.Printf("Mkdir dir %s error. %v", writeURL, err)
	}
}


func CreateMountPoint(rootURL string, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		log.Printf("Mkdir dir %s error. %v", mntURL, err)
	}
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("%v", err)
	}
}

//Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootURL string, mntURL string){
	DeleteMountPoint(rootURL, mntURL)
	DeleteWriteLayer(rootURL)
}

func DeleteMountPoint(rootURL string, mntURL string){
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout=os.Stdout
	cmd.Stderr=os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("%v",err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Printf("Remove dir %s error %v", mntURL, err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Printf("Remove dir %s error %v", writeURL, err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
