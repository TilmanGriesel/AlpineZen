#!/bin/sh

# Start CLI in the background
./cli --loglevel 3 --clock-disable --runtime-headless "$@" &

# Create config
cat <<EOF > /tmp/Caddyfile
:80 {
    root * /srv/latest_wallpaper
    file_server
    rewrite / /latest.jpg
}
EOF

# Serve
caddy run --config /tmp/Caddyfile
