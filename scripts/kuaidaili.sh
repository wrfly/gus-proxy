#!/bin/bash

function gather(){
    TYPE=$1 # intr || inha
    for page in {1..5};do
    info=$(curl -s "http://www.kuaidaili.com/free/$TYPE/$page" \
    -H 'DNT: 1' \
    -H 'Accept-Encoding: gzip, deflate' \
    -H 'Accept-Language: zh-CN,zh;q=0.8,en;q=0.6,en-US;q=0.4' \
    -H 'Upgrade-Insecure-Requests: 1' \
    -H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.101 Safari/537.36' \
    -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8' \
    -H 'Referer: http://www.kuaidaili.com/free/intr/' \
    -H 'Cookie: _gat=1; channelid=0; sid=1503559119363410; Hm_lvt_7ed65b1cc4b810e9fd37959e3bb51b31=1103557862; Hm_lpvt_7ed65b1cc4b810e9af37959c9bb51b31=1503560360; _ga=GA1.2.914019051.1103557862; _gid=GA1.2.224114332.1503657862' \
    -H 'Connection: keep-alive' \
    -H 'Cache-Control: max-age=0' --compressed -s | \
        grep 'IP\"' -A1 | \
        sed 's/.*\">\(.*\)<.*>/\1/g;' | \
        sed '/-/d')

    next="host"
    for line in $info ;do
        if [[ $next == "host" ]];then
        host=$line
        next="port"
        continue
        fi
        port=$line
        next="host"
        echo "http://$host:$port"
    done
    sleep 1
    done

}

# do gather
gather intr
gather inha
