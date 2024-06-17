package image

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/codingsince1985/checksum"
)

type ImageReader struct {
	Path      string
	Stat      fs.FileInfo
	Name      string
	Version   string
	CheckSum  string
	ImgSize   uint64
	Reader    *gzip.Reader
	HeaderLen uint16
}

type ImageWriter struct {
	Path         string
	RawImagePath string
	Name         string
	Version      string
	CheckSum     string
	ImgSize      uint64
	Writer       *gzip.Writer
	HeaderLen    uint16
	file         *os.File
}

func CreateImage(path string, rawImgPath string, name string, version string) (ImageWriter, func() error, error) {
	var image ImageWriter
	stat, err := os.Stat(rawImgPath)
	if err != nil {
		return image, func() error { return nil }, err
	}
	rawImgPath, _ = filepath.Abs(rawImgPath)
	file, err := os.Create(path)
	if err != nil {
		return image, func() error { return nil }, err
	}
	path, _ = filepath.Abs(path)

	checkSum, err := checksum.MD5sum(rawImgPath)
	if err != nil {
		return image, func() error { return nil }, err
	}
	headerLen := uint16(len(name)+len(version)) + 36
	gzWriter := gzip.NewWriter(file)

	return ImageWriter{
		Path:         path,
		RawImagePath: rawImgPath,
		Name:         strings.TrimSuffix(name, "\n"),
		Version:      strings.TrimSuffix(version, "\n"),
		CheckSum:     strings.TrimSuffix(checkSum, "\n"),
		ImgSize:      uint64(stat.Size()),
		Writer:       gzWriter,
		HeaderLen:    headerLen,
		file:         file,
	}, file.Close, nil
}

func (i *ImageWriter) WriteMetadata() {
	i.file.Seek(0, io.SeekEnd)
	writer := bufio.NewWriter(i.file)

	writer.WriteString(fmt.Sprintf("%s\n", strings.ReplaceAll(i.Name, "\n", "")))

	writer.WriteString(fmt.Sprintf("%s\n", strings.ReplaceAll(i.Version, "\n", "")))

	checkSumBytes, _ := hex.DecodeString(i.CheckSum)
	writer.Write(checkSumBytes)

	imgSizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(imgSizeBytes, uint64(i.ImgSize))
	writer.Write(imgSizeBytes)

	headerLenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(headerLenBytes, i.HeaderLen)
	writer.Write(headerLenBytes)

	writer.Write([]byte{0x4C, 0x4F, 0x4E, 0x49, 0x4D, 0x41, 0x47, 0x45})
	writer.Flush()
}

func ReadImage(path string) (ImageReader, func() error, error) {
	var image ImageReader
	stat, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return image, func() error { return nil }, err
	}
	path, _ = filepath.Abs(path)
	file, err := os.Open(path)
	if err != nil {
		return image, func() error { return nil }, err
	}

	file.Seek(-8, io.SeekEnd)

	signatureBytes := make([]byte, 8)
	file.Read(signatureBytes)
	if string(signatureBytes) != "LONIMAGE" {
		return image, func() error { return nil }, errors.New("not a LoN Image")
	}

	file.Seek(-10, io.SeekEnd)
	offsetBytes := make([]byte, 2)
	file.Read(offsetBytes)
	headerLen := binary.LittleEndian.Uint16(offsetBytes)
	file.Seek(-int64(headerLen), io.SeekEnd)

	reader := bufio.NewReader(file)

	name, err := reader.ReadString('\n')
	if err != nil {
		return image, func() error { return nil }, err
	}

	version, err := reader.ReadString('\n')
	if err != nil {
		return image, func() error { return nil }, err
	}

	checkSumBytes := make([]byte, 16)
	_, err = reader.Read(checkSumBytes)
	checkSum := hex.EncodeToString(checkSumBytes)
	if err != nil {
		return image, func() error { return nil }, err
	}

	imgSizeBytes := make([]byte, 8)
	_, err = reader.Read(imgSizeBytes)
	if err != nil {
		return image, func() error { return nil }, err
	}
	imgSize := binary.LittleEndian.Uint64(imgSizeBytes)

	file.Seek(0, io.SeekStart)
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return image, func() error { return nil }, err
	}

	return ImageReader{
		Path:      path,
		Stat:      stat,
		Name:      strings.TrimSuffix(name, "\n"),
		Version:   strings.TrimSuffix(version, "\n"),
		CheckSum:  strings.TrimSuffix(checkSum, "\n"),
		ImgSize:   imgSize,
		Reader:    gzReader,
		HeaderLen: headerLen,
	}, file.Close, nil
}
