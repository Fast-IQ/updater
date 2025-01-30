package updater

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Updater struct {
	urlServer  string
	updateFile string
	updatePath string
	pathApp    string
	nameApp    string
	time       time.Time
}

func NewUpdater() *Updater {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	app := filepath.Base(os.Args[0])
	path := strings.Replace(ex, app, "", 1)
	return &Updater{
		pathApp:   path,
		nameApp:   app,
		urlServer: "",
	}
}

func (u *Updater) UpdateFile(url string, nameFile string) error {
	if nameFile == "" {
		u.updateFile = u.nameApp
	} else {
		u.updateFile = filepath.Base(nameFile)
	}

	if filepath.Dir(nameFile) == "." {
		u.updatePath = ""
	} else {
		u.updatePath = filepath.Dir(nameFile)
	}

	if url > "" {
		u.urlServer = url
	}

	err := u.delOldVersion()
	if err != nil {
		slog.Error("Can`t delete old version.", err)
	}

	needUpdate, err := u.compareTimeNeedUpdate()
	if err != nil {
		return err
	}
	if needUpdate {
		err = u.updateFromServer()
	}

	return nil
}

func (u *Updater) compareTimeNeedUpdate() (needUpdate bool, err error) {
	var fileTime time.Time
	pf := filepath.Join(u.pathApp, u.updatePath, u.updateFile)
	if _, err = os.Stat(pf); err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		//Other error
		return false, err
	} else {

		timeSpec, err := os.Stat(pf)
		if err != nil {
			return false, err
		}
		fileTime = timeSpec.ModTime() //	BirthTime()	ChangeTime() ModTime() AccessTime()

		serverTime, err := u.getIntTimeFromServer()
		fmt.Println("Server time:", serverTime)
		fmt.Println("Local time:", fileTime.Unix())
		u.time = time.Unix(serverTime, 0)
		return serverTime > fileTime.Unix(), nil
	}

}

func (u *Updater) getTimeFromServer() (time.Time, error) {
	timeSrv := time.Time{}
	var buf bytes.Buffer
	err := u.urlRequest("version", &buf)
	if err != nil {
		return timeSrv, err
	}

	bodyBytes, err := io.ReadAll(&buf)
	if err != nil {
		return time.Time{}, err
	}

	//loc, _ := time.LoadLocation("")
	timeSrv, err = time.ParseInLocation("02.01.2006 15:04:05", string(bodyBytes), time.Local)

	return timeSrv, err
}

func (u *Updater) getIntTimeFromServer() (int64, error) {
	iTimeSrv := int64(0)
	var buf bytes.Buffer
	err := u.urlRequest("version2", &buf)
	if err != nil {
		return iTimeSrv, err
	}
	bodyBytes, err := io.ReadAll(&buf)
	if err != nil {
		return iTimeSrv, err
	}

	i, err := strconv.Atoi(string(bodyBytes))
	return int64(i), err
}

func (u *Updater) updateFromServer() error {
	var buf bytes.Buffer
	err := u.urlRequest("download", &buf)
	if err != nil {
		return err
	}

	return u.replaceFile(&buf)
}

func (u *Updater) urlRequest(pathUrl string, buf *bytes.Buffer) error {
	reqUrl, err := url.JoinPath(u.urlServer, "update", pathUrl, u.updatePath, u.updateFile)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Get(reqUrl)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Response error: " + resp.Status)
	}
	_, err = buf.ReadFrom(resp.Body)

	return err
}

func (u *Updater) replaceFile(data io.Reader) error {
	pu := filepath.Join(u.pathApp, u.updatePath, u.updateFile)

	err := os.MkdirAll(filepath.Dir(pu), 0755)
	if err != nil {
		return err
	}
	//update
	if _, err := os.Stat(pu); err == nil {
		//file exist -> Rename File
		err = os.Rename(pu, pu+"_old")
		if err != nil {
			return err
		}
	}
	//file not exist or rename
	newFile, err := os.OpenFile(pu, os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	defer func() { _ = newFile.Close() }()

	_, err = io.Copy(newFile, data)
	if err != nil {
		return err
	}

	err = os.Chtimes(pu, u.time, u.time)
	if err != nil {
		return err
	}
	return nil
}

func (u *Updater) delOldVersion() error {
	pu := filepath.Join(u.pathApp, u.updatePath, u.updateFile)

	//delete old file before update
	if _, err := os.Stat(pu + "_old"); err == nil {
		//file exist
		err = os.Remove(pu + "_old")
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *Updater) Update(url string) error {
	return u.UpdateFile(url, u.nameApp)
}
