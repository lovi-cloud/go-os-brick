#!/bin/bash

apt update -y -qq
apt install -y targetcli-fb dbus kmod

# enable d-bus daemon
mkdir /run/dbus
dbus-daemon --system

# create LUN
mkdir /iscsi_data

targetcli /iscsi delete iqn.0000-00.com.example:target0   # what is immutable????

targetcli /backstores/fileio create disk01 /iscsi_data/disk01 1G
targetcli /iscsi create iqn.0000-00.com.example:target0
targetcli /iscsi/iqn.0000-00.com.example:target0/tpg1/luns create /backstores/fileio/disk01
targetcli /iscsi/iqn.0000-00.com.example:target0/tpg1/acls create iqn.0000-00.com.example:initiator0
targetcli /iscsi/iqn.0000-00.com.example:target0/tpg1 set attribute authentication=0

targetcli saveconfig