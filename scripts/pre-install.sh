#!/bin/bash

if [[ -d /etc/opt/telegraf ]]; then
    # Legacy configuration found
    if [[ ! -d /etc/telegraf ]]; then
        # New configuration does not exist, move legacy configuration to new location
        echo -e "Please note, Telegraf's configuration is now located at '/etc/telegraf' (previously '/etc/opt/telegraf')."
        mv -vn /etc/opt/telegraf /etc/telegraf

        if [[ -f /etc/telegraf/telex.conf ]]; then
            backup_name="telex.conf.$(date +%s).backup"
            echo "A backup of your current configuration can be found at: /etc/telegraf/${backup_name}"
            cp -a "/etc/telegraf/telex.conf" "/etc/telegraf/${backup_name}"
        fi
    fi
fi
