package cmd

import (
	"io"
	"io/fs"
	"lon-tool/image"
	"lon-tool/utils"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/timoxa0/goadb/adb"
	"github.com/timoxa0/gofastboot/fastboot"
)

var username string
var password string
var serail string
var partsize string
var partpercent int
var deployCmd = &cobra.Command{
	Use:   "deploy <rootfs.lni>",
	Short: "Deploy system to device",
	Long:  "Deploy system to device in fastboot mode",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var msg string

		req_repartition := partsize != ""

		image, close, err := image.ReadImage(args[0])
		defer close()
		if err != nil {
			logger.Error("Failed to read image info:", logger.Args("Error", err))
			os.Exit(1)
		}

		adbc, err := adb.New()
		if err != nil {
			logger.Fatal("Failed to get adb client", logger.Args(err))
		}
		fb_devs, err := fastboot.FindDevices()
		logger.Debug("Devices", logger.Args("Devices", fb_devs, "err", err))
		if err != nil {
			logger.Fatal("Failed to get fastboot device", logger.Args(err))
		}

		if len(fb_devs) < 1 {
			logger.Fatal("No fastboot devices found")
			os.Exit(170)
		}

		if serail == "autodetect" {
			for _, dev := range fb_devs {
				product, err := dev.GetVar("product")
				devSerial, _ := dev.Device.SerialNumber()
				if err != nil {
					logger.Warn("Unable to communicate with device", logger.Args("Serial", devSerial))
					logger.Debug("Unable to communicate with device", logger.Args("err", err))
					continue
				}
				logger.Debug("Found device", logger.Args("Product", product, "Sraial", devSerial, "object", dev))
				if product == "nabu" {
					logger.Debug("Nabu found", logger.Args("Serial", devSerial))
					serail = devSerial
					if !req_repartition {
						_, err1 := dev.GetVar("partition-type:linux")
						_, err2 := dev.GetVar("partition-type:esp")
						req_repartition = err1 != nil && err2 != nil
					}
					break
				}
			}
			for _, dev := range fb_devs {
				dev.Close()
			}
			if serail == "autodetect" {
				logger.Fatal("Nabu in fastboot mode not found")
				os.Exit(170)
			}
		}

		for {
			if username != "" {
				break
			}
			username, _ = pterm.DefaultInteractiveTextInput.Show("Linux username")
			if r, _ := regexp.MatchString(`[A-z0-9]+`, username); !r {
				username = ""
				logger.Error("Username must contain only latin letter or digits")
			}
		}

		for {
			if password != "" {
				break
			}
			password, _ = pterm.DefaultInteractiveTextInput.WithMask("*").Show("Linux password")
			if r, _ := regexp.MatchString(`[A-z0-9$%#@!]+`, password); !r {
				password = ""
				logger.Error("Password must contain only latin letter, digits or $%#@!")
			}
		}

		var run_repartition bool
		if req_repartition {
			run_repartition, _ = pterm.DefaultInteractiveConfirm.WithDefaultValue(true).Show("Found incompatible partition table. Do you want to change it?")
		} else {
			run_repartition, _ = pterm.DefaultInteractiveConfirm.Show("Found compatible partition table. Do you want to change it?")
		}
		if req_repartition && !run_repartition {
			pterm.Println("Bye")
			os.Exit(0)
		}

		for {
			if partsize != "" || !run_repartition {
				break
			}
			partsize, _ = pterm.DefaultInteractiveTextInput.WithDefaultValue("50%").Show("Partition size [20;90]%")
			v, err := strconv.Atoi(strings.TrimRight(partsize, "%"))
			partpercent = v
			if !(err == nil && 20 <= v && v <= 90 && partsize[len(partsize)-1] == '%') {
				partsize = ""
				logger.Error("Invalid partsize. Parsize must be in percents")
			}
		}

		msg, _ = pterm.DefaultTree.WithRoot(pterm.TreeNode{
			Text: pterm.Green("System"),
			Children: []pterm.TreeNode{
				{Text: pterm.Sprintf("Name: %s", image.Name)},
				{Text: pterm.Sprintf("Version: %s", image.Version)},
			},
		}).Srender()
		pterm.Print(msg)
		msg, _ = pterm.DefaultTree.WithRoot(pterm.TreeNode{
			Text: pterm.Green("Settings"),
			Children: []pterm.TreeNode{
				{Text: pterm.Sprintf("Username: %s", username)},
				{Text: pterm.Sprintf("Password: %s", password)},
				{Text: pterm.Sprintf("Device serial: %s", serail)},
				{Text: pterm.Sprintf("Run repatition: %v", map[bool]string{
					true:  pterm.Green("Yes"),
					false: pterm.Red("No"),
				}[run_repartition])},
				{Text: pterm.Sprintf("Partition size: %s", map[bool]string{
					true:  "Not changed",
					false: partsize,
				}[partsize == ""])},
			},
		}).Srender()
		pterm.Print(msg)
		var isok bool
		if run_repartition {
			isok, _ = pterm.DefaultInteractiveConfirm.Show("Start installation? All user data will be ERASED")
		} else {
			isok, _ = pterm.DefaultInteractiveConfirm.Show("Start installation?")
		}
		if !isok {
			pterm.Println("Bye")
			os.Exit(253)
		}
		fb_dev, err := fastboot.FindDevice(serail)
		if err != nil {
			logger.Error("Unable to find device", logger.Args("object", fb_dev, "err", err))
			fb_devs, err := fastboot.FindDevices()
			logger.Debug("Devices", logger.Args("Devices", fb_devs, "err", err))
			os.Exit(255)
		}

		adbd := adbc.Device(adb.DeviceWithSerial(serail))
		bootdata, err := utils.Files.OrangeFox.Get(*pbar.WithTitle("Downloading orangefox"))
		if err != nil {
			if bootdata != nil {
				logger.Warn("Unable to verify recovery image")
			} else {
				logger.Error("Unable to download recovery image")
				os.Exit(179)
			}
		}

		if run_repartition {
			gpt, err := utils.Files.GPT.Get(*pbar.WithTitle("Downloading default partition table"))
			if err != nil {
				if gpt != nil {
					logger.Warn("Unable to verify gpt image")
				} else {
					logger.Error("Unable to download gpt image")
					os.Exit(179)
				}
			}

			userdata, err := utils.Files.CleanUserdata.Get(*pbar.WithTitle("Downloading empty userdata"))
			if err != nil {
				if userdata != nil {
					logger.Warn("Unable to verify userdata image")
				} else {
					logger.Error("Unable to download userdata image")
					os.Exit(179)
				}
			}

			if e := fb_dev.Flash("partition:0", gpt); e != nil {
				logger.Error("Failed to flash GPT. Reflash stock rom and try again")
				logger.Debug("Error", logger.Args("e", e))
				os.Exit(180)
			}

			if e := fb_dev.Flash("userdata", userdata); e != nil {
				logger.Error("Failed to clean userdata. Reflash stock rom and try again")
				logger.Debug("Error", logger.Args("e", e))
				os.Exit(180)
			}

			ofoxSpinner, _ := spinner.Start("Booting recovery")
			if e := fb_dev.BootImage(bootdata); e != nil {
				logger.Error("Failed to boot recovery. Reflash stock rom and try again")
				logger.Debug("Error", logger.Args("e", e))
				os.Exit(177)
			}

			for i := 0; i <= 120; i++ {
				if i == 120 {
					ofoxSpinner.Stop()
					logger.Error("Recovery device timeout")
					os.Exit(173)
				}
				if s, _ := adbd.State(); s == adb.StateRecovery {
					ofoxSpinner.Stop()
					break
				}
				time.Sleep(time.Second)
			}

			block_size, _ := adbd.RunCommand("blockdev --getsize64 /dev/block/sda")
			block_size = strings.TrimRight(block_size, "\n")
			is128 := false
			if r, _ := regexp.MatchString(`^125[0-9]{9}$`, block_size); r {
				is128 = true
			} else if r, _ := regexp.MatchString(`^253[0-9]{9}$`, block_size); r {
				is128 = false
			}

			for _, cmd := range utils.GenRepartCommands(partpercent, is128) {
				adbd.RunCommand(cmd)
				logger.Debug("Executed command", logger.Args("cmd", cmd))
			}

			adbd.RunCommand("reboot bootloader")
			fbSpinner, _ := spinner.Start("Rebooting to bootloader")
			for i := 0; i <= 120; i++ {
				if i == 120 {
					fbSpinner.Stop()
					logger.Error("Fastboot device timeout")
					os.Exit(172)
				}
				if d, e := fastboot.FindDevice(serail); e == nil {
					fbSpinner.Stop()
					fb_dev = d
					break
				}
				time.Sleep(time.Second)
			}

			if e := fb_dev.Flash("userdata", userdata); e != nil {
				logger.Error("Failed to clean userdata. Reflash stock rom and try again")
				logger.Debug("Error", logger.Args("e", e))
				os.Exit(180)
			}
			logger.Info("Repartiton done")
		}

		ofoxSpinner, _ := spinner.Start("Booting recovery")
		fb_dev.BootImage(bootdata)
		for i := 0; i <= 120; i++ {
			if i == 120 {
				ofoxSpinner.Stop()
				logger.Error("Recovery device timeout")
				os.Exit(173)
			}
			if s, _ := adbd.State(); s == adb.StateRecovery {
				ofoxSpinner.Stop()
				break
			}
			time.Sleep(time.Second)
		}

		port, err := utils.GetFreePort()
		if err != nil {
			logger.Error("Failled to find free tcp port")
			os.Exit(181)
		}
		logger.Debug("Flasher", logger.Args("port", port))

		adbd.Forward(pterm.Sprintf("tcp:%v", port), "tcp:4444")
		logger.Debug("ListForwards", logger.Args(adbd.ListForwards()))
		doneChan1 := make(chan bool)
		doneChan2 := make(chan bool)
		go func() {
			adbd.RunCommand("busybox nc -w 10 -l 127.0.0.1:4444 > /dev/block/platform/soc/1d84000.ufshc/by-name/linux 2> /tmp/nclog.txt")
			logger.Debug("Busybox closed")
			doneChan1 <- true
		}()
		time.Sleep(time.Second * 3)

		go func() {
			conn, err := net.Dial("tcp", pterm.Sprintf("127.0.0.1:%v", port))
			if err != nil {
				logger.Error("Failled to connect to device")
			}
			buf := make([]byte, 409600)
			bar, _ := pbar.WithTotal(int(image.ImgSize)).WithTitle("Flashing rootfs").WithRemoveWhenDone(false).Start()
			dataLegth := 0
			for {
				n, err := image.Reader.Read(buf)
				dataLegth += n
				if err != nil && err != io.EOF && dataLegth < int(image.ImgSize) {
					logger.Error("Failed to read from GZ reader", logger.Args("error", err))
					return
				}
				if n == 0 {
					break
				}
				bar.Add(n)
				n, err = conn.Write(buf[:n])
				if n == 0 {
					logger.Error("Failed to write to device. Reflash stock rom and try again", logger.Args("err", err))
					adbd.RunCommand("reboot bootloader")
					os.Exit(179)
					break
				}
			}
			conn.Close()
			<-doneChan1
			doneChan2 <- true
		}()
		<-doneChan2
		adbd.KillForwardAll()

		out, err := adbd.RunCommand(pterm.Sprintf("postinstall %s %s > /dev/null 2>&1; echo $?", username, password))
		out = strings.TrimRight(out, "\n")
		logger.Debug("Postinstall", logger.Args("out", out, "err", err))
		if out != "0" || err != nil {
			logger.Error("Postinstall failed. Reflash stock rom and try again", logger.Args("Device error", out, "Go error", err))
		} else {
			logger.Info("System cofigured")
		}

		adbd.RunCommand("mkdir /tmp/uefi-install")
		uploadBar, _ := pbar.WithTotal(2).WithTitle("Uploading uefi files").Start()

		bootshim, err := utils.Files.UEFIBootshim.Get(*pbar.WithTitle("Downloading uefi bootshim"))
		if err != nil {
			if bootshim != nil {
				logger.Warn("Unable to verify uefi bootship image")
			} else {
				logger.Error("Unable to download uefi bootshim image")
				os.Exit(179)
			}
		}

		conn, err := adbd.OpenWrite(pterm.Sprintf("/tmp/uefi-install/%s", utils.Files.UEFIBootshim.Name), fs.FileMode(0777), adb.MtimeOfClose)
		if err != nil {
			uploadBar.Stop()
			logger.Error("Failed to send uefi bootshim", logger.Args("Error", err))
		}
		_, err = conn.Write(bootshim)
		if err != nil {
			uploadBar.Stop()
			logger.Error("Failed to send uefi bootshim", logger.Args("Error", err))
		}
		conn.Close()
		uploadBar.Add(1)

		payload, err := utils.Files.UEFIPayload.Get(*pbar.WithTitle("Downloading uefi payload"))
		if err != nil {
			if payload != nil {
				logger.Warn("Unable to verify uefi payload image")
			} else {
				logger.Error("Unable to download uefi payload image")
				os.Exit(179)
			}
		}
		conn, err = adbd.OpenWrite(pterm.Sprintf("/tmp/uefi-install/%s", utils.Files.UEFIPayload.Name), fs.FileMode(0777), adb.MtimeOfClose)
		if err != nil {
			uploadBar.Stop()
			logger.Error("Failed to send uefi payload", logger.Args("Error", err))
		}
		_, err = conn.Write(payload)
		if err != nil {
			uploadBar.Stop()
			logger.Error("Failed to send uefi payload", logger.Args("Error", err))
		}
		conn.Close()
		uploadBar.Add(1)

		uefiSpinner, _ := spinner.Start("Patching UEFI")
		out, err = adbd.RunCommand("uefi-patch  > /dev/null 2>&1; echo $?")
		out = strings.TrimRight(out, "\n")
		if err != nil {
			logger.Error("Failed to install uefi. Reflash stock rom and try again", logger.Args("Error", err))
			os.Exit(176)
		}
		logger.Debug("Uefi patch", logger.Args("Out", out))

		switch out {
		case "1":
			uefiSpinner.Stop()
			logger.Error("Failed to install uefi. Reflash stock rom and try again", logger.Args("Error", err))
			adbd.RunCommand("reboot bootloader")
			os.Exit(176)
		case "2":

			adbd.RunCommand("reboot")
			uefiSpinner.Stop()
			logger.Info("Bootimage already patched")
		case "0":
			adbd.RunCommand("reboot")
			uefiSpinner.Stop()
			logger.Info("Installation done!")
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(&username, "username", "u", "", "User name")
	deployCmd.Flags().StringVarP(&password, "password", "p", "", "User password")
	deployCmd.Flags().StringVarP(&serail, "serial", "s", "autodetect", "Device serial")
	deployCmd.Flags().StringVarP(&partsize, "part-size", "S", "", "Linux partition size in percents")
}
