#!/bin/bash

mv /etc/iscsi/initiatorname.iscsi.original /etc/iscsi/initiatorname.iscsi

targetcli /iscsi delete iqn.0000-00.com.example:target0
rm -rf /iscsi_data