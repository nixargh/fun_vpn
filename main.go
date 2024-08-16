package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	//	"github.com/pkg/profile"
)

var version string = "2.0.2"

var clog *log.Entry

func main() {
	//	defer profile.Start().Stop()

	var debug bool
	var showVersion bool

	var instance string

	var connection string
	var password string
	var otpSecret string

	// Common flags
	flag.BoolVar(&debug, "debug", false, "Log debug messages")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.StringVar(&instance, "instance", "default", "Configuration instance name to save config to.")

	// VPN flags
	flag.StringVar(&connection, "connection", "", "VPN connection name (use 'nmcli connection' to find out)")
	flag.StringVar(&password, "password", "", "VPN user password")
	flag.StringVar(&otpSecret, "otpSecret", "", "VPN OTP secret")

	flag.Parse()

	// Setup logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if debug == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	clog = log.WithFields(log.Fields{
		"pid":     os.Getpid(),
		"thread":  "main",
		"version": version,
	})

	clog.Info("Let's have some fun with 2FA VPN via NM!")

	var config Config
	config.read(instance)

	// Overriding or settings instance config parameters.
	clog.Debug("Configuring connection name.")
	if connection != "" {
		config.Connection = connection
	} else {
		if config.Connection == "" {
			clog.Info("Hint: Use 'nmcli connection' to find out your config names.")
			config.Connection = askValue("connection", false)
		}
	}

	clog.Debug("Configuring password.")
	if password != "" {
		config.Password = password
	} else {
		if config.Password == "" {
			config.Password = askValue("password", true)
		}
	}

	clog.Debug("Configuring OTP secret.")
	if otpSecret != "" {
		config.OtpSecret = connection
	} else {
		if config.OtpSecret == "" {
			config.OtpSecret = askValue("OTP secret", true)
		}
	}

	// Save currently built config
	config.write(instance)

	go waitForDeath(config.Connection)

	sleepSeconds := 5
	clog.WithFields(log.Fields{"sleepSeconds": sleepSeconds}).Info("Starting the main loop.")
	for {
		active := nmcliConnectionActive(config.Connection, false)
		if !active {
			// Check whether any network connection is active
			activeConns := nmcliGetActiveConnections(true)
			if len(activeConns) > 0 {
				clog.WithFields(log.Fields{
					"connection": config.Connection,
				}).Info("VPN connection isn't active. Starting.")

				if config.Password != "Null" && config.OtpSecret != "Null" {
					passcode := GeneratePassCode(config.OtpSecret)
					clog.WithFields(log.Fields{"passcode": passcode}).Info("Got a new pass code.")

					// Update VPN config to store password only for current user
					nmcliConnectionUpdatePasswordFlags(config.Connection, 1)

					nmcliConnectionUpdatePassword(config.Password, passcode, config.Connection)

					nmcliConnectionUp(config.Connection)

					// Update VPN config to ask password every time.
					// That should prevent NM reconections with an old password.
					nmcliConnectionUpdatePasswordFlags(config.Connection, 2)
				} else {
					nmcliConnectionUp(config.Connection)
				}

			} else {
				clog.Info("No active connection found, thus posponding VPN connection.")
			}
		}
		clog.WithFields(log.Fields{
			"connection":   config.Connection,
			"sleepSeconds": sleepSeconds,
		}).Debug("Connection is active. Sleeping.")

		// Sleep for a minute
		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}

func askValue(parameter string, hide bool) string {
	var parameterValue string

	fmt.Printf("New '%v' value: ", parameter)

	if hide {
		bytespw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
			clog.WithFields(log.Fields{
				"parameter": parameter,
				"error":     err,
			}).Fatal("Reading hidden parameter value from cmd failed.")
		}
		parameterValue = string(bytespw)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
			clog.WithFields(log.Fields{
				"parameter": parameter,
				"error":     err,
			}).Fatal("Reading parameter value from cmd failed.")
		}
		parameterValue = scanner.Text()
	}
	fmt.Print("\n")

	// To understand when we have an empty password and when we just haven't set it yet.
	if parameterValue == "" {
		parameterValue = "Null"
	}

	return parameterValue
}

func GeneratePassCode(secret string) string {
	passcode, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if err != nil {
		clog.Fatal("TOTP pass code generation failed.")
	}
	return passcode
}

func basher(command string, hide string) string {
	commandStr := command

	cmd, err := exec.Command("/bin/bash", "-c", command).Output()
	output := string(cmd)

	if hide != "" {
		commandStr = strings.Replace(commandStr, hide, "*****", -1)
	}

	clog.WithFields(log.Fields{"command": commandStr, "output": output}).Debug("Command output.")

	if err != nil {
		clog.WithFields(log.Fields{"command": commandStr, "error": err}).Fatal("Shell command failed.")
	}

	return output
}

func waitForDeath(connection string) {
	clog.Info("Starting Wait For Death loop.")
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	for {
		time.Sleep(time.Duration(1) * time.Second)

		sig := <-cancelChan
		clog.WithFields(log.Fields{"signal": sig}).Info("Caught signal. Terminating.")

		active := nmcliConnectionActive(connection, false)
		if active {
			nmcliConnectionDown(connection)
		}

		clog.WithFields(log.Fields{"signal": sig}).Info("We are good to go, see you next time!.")
		os.Exit(0)
	}
}
