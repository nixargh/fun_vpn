# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [2.0.2] - 2024-08-17
### Fixed
- `nmcli.go` remove connections with state **activating** from list of active connections. It was iterrapting **roly-poly-vpn** during connection to some new Wi-Fi point because connection was waiting for a password.

## [2.0.0] - 2024-07-31
The format of settings at Keyring changed. So please remove old Keyring keys containg `roly-poly-vpn` at their names. After start you will be asked for new settings.
To remove old Keyrings records I prefere to use `seahorse` utillity.

### Removed
- `main.go` recently added **-noSecrets** flag.
- `main.go` **-config** flag in favor of **-connection** flag.

### Added
- `main.go` **-connection** flag as a replacement for **-config** flag.
- `config.go` that contains **Config** type and its methods.

### Changed
- The logic of working with configuration. Now it is possible to store many sets of configurations at Keyring as JSON string using separate `instance` identifiers (**-instance** flag). This makes old configs obsolete.

## [1.4.0] - 2024-07-25
### Added
- `main.go` new flag **-noSecrets** that allows to control any connection without changing password (and anything else).

### Fixed
- `main.go` check whether `config` connection is active before try to get it down on termination.
- `nmcli.go` change **nmcliGetActiveConnections** function logic to list all connections or only physical (wifi, ethernet). This should prevent periodic failures because of some tunnel etc.

## [1.3.2] - 2024-01-24
### Fixed
- `nmcli.go` issue when password contains a single quote sign.

## [1.3.1] - 2024-01-24
### Fixed
- `nmcli.go` fix the way NM VPN config is being updated.
- `main.go` use full timestamp even when TTY is attached.

## [1.3.0] - 2024-01-21
### Added
- `nmcli.go` new function **nmcliConnectionUpdatePassword** that updates NM VPN config with a generated password for only current user.
- `nmcli.go` new function **nmcliConnectionUp** that only brings up a NM VPN connection with previously set password.

### Changed
- `main.go` use **nmcliConnectionUpdatePassword** + **nmcliConnectionUp** to bring up VPN connection.
- `systemd/roly-poly-vpn.service` update to run as per user service.
- `README.md` update **Installation** section.

## [1.2.0] - 2024-01-18
### Changed
- `main.go` rewrite secret at keyring if set by flag.
- `README.md` describe installation process.

## [1.1.1] - 2023-07-27
### Fixed
- `nmcli.go` **nmcliConnectionActive** search of element in slice.

## [1.1.0] - 2023-07-26
### Added
- Filter out **bridge** type of NetworkManager connection during check of active connection before VPN activation.

## [1.0.1] - 2023-07-07
### Added
- First release.
