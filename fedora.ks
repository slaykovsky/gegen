install

auth --enableshadow --passalgo=sha512
repo --name="main" --baseurl="http://mirror.yandex.ru/fedora/linux/releases/24/Server/x86_64/os/"

eula --agreed
text

reboot

url --url="http://mirror.yandex.ru/fedora/linux/releases/24/Server/x86_64/os/"

firstboot --enable
# ignoredisk --only-use=vda
keyboard us
lang en_US.UTF-8

network --bootproto=dhcp --device=ens3 --ipv6=auto --activate
network --hostname=fedora24
rootpw password
services --enabled=NetworkManager,sshd
timezone Europe/Prague

user --groups=wheel --homedir=/home/user --name=user --password=password --gecos="user"
bootloader --location=mbr --boot-drive=vda
autopart --type=btrfs
zerombr
clearpart --drives=vda --all --initlabel

firewall --enabled --ssh
selinux --permissive

%packages
@core
%end
