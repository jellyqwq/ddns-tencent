# ddns-tencent

## 用法
编辑 `docker-compose.yml` 环境变量然后运行
```sh
docker compose up -d
```

## 环境变量说明

|变量名|说明
|-|-|
|Ipv4Domain|ipv4要解析的域名, 若为空或无该字段则Ipv4的字段均不生效|
|Ipv4SubDomain|ipv4子域名|
|Ipv4RecordLine|某条线路名称的解析记录, 选填
|Ipv4RecordId|记录Id, 详见 <https://cloud.tencent.com/document/api/1427/56166>
|Ipv6Domain|ipv6要解析的域名, 若为空或无该字段则Ipv6的字段均不生效|
|Ipv6SubDomain|ipv6子域名|
|Ipv6RecordLine|某条线路名称的解析记录, 选填
|Ipv6RecordId|记录Id, 详见 <https://cloud.tencent.com/document/api/1427/56166>
|SecretId|详见 <https://console.dnspod.cn/account/token/apikey>|
|SecretKey|详见 <https://console.dnspod.cn/account/token/apikey>|
|CronExpression|定时任务表达式, 默认值为 `*/5 * * * *`, 选填

## 例子
只对ipv6的配置ddns, 该配置会将ipv6解析到 `ddns.test.example.com` 上 
```yaml
services:
  ddns:
    image: jellyqwq/ddns-tencent:latest
    container_name: ddns-tencent
    restart: unless-stopped
    network_mode: host
    environment:
      - Ipv6Domain=example.com
      - Ipv6SubDomain=ddns.test
      - Ipv6RecordId=xxxxxxxx
      - SecretId=xxxxxxxxxxxxxxxxxxx
      - SecretKey=xxxxxxxxxxxxxxxxxx
```

