#!/bin/sh

# check params
if [ $# -lt 3 ]
then
    echo "<Usage: sh $0  App  Server  Servant>"
    echo ">>>>>>  sh $0  TeleSafe PhonenumSogouServer SogouInfo"
    exit 1
fi

cd $(dirname .) || exit 1
SRC_DIR=$(dirname .)

APP=$1
SERVER=$2
SERVANT=$3

if [ "$SERVER" == "$SERVANT" ]
then
    echo "Error!(ServerName == ServantName)"
    exit 1
fi
echo "[create server: $APP.$SERVER ...]"

TARGET=$(echo $SERVER | awk '{ print tolower($1) }')

if [ -d $TARGET ];then
    echo "! Already have some file in $TARGET! Please clear files in prevent of overwrite!"
    exit 1
fi


DEMODIR=$SRC_DIR/Demo
echo "[mkdir: $TARGET]"
mkdir -p $TARGET

cp -r $DEMODIR/* $TARGET/

cd $TARGET || exit 1

SRC_FILE=`find . -maxdepth 2 -type f`

if [ `uname` == "Darwin" ] # support macOS
then
    for FILE in $SRC_FILE
    do
        echo ">>>Now doing:"$FILE" >>>>"
        sed  -i "" "s/_APP_/$APP/g"   $FILE
        sed  -i "" "s/_SERVER_/$SERVER/g" $FILE
        sed  -i "" "s/_SERVANT_/$SERVANT/g" $FILE
    done
else
    for FILE in $SRC_FILE
    do
        echo ">>>Now doing:"$FILE" >>>>"
        sed -i "s/_APP_/$APP/g"   $FILE
        sed -i "s/_SERVER_/$SERVER/g" $FILE
        sed -i "s/_SERVANT_/$SERVANT/g" $FILE
    done
fi

# try build tars2go
go install github.com/TarsCloud/TarsGo/tars/tools/tars2go
echo ">>> Great！Done! You can jump in "`pwd`
