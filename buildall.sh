#!/bin/sh
find . -type d -name goadif\* -print0 | xargs -0 -I % sh -c '(cd % && go build)'
