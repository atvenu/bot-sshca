#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# chown as root
chown -R keybase:keybase /mnt

# Run everything else as the keybase user
sudo -i -u keybase bash << EOF
source docker/env.sh
nohup bash -c "run_keybase -g 2>&1 | grep -v 'KBFS failed to FUSE mount' &"
sleep 3
keybase oneshot --username \$KEYBASE_USERNAME --paperkey "\$KEYBASE_PAPERKEY"
bin/keybaseca service
EOF