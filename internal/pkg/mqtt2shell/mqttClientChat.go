package mqtt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

const prompt = ">"
const login = "-------------------------------------------------\r\n|  Mqtt-shell client \r\n|\r\n|  IP: %s \r\n|  SERVER VER: %s - CLIENT VER: %s\r\n|  TX: %s\r\n|  RX: %s\r\n|\r\n-------------------------------------------------\r\n"

type MqttClientChat struct {
	*MqttChat
	waitServerChan chan bool
	io             *ClientChatIO
}

func (m *MqttClientChat) print(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(m.io.Writer, a...)
}

func (m *MqttClientChat) println() (n int, err error) {
	return fmt.Fprintln(m.io.Writer)
}

func (m *MqttClientChat) printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(m.io.Writer, format, a...)
}

func (m *MqttClientChat) printWithoutLn(a ...interface{}) (n int, err error) {
	return fmt.Fprint(m.io.Writer, a...)
}

func (m *MqttClientChat) OnDataRx(data MqttJsonData) {

	if data.Uuid == "" || data.Cmd == "" || data.Data == "" {
		return
	}
	out := strings.TrimSuffix(data.Data, "\n") // remove newline
	m.print(out)
	m.println()
	m.printPrompt()
}

func (m *MqttClientChat) waitServerCb(data MqttJsonData) {

	if data.Uuid == "" || data.Cmd != "shell" || data.Data == "" {
		return
	}
	m.waitServerChan <- true
	ip := data.Ip
	serverVersion := data.Version
	m.printLogin(ip, serverVersion)
}

func (m *MqttClientChat) printPrompt() {
	m.printWithoutLn(prompt)
}

func (m *MqttClientChat) printLogin(ip string, serverVersion string) {
	log.Info("Connected")
	m.printf(login, ip, serverVersion, m.version, m.txTopic, m.rxTopic)
	m.printPrompt()
}

func (m *MqttClientChat) waitServer() {
	m.SetDataCallback(m.waitServerCb)
	for {
		log.Info("Connecting to server...")
		m.Transmit("whoami", "")
		select {
		case ok := <-m.waitServerChan:
			if ok {
				m.SetDataCallback(m.OnDataRx)
				return
			}
		case <-time.After(5 * time.Second):
			log.Info("TIMEOUT , retry...")
		}
	}

}

func (m *MqttClientChat) clientTask() {
	m.waitServer()
	for {
		scanner := bufio.NewScanner(m.io.Reader)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				m.printPrompt()
			} else {
				m.Transmit(line, "")
			}
		}
	}
}

type ClientChatIO struct {
	io.Reader
	io.Writer
}

func defaultIO() ClientChatIO {
	return struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
}

func NewClientChat(mqttOpts *MQTT.ClientOptions, rxTopic string, txTopic string, version string, customIO *ClientChatIO, opts ...MqttChatOption) *MqttClientChat {
	var actualIO *ClientChatIO

	if customIO == nil {
		defaultIO := defaultIO()
		actualIO = &defaultIO
	} else {
		actualIO = customIO
	}

	cc := MqttClientChat{io: actualIO}
	chat := NewChat(mqttOpts, rxTopic, txTopic, version, opts...)
	chat.SetDataCallback(cc.OnDataRx)
	cc.MqttChat = chat
	cc.waitServerChan = make(chan bool)
	go cc.clientTask()

	return &cc
}
