package nvme

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	NvmeDevPath       = "/dev/longhorn/nvme-fabrics"
	NvmeFabricsDevice = "/sys/devices/virtual/misc/nvme-fabrics/dev"
)

func OpenDevice() (*os.File, error) {
	if _, err := os.Stat(DevPath); os.IsNotExist(err) {
		if err := os.MkdirAll(DevPath, 0755); err != nil {
			logrus.Fatalf("Cannot create directory %v", DevPath)
		}
	}

	if _, err := os.Stat(NvmeDevPath); err != nil {
		if err := createDevice(); err != nil {
			return nil, err
		}
	}

	return os.OpenFile(NvmeDevPath, os.O_RDWR, 0600)
}

func createDevice() error {
	contents, err := ioutil.ReadFile(NvmeFabricsDevice)

	if err != nil {
		return err
	}

	majorminor := strings.Replace(string(contents), "\n", "", -1)
	mm := strings.Split(majorminor, ":")

	if len(mm) == 2 {
		major, _ := strconv.Atoi(mm[0])
		minor, _ := strconv.Atoi(mm[1])

		return mknodChar(NvmeDevPath, major, minor)
	}

	return fmt.Errorf("Cannot find nvme-fabrics device")
}

func mknodChar(device string, major, minor int) error {
	var fileMode os.FileMode = 0660
	fileMode |= unix.S_IFCHR
	dev := int(unix.Mkdev(uint32(major), uint32(minor)))

	logrus.Infof("Creating device %s %d:%d", device, major, minor)
	return unix.Mknod(device, uint32(fileMode), dev)
}
