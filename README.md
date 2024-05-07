# DOP Ansible Version Manager

<!-- ⚠️ This README has been generated from the file(s) "blueprint.md" ⚠️--><p align="center">
  <img src="https://static.wixstatic.com/media/09a6dd_eae6b87971dd4d14ba7792cdd237dd76~mv2.png" alt="Logo" width="300" height="auto" />
</p>
<p align="center">
		<a href="https://github.com/devopspass/dop-avm"><img alt="Release" src="https://img.shields.io/github/release/devopspass/dop-avm.svg" height="20"/></a>
<a href=""><img alt="Downloads" src="https://img.shields.io/github/downloads/devopspass/dop-avm/total" height="20"/></a>
<a href="https://medium.com/@devopspass/"><img alt="Medium" src="https://img.shields.io/badge/Medium-12100E?style=for-the-badge&logo=medium&logoColor=white" height="20"/></a>
<a href="https://dev.to/devopspass"><img alt="dev.to" src="https://img.shields.io/badge/dev.to-0A0A0A?style=for-the-badge&logo=devdotto&logoColor=white" height="20"/></a>
<a href="https://www.linkedin.com/company/devopspass-ai"><img alt="LinkedIn" src="https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white" height="20"/></a>
<a href="https://www.youtube.com/@DevOpsPassAI"><img alt="YouTube" src="https://img.shields.io/badge/YouTube-FF0000?style=for-the-badge&logo=youtube&logoColor=white" height="20"/></a>
<a href="https://twitter.com/devops_pass_ai"><img alt="Twitter" src="https://img.shields.io/badge/Twitter-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white" height="20"/></a>
	</p>

<p align="center">
  <b>Run Ansible anywhere, using Docker</b></br>
  <sub>Run Ansible on Windows, MacOS and Linux via Docker.<sub>
</p>

<br />



[![-----------------------------------------------------](https://raw.githubusercontent.com/andreasbm/readme/master/assets/lines/water.png)](#-join-community)

## 💬 Join community

Join our Slack community, ask questions, contribute, get help!

[<img src="https://cloudberrydb.org/assets/images/slack_button-7610f9c51d82009ad912aded124c2d88.svg" width="150">](https://join.slack.com/t/devops-pass-ai/shared_invite/zt-2gyn62v9f-5ORKktUINe43qJx7HtKFcw)


## 🧐 Why?

As part of [DevOps Pass AI](https://github.com/devopspass/devopspass/) project we needed integration for Ansible, which will work on all platforms in the same way, including Windows, MacOS and Linux.
As you probably know Ansible is not suported natively on Windows - https://docs.ansible.com/ansible/latest/os_guide/windows_faq.html#can-ansible-run-on-windows

Inspired by [tofuutils/tenv](https://github.com/tofuutils/tenv) been created DOP-AVM (DevOps Pass AI Ansible Version Manager).

It uses Docker under the hood and allowing you to run Ansible tools from your local without installation of Python and on Windows.

## 🚀 Installation

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

cd %USERPROFILE%\bin\
dop-avm setup
```

### DevOps Pass AI

In DOP you can add **Ansible** app and run action **Install Ansible Version Manager**, it will download and install `dop-avm`.

## 🤔 How it works?

`dop-avm` copying own binary with different names, which will be used later by user:

* ansible
* ansible-playbook
* ansible-galaxy
* ansible-vault
* ansible-doc
* ansible-config
* ansible-console
* ansible-inventory
* ansible-adhoc
* ansible-lint
* molecule

When you're running any of this command, it will run Docker container `devopspass/ansible:latest` and binary inside (source Dockerfile in repo).

AVM will pass environment variables from host machine:

* `ANSIBLE_*`
* `MOLECULE_*`
* `GALAXY_*`
* `AWS_*`
* `GOOGLE_APPLICATION_CREDENTIALS`

Plus volumes (if exist):

* `.ssh`
* `.aws`
* `.azure`
* `.ansible`

And services, like SSH-agent and Docker socket.
As a result you can run Ansible on Windows, MacOS and Linux via Docker without installation of Python and Ansible on your local, especially it's useful for Windows where it's not possible to run Ansible at all.

## 🐳 Use another Docker container

Probably you may have Docker container built in your organization, which is used in pipelines or recommended for local, you can use it by specifying `DOP_AVM_IMAGE_NAME` environment variable. Be sure that all necessary binaries, like `molecule` are inside, when you're running it.
