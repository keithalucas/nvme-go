package nvme

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

const (
	DevPath = "/dev/longhorn/"
)

type BlockDevice struct {
	Major int
	Minor int
}

func (bd *BlockDevice) createDev(name string) (string, error) {
	if _, err := os.Stat(DevPath); os.IsNotExist(err) {
		if err := os.MkdirAll(DevPath, 0755); err != nil {
			logrus.Fatalf("device %v: Cannot create directory %v", name, DevPath)
		}
	}

	dev := fmt.Sprintf("%s%s", DevPath, name)

	if _, err := os.Stat(dev); err == nil {
		logrus.Warnf("Device %s already exists, clean it up", dev)
		if err := os.Remove(dev); err != nil {
			return "", errors.Wrapf(err, "cannot cleanup block device file %v", dev)
		}
	}

	if err := DuplicateDevice(bd, dev); err != nil {
		return "", err
	}

	logrus.Debugf("device %v: Device %s is ready", name, dev)

	return dev, nil
}

func DuplicateDevice(dev *BlockDevice, dest string) error {
	if err := mknod(dest, dev.Major, dev.Minor); err != nil {
		return fmt.Errorf("Cannot create device node %s for device", dest)
	}
	if err := os.Chmod(dest, 0660); err != nil {
		return fmt.Errorf("Couldn't change permission of the device %s: %w", dest, err)
	}
	return nil
}

func mknod(device string, major, minor int) error {
	var fileMode os.FileMode = 0660
	fileMode |= unix.S_IFBLK
	dev := int(unix.Mkdev(uint32(major), uint32(minor)))

	logrus.Infof("Creating device %s %d:%d", device, major, minor)
	return unix.Mknod(device, uint32(fileMode), dev)
}
