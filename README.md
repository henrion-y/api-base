# api-base
基础api包

db2struct.exe -H 127.0.0.1 --mysql_port=3306 -d electronic_sports -t t_club -u root --package models --struct TClub -p 123456 --gorm --guregu
 
要看下配置文件这块有没有什么好的办法， 既可以兼容 ini 这种本地文件， 又可以兼容配置中心， 在使用的时候， 还不需要预先定义结构体， 可以直接set、get操作配置。

