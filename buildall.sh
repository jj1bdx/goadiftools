#!/bin/sh
find . -type d \( -name goadif\* -o -name noasciitostar \) -print0 | xargs -0 -I % sh -c '(cd % && go build)'
