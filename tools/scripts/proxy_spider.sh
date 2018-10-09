#!/bin/bash

# prepare
git clone https://github.com/zhangchenchen/proxyspider.git
cd proxyspider
pip install -r requirements.txt

# fetch proxys
python proxyspider.py

cat proxy_list.txt | cut -d " " -f1 | grep ":" | sed 's#^#http://#g'