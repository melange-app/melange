set -e

if [ ! -f run.sh ] ; then
    echo 'can only be run from getmelange.com/mobile/android'
    exit 1
fi

if [ -f assets/client/ ] ; then
    rm -rf assets/client/
fi

cp -r ../../client assets/

lessc --clean-css assets/client/css/melange.less > assets/client/css/melange.css
rm assets/client/css/*.less

if [ -f assets/lib/ ] ; then
    rm -rf assets/lib/
fi

cp -r ../../lib assets/

docker run -v $GOPATH/src:/src mobile /bin/bash -c 'cd /src/getmelange.com/mobile/android && ./make.bash' && ./all.bash
