package upload

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Config struct {
	SaveDir   string `mapstructure:"save_dir"`
	UrlPrefix string `mapstructure:"url_prefix"`
	Host      string `mapstructure:"host"`
}

func (c Config) New() (*Uploader, error) {
	if c.SaveDir == "" {
		return nil, fmt.Errorf("empty save dir")
	}

	u := Uploader{
		conf: c,
	}
	os.MkdirAll(c.SaveDir, 0777)
	return &u, nil
}

type Uploader struct {
	conf Config
}

func (u *Uploader) GetFile(uri string) (file *os.File, err error) {
	dst, err := u.GetFilePath(uri)
	if err != nil {
		return nil, err
	}

	//out, err := os.Create(dst)
	file, err = os.Open(dst)
	return
}

func (u *Uploader) GetFilePath(uri string) (fpath string, err error) {
	filename := strings.TrimLeft(uri, u.conf.UrlPrefix)
	dst := path.Join(u.conf.SaveDir, filename)
	//out, err := os.Create(dst)
	return dst, nil
}

func (u *Uploader) RemoveFile(uri string) (err error) {
	filename := strings.TrimLeft(uri, u.conf.UrlPrefix)
	dst := path.Join(u.conf.SaveDir, filename)
	//out, err := os.Create(dst)
	return os.Remove(dst)
}

func (u *Uploader) Exist(filename string) (ok bool, err error) {
	dst := path.Join(u.conf.SaveDir, filename)
	_, err = os.Stat(dst)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (u *Uploader) SaveFile(src io.Reader, filename string) (uri string, err error) {
	dst := path.Join(u.conf.SaveDir, filename)
	dir, _ := filepath.Split(dst)

	err = os.MkdirAll(dir, 0775)
	if err != nil {
		fmt.Println("mkdir err:", err, "dir:", dir)
		return "", err
	}
	//out, err := os.Create(dst)
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	uri = u.conf.UrlPrefix + filename
	return uri, err
}

func (u *Uploader) SaveGinUploadFile(file *multipart.FileHeader, filename string) (uri string, err error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	return u.SaveFile(src, filename)
}

func (u *Uploader) FillUrl(url string) string {
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return u.conf.Host + url
	//return filepath.Join(u.conf.Host, url)
}
