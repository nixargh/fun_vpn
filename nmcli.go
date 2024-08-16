package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func nmcliGetActiveConnections(physical bool) []string {
	output := basher("nmcli -f NAME,TYPE,STATE -t connection show --active", "")
	var connections, filteredConnections []string

	if len(output) > 0 {
		connections = strings.Split(output, "\n")
	}
	clog.WithFields(log.Fields{"raw_connections": connections}).Debug("All actived/activating connections found.")

	for i := 0; i < len(connections)-1; i++ {
		splitCon := strings.Split(connections[i], ":")
		name := splitCon[0]
		cType := splitCon[1]
		state := splitCon[2]

		if state != "activated" {
			continue
		}

		isPhysical := false
		if strings.HasSuffix(cType, "ethernet") || strings.HasSuffix(cType, "wireless") {
			isPhysical = true
		}

		clog.WithFields(log.Fields{"name": name, "type": cType, "isPhysical": isPhysical}).Debug("Connections name and type.")

		if physical && !isPhysical {
			continue
		}
		filteredConnections = append(filteredConnections, name)
	}

	clog.WithFields(log.Fields{"connections": filteredConnections, "physical": physical}).Debug("Filtered active connections found.")
	return filteredConnections
}

func nmcliConnectionActive(connection string, physical bool) bool {
	connections := nmcliGetActiveConnections(physical)
	index := slices.Index(connections, connection)

	if index == -1 {
		return false
	} else {
		return true
	}
}

func nmcliConnectionUpPasswd(password string, passcode string, connection string) {
	clog.WithFields(log.Fields{"connection": connection}).Info("Starting VPN connection.")

	passwdFile := "/tmp/roly-poly-vpn.nmcli.passwd"
	fullPassword := fmt.Sprintf("vpn.secrets.password:\"%v%v\"", password, passcode)

	err := os.WriteFile(passwdFile, []byte(fullPassword), 0600)
	if err != nil {
		clog.WithFields(log.Fields{"file": passwdFile, "error": err}).Fatal("Can't create temporary passwd file for nmcli.")
	}

	cmd := fmt.Sprintf("nmcli connection up %v passwd-file %v", connection, passwdFile)
	basher(cmd, password)
	clog.WithFields(log.Fields{"connection": connection}).Info("VPN is connected.")

	os.Remove(passwdFile)
}

func nmcliConnectionUpAsk(password string, passcode string, connection string) {
	var cmd string

	clog.WithFields(log.Fields{"connection": connection}).Info("Starting VPN connection.")

	// Update VPN config to ask password every time
	cmd = fmt.Sprintf("nmcli connection mod %v vpn.secrets 'password-flags=2'", connection)
	basher(cmd, "")

	// Answer to password request interactively
	fullpass := fmt.Sprintf("\"%v%v\"", password, passcode)
	cmd = fmt.Sprintf("nmcli connection mod %v vpn.secrets password=%v", connection, fullpass)
	basher(cmd, fullpass)

	clog.WithFields(log.Fields{"connection": connection}).Info("VPN is connected.")
}

func nmcliConnectionUp(connection string) {
	clog.WithFields(log.Fields{"connection": connection}).Info("Starting VPN connection.")

	cmd := fmt.Sprintf("nmcli connection up %v", connection)
	basher(cmd, "")
	clog.WithFields(log.Fields{"connection": connection}).Info("VPN is connected.")
}

func nmcliConnectionUpdatePasswordFlags(connection string, value int) {
	var cmd string

	clog.WithFields(log.Fields{
		"connection":     connection,
		"password-flags": value,
	}).Debug("Updating VPN connection with a new password-flags.")

	cmd = fmt.Sprintf("nmcli connection mod %v +vpn.data 'password-flags=%d'", connection, value)
	basher(cmd, "")

	clog.WithFields(log.Fields{"connection": connection}).Debug("VPN password-flags is updated.")
}

func nmcliConnectionUpdatePassword(password string, passcode string, connection string) {
	var cmd string

	clog.WithFields(log.Fields{"connection": connection}).Info("Updating VPN connection with a new password.")

	// Update VPN config with a newly generated password
	fullpass := fmt.Sprintf("\"%v%v\"", password, passcode)
	cmd = fmt.Sprintf("nmcli connection mod %v vpn.secrets password=%v", connection, fullpass)
	basher(cmd, fullpass)

	clog.WithFields(log.Fields{"connection": connection}).Info("VPN connection is updated.")
}

func nmcliConnectionDown(connection string) {
	clog.WithFields(log.Fields{"connection": connection}).Info("Stopping VPN connection.")
	cmd := fmt.Sprintf("nmcli connection down %v", connection)
	basher(cmd, "")
}
