package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/codingsince1985/checksum"
	"github.com/pterm/pterm"
)

type file struct {
	Name string
	Url  string
}

type hashesInside struct {
	Hashes hashes
}
type hashes struct {
	Md5    string
	Sha1   string
	Sha256 string
}

var Files = struct {
	OrangeFox     file
	CleanUserdata file
	GPT           file
	UEFIBootshim  file
	UEFIPayload   file
}{
	OrangeFox: file{
		Name: "ofox.img",
		Url:  "https://timoxa0.su/share/nabu/deployer/ofox.img",
	},
	CleanUserdata: file{
		Name: "userdata.img",
		Url:  "https://timoxa0.su/share/nabu/deployer/userdata.img",
	},
	GPT: file{
		Name: "gpt_both0.bin",
		Url:  "https://timoxa0.su/share/nabu/deployer/gpt_both0.bin",
	},
	UEFIBootshim: file{
		Name: "BootShim.Dualboot.bin",
		Url:  "https://timoxa0.su/share/nabu/deployer/uefi/BootShim.Dualboot.bin",
	},
	UEFIPayload: file{
		Name: "nabu_UEFI.fd",
		Url:  "https://timoxa0.su/share/nabu/deployer/uefi/nabu_UEFI.fd",
	},
}

func (f *file) Download(pbar pterm.ProgressbarPrinter) ([]byte, error) {
	var ret []byte = nil
	pwd, _ := os.Getwd()
	resp, err := http.Get(f.Url)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	file, err := os.Create(path.Join(pwd, "/files/", f.Name))
	if err != nil {
		return ret, err
	}
	defer file.Close()
	bar, _ := pbar.WithTotal(int(resp.ContentLength)).Start()
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return ret, nil
		}
		if n == 0 {
			break
		}
		bar.Add(n)
		file.Write(buf[:n])
		ret = append(ret, buf[:n]...)
	}
	return ret, nil
}

func (f *file) Get(pbar pterm.ProgressbarPrinter) ([]byte, error) {
	var ret []byte = nil
	pwd, _ := os.Getwd()
	if _, err := os.Stat(path.Join(pwd, "/files/")); os.IsNotExist(err) {
		os.Mkdir(path.Join(pwd, "/files/"), 0755)
	}

	if _, err := os.Stat(path.Join(pwd, "/files/", f.Name)); os.IsNotExist(err) {
		return f.Download(pbar)
	}

	parsedUrl, _ := url.Parse(f.Url)
	baseURL := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	resp, err := http.Get(fmt.Sprintf("%s/?info=%s", baseURL, parsedUrl.Path))
	if err != nil {
		e := err
		file, err := os.Open(path.Join(pwd, "/files/", f.Name))
		if err != nil {
			return ret, err
		}
		ret, _ := io.ReadAll(file)
		return ret, e
	}
	defer resp.Body.Close()
	var hashesI hashesInside
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, err
	}
	json.Unmarshal(body, &hashesI)

	if c, _ := checksum.MD5sum(path.Join(pwd, "/files/", f.Name)); c != hashesI.Hashes.Md5 {
		return f.Download(pbar)
	} else {
		file, err := os.Open(path.Join(pwd, "/files/", f.Name))
		if err != nil {
			return ret, err
		}
		return io.ReadAll(file)
	}
}
