# alias goios=/Users/hunter/Documents/Developer/Melange/goios/bin/go
# CGO_ENABLED=1 GOARCH=arm goios build getmelange.com/mobile/ios/main
# rm Melange.app/Melange
# mv Melange Melange.app/
codesign -f -s "Hunter Leath" --entitlements Melange.app/Entitlements.plist Melange.app/Melange
ios-deploy --debug --uninstall --bundle Melange.app

