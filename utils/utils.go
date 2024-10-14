package utils

import (
	"fmt"
	"math"
	"net"
	"regexp"
)

func GenRepartCommands(percent int, blocksize string) []string {
	var maxsize uint16
	if r, _ := regexp.MatchString(`^125[0-9]{9}$`, blocksize); r {
		maxsize = 126
	} else if r, _ := regexp.MatchString(`^253[0-9]{9}$`, blocksize); r {
		maxsize = 254
	} else if r, _ := regexp.MatchString(`^509[0-9]{9}$`, blocksize); r {
		maxsize = 509
	}
	linux_max := maxsize - 12
	size := math.Round(float64(linux_max)*float64(percent)) / 100
	userdata_end := float64(maxsize) - 1 - size
	linux_end := userdata_end + size
	return []string{
		"sgdisk --resize-table 64 /dev/block/sda",
		"parted -s /dev/block/sda rm 31",
		fmt.Sprintf("parted -s /dev/block/sda mkpart userdata ext4 10.9GB 89.5GB"),
		fmt.Sprintf("parted -s /dev/block/sda mkpart esp vfat 89.5GB 90GB"),
		fmt.Sprintf("parted -s /dev/block/sda mkpart Arch ext4 90GB 160GB"),
		fmt.Sprintf("parted -s /dev/block/sda mkpart Windows fat32 160GB %vGB", maxsize),
	}
}

func GetFreePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	return port, nil
}
