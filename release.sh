#! /bin/bash
################################################################################
##
## MIT License
##
## Copyright (c) 2017 Kevin Druelle
##
## Permission is hereby granted, free of charge, to any person obtaining a copy
## of this software and associated documentation files (the "Software"), to deal
## in the Software without restriction, including without limitation the rights
## to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
## copies of the Software, and to permit persons to whom the Software is
## furnished to do so, subject to the following conditions:
##
## The above copyright notice and this permission notice shall be included in all
## copies or substantial portions of the Software.
##
## THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
## IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
## FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
## AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
## LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
## OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
## SOFTWARE
##
###############################################################################

OS=$(uname)

for BIN in $(ls build); do
    BINNAME=$(echo ${BIN} | sed 's/\.bin//g')
    ARCH=$(echo ${BIN} | sed -E 's/.*-([a-z0-9]+)\.bin(\.exe)?/\1/g')
    ZIP=$(echo ${BIN} | sed -E 's/\.bin(\.exe)?/.zip/g')
    if [[ "${BIN}" =~ "windows" ]]; then
        cp build/${BIN} dist/sslcap.exe
        cd dist && zip ${ZIP} sslcap.exe && rm sslcap.exe && cd ../
    else
        cp build/${BIN} dist/sslcap
        cd dist && zip ${ZIP} sslcap && rm sslcap && cd ../
    fi
    if [[ "${BIN}" =~ "linux" ]]; then
        DEB=$(echo ${BIN} | sed -E 's/\.bin/.deb/g')
        RPM=$(echo ${BIN} | sed -E 's/\.bin/.rmp/g')
        if [[ "$OS" == "Darwin" ]]; then
            sed -i '' "s#.*/usr/local/bin/sslcap.*#    ./build/${BIN}: /usr/local/bin/sslcap#g" nfpm.yaml
            sed -i '' "s#.*arch.*#arch:  ${ARCH}#g" nfpm.yaml
        else
            sed -i "s#.*/usr/local/bin/sslcap.*#    ./build/${BIN}: /usr/local/bin/sslcap#g" nfpm.yaml
            sed -i "s#.*arch.*#arch:  ${ARCH}#g" nfpm.yaml
        fi
        nfpm pkg -t dist/${DEB}
    fi
done

