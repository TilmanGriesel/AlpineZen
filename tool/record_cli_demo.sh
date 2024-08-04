#!/bin/bash

OLD_PS1=$PS1
export PS1='$ '

echo "Recording terminal session..."
asciinema rec temp.cast

echo "Converting to SVG..."
cat temp.cast | svg-term --window --out cli_demo.svg

echo "Copying SVG to docs/demos..."
mkdir -p docs/assets/demos
cp cli_demo.svg docs/assets/demos

echo "Cleaning up..."
rm temp.cast cli_demo.svg

export PS1=$OLD_PS1

echo "Demo recording and processing complete!"
