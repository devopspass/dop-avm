# dop-avm

DOP Ansible Version Manager

## Why?

As part of DevOps Pass AI project we needed integration for Ansible, which will work on all platforms in the same way including Windows, MacOS and Linux.
As you probably know Ansible is not suported natively on Windows - https://docs.ansible.com/ansible/latest/os_guide/windows_faq.html#can-ansible-run-on-windows

Inspired by [tofuutils/tenv](https://github.com/tofuutils/tenv) been created DOP-AVM (DevOps Pass AI Ansible Version Manager).

It uses Docker under the hood and allowing you to run Ansible tools from your local without installation of Python and on Windows.

## Installation

### MacOS / Linux

```bash
# Linux
curl -sL $(curl -s https://api.github.com/repos/devopspass/dop-avm/releases/latest | grep "https.*linux_amd64" | awk '{print $2}' | sed 's/"//g') | tar xzvf - dop-avm
# MacOS
curl -sL $(curl -s https://api.github.com/repos/devopspass/dop-avm/releases/latest | grep "https.*darwin_amd64" | awk '{print $2}' | sed 's/"//g') | tar xzvf - dop-avm

sudo mv dop-avm /usr/local/bin/
sudo sh -c "cd /usr/local/bin/ && dop-avm setup"
```

### Windows

Download latest binary for Windows - https://github.com/devopspass/dop-avm/releases/

```cmd
tar xzf dop-avm*.tar.gz
md %USERPROFILE%\bin
move dop-avm.exe %USERPROFILE%\bin\
setx PATH "%USERPROFILE%\bin;%PATH%"

dop-avm setup
```
