package nvme

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

type NvmfDevice struct {
	nqn     string
	address string
	port    uint16
}

const (
	SubsystemPath = "/sys/class/nvme-subsystem/"
	b
)

type blockDevice struct {
	major int
	minor int
}

func checkContents(path string, value string) (bool, error) {
	contents, err := ioutil.ReadFile(path)

	if err != nil {
		return false, err
	}

	line := strings.Replace(string(contents), "\n", "", -1)

	return line == value, nil
}

func getDevice(path string) *BlockDevice {
	glob := fmt.Sprintf("%sn*/dev", path)

	devFiles, err := filepath.Glob(glob)

	if len(devFiles) == 0 {
		end := time.Now().Add(5 * time.Second)

		for len(devFiles) == 0 && time.Now().Before(end) {
			time.Sleep(10 * time.Millisecond)

			devFiles, err = filepath.Glob(glob)
		}
	}

	if err != nil {
		fmt.Printf("%e\n", err)
	}

	blockDev := &BlockDevice{Major: -1, Minor: -1}

	for _, devFile := range devFiles {
		contents, err := ioutil.ReadFile(devFile)

		if err != nil {
			continue
		}

		majorminor := strings.Replace(string(contents), "\n", "", -1)
		mm := strings.Split(majorminor, ":")

		if len(mm) == 2 {
			blockDev.Major, _ = strconv.Atoi(mm[0])
			blockDev.Minor, _ = strconv.Atoi(mm[1])
		}
	}

	return blockDev
}

func findDevice(nqn, address string, port uint16, instance, cntlid string) (*BlockDevice, error) {
	glob := fmt.Sprintf("%s/*/*/cntlid", SubsystemPath)

	cntlidPaths, err := filepath.Glob(glob)

	if err != nil {
		fmt.Printf("%e\n", err)
	}

	for _, cntlidPath := range cntlidPaths {
		if ok, _ := checkContents(cntlidPath, cntlid); ok {

			blockDev := getDevice(filepath.Dir(cntlidPath))

			if blockDev.Major != -1 && blockDev.Minor != -1 {

				return blockDev, nil
			}
		}

	}

	return nil, nil
}

func RegisterDevice(name, nqn, address string, port uint16) (string, error) {
	file, err := OpenDevice()

	if err != nil {
		return "", err
	}

	defer file.Close()

	command := fmt.Sprintf("nqn=%s,transport=tcp,traddr=%s,trsvcid=%d\000", nqn, address, port)

	_, err = file.Write([]byte(command))

	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(bufio.NewReader(file))

	if !scanner.Scan() {
		return "", err
	}

	line := scanner.Text()

	fields := strings.Split(line, ",")

	instance := ""
	cntlid := ""

	for _, field := range fields {
		subfields := strings.Split(field, "=")

		if len(subfields) == 2 {
			if subfields[0] == "instance" {
				instance = subfields[1]
			} else if subfields[0] == "cntlid" {
				cntlid = subfields[1]
			}
		}
	}

	if instance == "" && cntlid == "" {
		return "", fmt.Errorf("nvme-fabrics returned invalid string")
	}

	blockDev, err := findDevice(nqn, address, port, instance, cntlid)

	if blockDev == nil || err != nil {
		return "", err
	}

	return blockDev.createDev(name)
}

func deleteBlockDevice(devFile string) {
	subsystem := filepath.Dir(devFile)
	subsystem = filepath.Dir(subsystem)

	glob := fmt.Sprintf("%s/*/delete_controller", subsystem)

	delFiles, err := filepath.Glob(glob)

	if err != nil {
		fmt.Printf("%e\n", err)
	}

	for _, delFile := range delFiles {
		file, err := os.OpenFile(delFile, os.O_WRONLY, 0644)

		if err != nil {
			fmt.Printf("%e", err)
			continue
		}

		defer file.Close()
		_, err = file.Write([]byte("1"))

	}
}

func findBlockDevice(majorminor string) string {
	glob := fmt.Sprintf("%s*/*n*/dev", SubsystemPath)

	devFiles, err := filepath.Glob(glob)

	if err != nil {
		fmt.Printf("%e\n", err)
	}

	for _, devFile := range devFiles {
		if ok, _ := checkContents(devFile, majorminor); ok {
			return devFile
		}
	}

	return ""
}

func UnregisterDevice(name string) {
	var stat unix.Stat_t

	path := DevPath + name
	unix.Stat(path, &stat)

	os.Remove(path)

	majorminor := fmt.Sprintf("%d:%d", unix.Major(stat.Rdev), unix.Minor(stat.Rdev))

	devFile := findBlockDevice(majorminor)

	if devFile != "" {
		deleteBlockDevice(devFile)
	}

}
