package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

const (
	TorSocksUrl  = "socks5://127.0.0.1:"
	TorSocksPort = 10000
	torCommand   = "tor"
	torFlag      = "-f"
	torConfig    = "/home/daniel/.tor/data-0/torrc-0"
	psCommand    = "sh"
	psArg        = "findPid.sh"
	killBase     = "kill"
	killArg      = "-9"
	catBase      = "cat"
)

func TorTransport(index int) (*http.Transport, error) {
	torUrl := TorSocksUrl + strconv.Itoa(TorSocksPort+10*index)
	tbProxyURL, err := url.Parse(torUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse proxy URL: %v", err)
	}

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain proxy dialer: %v", err)
	}

	torTransport := &http.Transport{Dial: tbDialer.Dial}
	return torTransport, nil
}

func RestartTorServer(ip string, index int) error {
	err := stopTorServer(index)
	if err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	err = initTorServer(index)
	if err != nil {
		return err
	}

	return nil
}

func command(commands ...string) ([]byte, error) {
	cmd := exec.Command(commands[0], commands[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return nil, fmt.Errorf("cannot execute command %s", err.Error())
	}
	return out, nil
}

func stopTorServer(index int) error {
	if psCmd, err := command(psCommand, psArg, strconv.Itoa(index)); err != nil {
		return errors.New("ps cmd: " + err.Error())
	} else {
		pidTorServer := strings.TrimSpace(string(psCmd))
		if _, err := command(killBase, killArg, pidTorServer); err != nil {
			return errors.New("kill cmd: " + err.Error())
		} else {
			return nil
		}
	}
}

func initTorServer(index int) error {
	torConfigIndex := strings.Replace(torConfig, "0", strconv.Itoa(index), -1)
	if _, err := command(torCommand, torFlag, torConfigIndex); err != nil {
		return errors.New("tor cmd: " + err.Error())
	} else {
		return nil
	}
}

func checkTorServer(index int) error {
	if psCmd, err := command(psCommand, psArg, strconv.Itoa(index)); err != nil {
		return errors.New("ps cmd: " + err.Error())
	} else {
		pidTorServer := strings.TrimSpace(string(psCmd))
		if pidTorServer == "" {
			if err := initTorServer(index); err != nil {
				return err
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
}

func banIpAddress(ip string) error {
	for index := 1; index <= cantTorServer; index++ {
		torConfigIndex := strings.Replace(torConfig, "0", strconv.Itoa(index), -1)
		if catCmd, err := command(catBase, torConfigIndex); err != nil {
			return err
		} else {
			torCfg := string(catCmd)
			if strings.Contains(torCfg, "ExcludeExitNodes") {
				torCfg = strings.Replace(torCfg, "\nStrictNodes", ","+ip+"/32\nStrictNodes", 1)
				if err := newTorConfig(index, torCfg); err != nil {
					return err
				}
			} else {
				f, err := os.OpenFile(torConfigIndex, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
				defer f.Close()

				banIp := fmt.Sprintf("ExcludeExitNodes %s/32\nStrictNodes 1\n", ip)
				if _, err := f.WriteString(banIp); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func newTorConfig(index int, text string) error {
	torConfigIndex := strings.Replace(torConfig, "0", strconv.Itoa(index), -1)
	file, err := os.OpenFile(torConfigIndex, os.O_RDWR, 0644)
	if err != nil {
		return errors.New("new tor config: " + err.Error())
	}
	defer file.Close()

	buffer := bytes.NewBufferString(text)
	_, err = file.WriteAt(buffer.Bytes(), 0)
	if err != nil {
		return errors.New("new tor config: " + err.Error())
	}

	return nil
}
