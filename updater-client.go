package updater_client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Updater struct {
	url        string
	updateFile string
	path       string
	appName    string
	time       time.Time
}

func NewUpdater() *Updater {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return &Updater{
		path:    filepath.Dir(ex),
		appName: filepath.Base(os.Args[0]),
		url:     "",
	}
}

func (u *Updater) UpdateFile(url string, nameFile string) error {
	if nameFile == "" {
		u.updateFile = u.appName
	} else {
		u.path = filepath.Dir(nameFile)
		u.updateFile = filepath.Base(nameFile)
	}

	if url > "" {
		u.url = url
	}

	needUpdate, err := u.compareTimeNeedUpdate()
	if err == nil && needUpdate {
		err = u.updateFromServer()
	}

	return nil
}

func (u *Updater) compareTimeNeedUpdate() (needUpdate bool, err error) {
	var fileTime time.Time
	pf := filepath.Join(u.path, u.updateFile)
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
	reqUrl, err := url2.JoinPath(u.url, "update", pathUrl, u.path, u.updateFile)
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
	pu := filepath.Join(u.path, u.updateFile)

	//delete old file before update
	if _, err := os.Stat(pu + "_old"); err == nil {
		//file exist
		err = os.Remove(pu + "_old")
		if err != nil {
			return err
		}
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

func (u *Updater) Update(url string) error {
	return u.UpdateFile(url, u.appName)
}
