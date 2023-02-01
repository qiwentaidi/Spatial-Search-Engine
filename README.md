# Spatial Search Engine

Spatial Search Engine 目前支持 FOFA 和 Hunter 的查询，后续已支持子域名暴破、端口扫描功能

## FOFA&Hunter

### 功能提醒

```
右下角左右箭头只支持中间页码加减功能，翻页后需要点击查询

Hunter 查询后右上角会显示当前剩余积分，积分=每日免费积分+权益积分
FOFA 查询后左上角会显示当前用户等级，右下角会显示查询数量
由于考虑到大家应该都买不起企业会员，这里就不提供蜜罐排查，FID字段结果查询功能等
```

### 查询结果

![image-20230111172354064](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230111172354064.png)

![image-20230111172407313](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230111172407313.png)

### 语法查询

![image-20230111172430226](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230111172430226.png)

![image-20230111172437462](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230111172437462.png)

### 结果导出

```
点击数据导出按钮，会在当前对应的result目录下生成assets_当前时间戳.csv文件
```

### 配置文件

```
启动工具会在当前目录下生成 config.yaml 文件，可以在界面中更改也可以在 config.yaml 中更改，API表示网址（末尾不加/）防止如 fofa变更网址的情况
```

![image-20230111172539039](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230111172539039.png)

## 子域名暴破

```
点击解析按钮会进行域名解析

导入字典目录，点击暴破会进行子域名暴破
```

![image-20230201100000765](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230201100000765.png)

![image-20230201100110824](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230201100110824.png)

![image-20230201100133698](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230201100133698.png)

## 端口扫描

![image-20230201095227693](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230201095227693.png)

![image-20230201100249879](https://qwtd-image.oss-cn-hangzhou.aliyuncs.com/img/image-20230201100249879.png)

## 打包

```
windows 打包命令
go build -ldflags -H=windowsgui main.go 
```

