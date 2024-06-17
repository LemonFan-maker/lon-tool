package cmd

import (
	"io"
	"os"

	"github.com/codingsince1985/checksum"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/timoxa0/lon_image/image"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Linux on Nabu image tool",
	Long:  "A tool for packing and unpacking linux images",
}

var imageUnpack = &cobra.Command{
	Use:   "extract <file.lni> <rootfs.img>",
	Short: "Extract rootfs image",
	Long:  "Extract rootfs image from lni file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		printInfo(cmd, args)

		image, close, err := image.ReadImage(args[0])
		defer close()
		if err != nil {
			logger.Error("Error reading image info: %v", logger.Args(err))
			os.Exit(1)
		}

		buf := make([]byte, 16777216)
		rawImage, _ := os.Create(args[1])

		bar, _ := pbar.WithTotal(int(image.ImgSize)).WithTitle("Extracting").Start()
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
			rawImage.Write(buf[:n])
		}
		rawImage.Close()
		checkSumSpinner, _ := spinner.Start("Calculating checksum")
		checkSum, _ := checksum.MD5sum(args[1])
		checkSumSpinner.Success()
		var msg string
		if checkSum == image.CheckSum {
			msg, _ = pterm.DefaultTree.WithRoot(pterm.TreeNode{
				Text: pterm.Green("Image extracted"),
				Children: []pterm.TreeNode{
					{Text: pterm.Sprintf("Path: %s", args[1])},
				},
			}).Srender()
		} else {
			msg, _ = pterm.DefaultTree.WithRoot(pterm.TreeNode{
				Text: pterm.Yellow("Invalid MD5 checksum. Image corrupted!"),
				Children: []pterm.TreeNode{
					{Text: pterm.Sprintf("Path: %s", args[1])},
				},
			}).Srender()
		}
		pterm.Print(msg)
	},
}

var imageInfo = &cobra.Command{
	Use:   "info <file.lni>",
	Short: "Get lni file info",
	Long:  "Get lni file info",
	Args:  cobra.ExactArgs(1),
	Run:   printInfo,
}

var imageName string
var imageVers string
var imagePack = &cobra.Command{
	Use:   "create <rootfs.img> <new.lni>",
	Short: "Create lni file",
	Long:  "Create lni file from rootfs image",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		checkSumSpinner, _ := spinner.Start("Calculating checksum")
		image, close, err := image.CreateImage(args[1], args[0], imageName, imageVers)
		defer close()
		defer image.WriteMetadata()
		defer image.Writer.Close()
		if err != nil {
			logger.Fatal("Error reading creating info: %v", logger.Args(err))
		}
		checkSumSpinner.Success()
		imageInfo, _ := pterm.DefaultTree.WithRoot(pterm.TreeNode{
			Text: pterm.Green("Image info"),
			Children: []pterm.TreeNode{
				{Text: pterm.Sprintf("Name: %s", image.Name)},
				{Text: pterm.Sprintf("Version: %s", image.Version)},
				{Text: pterm.Sprintf("Checksum: %s", image.CheckSum)},
				{Text: pterm.Sprintf("Size: %vMB", image.ImgSize/1024/1024)},
			},
		}).Srender()
		pterm.Print(imageInfo)
		bar, _ := pbar.WithTotal(int(image.ImgSize)).WithTitle("Creating").Start()
		buf := make([]byte, 16777216)
		rawImage, err := os.Open(image.RawImagePath)
		if err != nil {
			logger.Fatal("Failed to read image", logger.Args("error", err))
		}
		for {
			n, err := rawImage.Read(buf)
			if err != nil && err != io.EOF {
				logger.Error("Failed to read from image reader", logger.Args("error", err))
				return
			}
			if n == 0 {
				break
			}
			bar.Add(n)
			image.Writer.Write(buf[:n])
		}
		msg, _ := pterm.DefaultTree.WithRoot(pterm.TreeNode{
			Text: pterm.Green("Image created"),
			Children: []pterm.TreeNode{
				{Text: pterm.Sprintf("Path: %s", args[1])},
			},
		}).Srender()
		pterm.Print(msg)
	},
}

func printInfo(cmd *cobra.Command, args []string) {
	image, close, err := image.ReadImage(args[0])
	defer close()
	if err != nil {
		logger.Fatal("Failed to read image info", logger.Args("error", err))
	}
	imageInfo, _ := pterm.DefaultTree.WithRoot(pterm.TreeNode{
		Text: pterm.Green("Image info"),
		Children: []pterm.TreeNode{
			{Text: pterm.Sprintf("Name: %s", image.Name)},
			{Text: pterm.Sprintf("Version: %s", image.Version)},
			{Text: pterm.Sprintf("Checksum: %s", image.CheckSum)},
			{Text: pterm.Sprintf("Size: %vMB", image.ImgSize/1024/1024)},
		},
	}).Srender()
	pterm.Print(imageInfo)
}

func init() {
	rootCmd.AddCommand(imageCmd)
	imageCmd.AddCommand(imageUnpack, imagePack, imageInfo)
	imagePack.Flags().StringVarP(&imageName, "name", "n", "Generic Image", "Rootfs image name")
	imagePack.Flags().StringVarP(&imageVers, "version", "v", "testing", "Rootfs image version")
}
