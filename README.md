# Roly-Poly VPN
## Why?
It is a daemon designed to ease 2FA VPN usage with one time passwords at GNU/Linux systems.

## How it works
At first start it takes your permanent password, TOTP secret and Network Manager config name and places them into your keyring.
Then you can start it as a daemon and it will bring up that VPN connection using password and TOTP generated passcode. Also it will deactivate VPN connection at termination.

## Systemd
Prepared to be run as a *per-user service* but one may also try to run it system-wide, so there is a `./systemd/net-up-adm.pkla` file to deal with audit settings.

## Installation
- Import your OpenVPN configuration to NetworkManager configuration.
- Download binary from [release page](https://github.com/nixargh/roly-poly-vpn/releases).
- Set execution bit for binary: ```chmod +x ./roly-poly-vpn```
- Move somewhere to your **PATH**. At Ubuntu I prefer `~/.local/bin/` directory: ```mv ./roly-poly-vpn ~/.local/bin/```.
- Run it and answer questions about NetworkManager VPN config name, your LDAP password and OTP secret.
If you make a mistake and want to change the value just run **roly-poly-vpn** with flag setting this secret and it will be overwritten at your keyring. Or as alternative **seahorse** utility, which is a GUI keyring manager, could be used.
- In case you like to run **roly-poly-vpn** as a **per-user systemd** service, do like this:  
```
cp ./systemd/roly-poly-vpn.service ~/.config/systemd/user/
systemctl --user enable roly-poly-vpn.service
systemctl --user start roly-poly-vpn.service
```

## Options
- `-instance` flag allows to have multiple sets of settings to control many VPN (and other) connections. Every set of settings is stored at keyring as JSON string inside its own named key. At Gnome Keyring it looks like `Password for 'default' on 'roly-poly-vpn'`.
