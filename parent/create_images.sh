source ./.env

cd ../auditserver
docker image build \
--build-arg auditaddr=${auditaddr} \
--build-arg auditport=${auditport} \
-t teamrandint/auditserver . 

cd ../transaction-server
docker image build \
--build-arg transaddr=${transaddr} \
--build-arg transport=${transport} \
--build-arg dbaddr=${dbaddr} \
--build-arg dbport=${dbport} \
--build-arg auditaddr=${auditaddr} \
--build-arg auditport=${auditport} \
--build-arg quoteaddr=${quoteaddr} \
--build-arg quoteport=${quoteport} \
--build-arg triggeraddr=${triggeraddr} \
--build-arg triggerport=${triggerport} \
-t teamrandint/transactionserver .

cd ../WebServer
docker image build \
--build-arg webaddr=${webaddr} \
--build-arg webport=${webport} \
--build-arg auditaddr=${auditaddr} \
--build-arg auditport=${auditport} \
--build-arg transaddr=${transaddr} \
--build-arg transport=${transport} \
-t teamrandint/webserver . 

cd ../database
docker image build \
--build-arg dbaddr=${dbaddr} \
--build-arg dbport=${dbport} \
-t teamrandint/database . 

cd ../quoteserver
docker image build \
--build-arg quoteaddr=${quoteaddr} \
--build-arg quoteport=${quoteport} \
--build-arg auditaddr=${auditaddr} \
--build-arg auditport=${auditport} \
--build-arg legacyquoteaddr=${legacyquoteaddr} \
--build-arg legacyquoteport=${legacyquoteport} \
-t teamrandint/quoteserver .

cd ../triggerserver
docker image build \
--build-arg triggeraddr=${triggeraddr} \
--build-arg triggerport=${triggerport} \
--build-arg quoteaddr=${quoteaddr} \
--build-arg quoteport=${quoteport} \
--build-arg auditaddr=${auditaddr} \
--build-arg auditport=${auditport} \
--build-arg transaddr=${transaddr} \
--build-arg transport=${transport} \
-t teamrandint/triggerserver .

docker pull dockercloud/haproxy

cd ../parent
docker save teamrandint/triggerserver teamrandint/quoteserver teamrandint/transactionserver teamrandint/database teamrandint/webserver teamrandint/auditserver dockercloud/haproxy > images.tar

