package cmd

import (
	"git.timoxa0.su/timoxa0/lon-tool/utils"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/timoxa0/gofastboot/fastboot"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Restore stock GPT",
	Long:  "Restore stock partition table",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fb_devs, err := fastboot.FindDevices()
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
					continue
				}
				logger.Debug("Found device", logger.Args("Product", product, "Sraial", devSerial))
				if product == "nabu" {
					logger.Debug("Nabu found", logger.Args("Serial", devSerial))
					serail = devSerial
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

		repartition, _ := pterm.DefaultInteractiveConfirm.Show("All user data will be ERASED. Continue?")
		if !repartition {
			logger.Info("Bye")
			os.Exit(253)
		}

		fb_dev, _ := fastboot.FindDevice(serail)

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
		logger.Info("Done! Reflash stock rom to remove uefi")
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().StringVarP(&serail, "serial", "s", "autodetect", "Device serial")
}
