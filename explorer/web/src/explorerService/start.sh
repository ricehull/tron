rm appLog*.log
logName=`date +%Y%m%d%H%M%S`
killall explorerService > /dev/null 2>&1
sleep 1
nohup ./explorerService -debug true > "appLog$logName.log" 2>&1 &