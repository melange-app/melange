set -e

if [ ! -f sign.sh ] ; then
    echo 'can only be run from getmelange.com/mobile/ios'
    exit 1
fi

# Copy the Assets into the Bundle
if [ -f Melange.app/Assets/client/ ] ; then
    rm -rf Melange.app/Assets/client/
fi

cp -r ../../client Melange.app/Assets/

lessc --clean-css Melange.app/Assets/client/css/melange.less > Melange.app/Assets/client/css/melange.css
rm Melange.app/Assets/client/css/*.less

if [ -f Melange.app/Assets/lib/ ] ; then
    rm -rf Melange.app/Assets/lib/
fi

cp -r ../../lib Melange.app/Assets/


# Compile the Golang Application
alias goios=/Users/hunter/Documents/Developer/Melange/goios/bin/go
GOROOT=/Users/hunter/Documents/Developer/Melange/goios/ CGO_ENABLED=1 GOARCH=arm goios build -o "Melange" getmelange.com/mobile/ios/main

# Delete the Old Melange
rm Melange.app/Melange

# Install the New Melange
mv Melange Melange.app/

# Codesign and Install the Application on the Phone
codesign -f -s "Hunter Leath" --entitlements Melange.app/Entitlements.plist Melange.app/Melange
ios-deploy --debug --uninstall --bundle Melange.app

