docker network create net \
--driver bridge \
--subnet 172.20.0.0/16 \
--gateway 172.20.0.1

docker run -d -p 44457:44457 \
--name database \
--net net \
--ip 172.20.0.4 \
teamrandint/database 

docker run -d -p 44455:44455 \
--name audit \
--net net \
--ip 172.20.0.3 \
teamrandint/auditserver 

docker run -d -p 44456:44456 \
--name web \
--net net \
--ip 172.20.0.5 \
teamrandint/webserver

docker run -d -p 44459:44459 \
--add-host quoteserve.seng:172.20.0.1 \
--name quote \
--net net \
--ip 172.20.0.7 \
teamrandint/quoteserver

docker run -d -p 44458:44458 \
--name trans \
--net net \
--ip 172.20.0.6 \
teamrandint/transactionserver

docker run -d -p 44460:44460 \
--name trigger \
--net net \
--ip 172.20.0.8 \
teamrandint/triggerserver
