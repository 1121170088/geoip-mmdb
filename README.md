# geoip-mmdb
~~数据库和程序点击Actions在artifact中下载~~

https://docs.github.com/en/actions/managing-workflow-runs/downloading-workflow-artifacts

这个artifact没登录没法下载，看来只能release了
# 命令行模式

geoip-mmdb -ip 1.1.1.1

# http server 模式  

geoip-mmdb -s  

curl http://127.0.0.1:9080/domains -v -d "[\"example.com\"]" 查域名解析  

curl http://127.0.0.1:9080/ip?1.1.1.1 查ip  

curl http://127.0.0.1:9080/ip  查自己  

