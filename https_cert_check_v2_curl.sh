#!/bin/bash
set -x
################################################
# Create Date: 2023-09-26
# Author:      Mario
# Mail:        Mario@xxxx.com
# Version:     2.1
# Attention:   通过域名获取证书的过期时间
################################################
# curl 自带超时控制参数 分为连接超时和数据传输超时
# 但是并发还是控制需借助shell里的匿名管道
# 这也是下一步为什么用golang重写脚本的原因
################################################


# 脚本所在目录即脚本名称
script_dir=$( cd "$( dirname "$0" )" && pwd )
script_name=$(basename ${0})
script_log="${script_dir}/domains_tls_check_$(date '+%Y%m%d-%H%M%S-%N').log"

domains_file="${script_dir}/hosts.txt"
domains_check_result="${script_dir}/domains_check_result_$(date '+%Y%m%d-%H%M%S-%N').csv"
# 在shell中通过printf加入BOM头 解决中文乱码问题
printf "\xEF\xBB\xBF" >> ${domains_check_result}


# 将标准输出保存到FD3 标准错误保存到FD4 然后重定向他们到日志文件 追加的形式
exec 3>&1 4>&2 >> ${script_log} 2>&1


# 读取需要监测的域名文件
# 过滤空行且过滤注释行
grep -v "^$" ${domains_file} | grep -v "^#" | while read line; do
	echo "当前域名 ${line}"
	# 使用 curl 获取域名的证书情况 然后获取其中的过期时间
	# 命令行太长记得换行
	# --connect-timeout 连接超时
	# -m 数据传输最大允许时间
	time_raw=$( curl --connect-timeout 3 -m 3 -vIs https://${line} 2>&1 \
	| grep "expire date" | awk -F 'expire date: ' '{print $2}' )
	echo "证书过期时间 ${time_raw}"
	# 如果没有获取到证书过期时间 比如超时 
	# 那么记录本次测试结果并跳过这个域名
	# 证书时间格式里有空格 所以一定要用双引号引起来
	if [ -z "${time_raw}" ]; then
		echo "获取证书错误"
		echo "${line},err_get_certficate" >> ${domains_check_result}
		# 跳出本次循环
		continue
	fi
	# 将日期转化为时间戳
	time_expired=$(date +%s -d "$time_raw")
	echo "过期时间戳 ${time_expired}"
	# 将目前的日期也转化为时间戳
	# date -u '+%b %d %T %Y GMT' 是证书时间默认的格式
	# time_current=$(date -d "$(date -u '+%b %d %T %Y GMT')" +%s)
	time_current=$(date +%s)
	echo "当前时间戳 ${time_current}"
	# 到期时间减去目前时间再转化为天数 注意全部是整数运算
	days_left=$(( $(($time_expired-$time_current))/((60*60*24)) ))
	echo "域名证书过期剩余天数"
	echo "${line},${days_left}" >> ${domains_check_result}
done

# 恢复标准输出和标准错误
exec 1>&3 2>&4
