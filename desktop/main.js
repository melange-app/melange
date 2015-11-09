var app = require('app');  // Module to control application life.
var BrowserWindow = require('browser-window');  // Module to create native browser window.

var buildMenu = require('./menu.js');

var VERSION = app.getVersion();
console.log("Running Melange v" + VERSION);

// Report crashes to our server.
require('crash-reporter').start();

// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the javascript object is GCed.
var mainWindow = null;

// Quit when all windows are closed.
app.on('window-all-closed', function() {
    // if (process.platform != 'darwin')
    app.quit();
});

// Functions for Spawning our Cousin Processes
var sys = require('sys')
    var spawn = require('child_process').spawn;

function logSpawn(spawn, name) {
    spawn.stdout.on('data', function (data) {
        data = data + ""
        data = data.replace(/(\r\n|\n|\r)$/gm,"");
        console.log('[' + name + '] ' + data);
    });
    spawn.stderr.on('data', function (data) {
        data = data + ""
        data = data.replace(/(\r\n|\n|\r)$/gm,"");
        console.log('[' + name + '] ERROR: ' + data);
    });
    spawn.on('close', function (code) {
        console.log('[' + name + '] Exited with Code ' + code);
    });
}

var dataDirectory = function() {
    return app.getDataPath();
}

var platform = function() {
    if(process.platform === 'darwin') {
        return 'mac'
    } else if (process.platform === 'win32') {
        return 'windows'
    } else {
        return 'linux'
    }
}

var path = require('path');
var applicationDirectory = function() {
    if(process.platform === 'darwin') {
        return path.join(process.execPath, "..", "..", "..");
    } else {
        return path.join(process.execPath, "..", "..");
    }
}


var debug = (process.argv.length === 3 && process.argv[2] === 'debug');
var go = {};

app.commandLine.appendSwitch("--host-rules", "MAP *.melange 127.0.0.1, MAP *.local.getmelange.com 127.0.0.1")

    global["__dirname"] = __dirname
// This method will be called when atom-shell has done everything
// initialization and ready for creating browser windows.
app.on('ready', function() {
    // Create the browser window.
    mainWindow = new BrowserWindow({
        "width": 800,
        "height": 600,
        "title": "Melange",
        "min-width": 320,
        "min-height": 480,
    });
    buildMenu(app);
    // and load the index.html of the app.
    if (!readyToLaunch) {
        mainWindow.loadUrl("file://" + __dirname + "/loading.html");
    } else {
        continueLaunching();
    }

    // Emitted when the window is closed.
    mainWindow.on('closed', function() {
        // Dereference the window object, usually you would store windows
        // in an array if your app supports multi windows, this is the time
        // when you should delete the corresponding element.
        mainWindow = null;
    });

    mainWindow.focus();

    var ipc = require("ipc");
    var dialog = require('dialog');

    ipc.on("start-upload", function(e, args) {
        var title = "Melange File Upload";
        if("title" in args) {
            title += (": " + args.title);
        }

        var options = {
            title: title,
        }
        if("options" in args) {
            options = args.options;
        }

        dialog.showOpenDialog(options, function(data) {
            e.sender.send("got-file", {
                data: data,
                id: args.id,
            });
        })
    });
});

var launched = false;
var readyToLaunch = false;
var continueLaunching = function() {
    mainWindow.loadUrl('http://app.local.getmelange.com:7776/Index.html#startup');
    launched = true;
}

var checkedLaunched = function() {
    if(!launched)
        continueLaunching;
}

app.on('will-finish-launching', function() {
    var http = require("http");

    function onRequest(request, response) {
        if(request.url === "/startup") {
            console.log("Starting up.");
            response.writeHead(200, {"Content-Type": "text/plain"});
            response.write("Starting up...");
            response.end();
            if(mainWindow === null) {
                readyToLaunch = true;
                // Hacky...
                setTimeout(checkLaunched, 1000);
            } else {
                continueLaunching();
            }
        } else if (request.url === "/kill") {
            app.quit();
        } else {
            response.writeHead(404, {"Content-Type": "text/plain"});
            response.write("That request is not allowed.");
            response.end();
        }
    }
    var server = http.createServer(onRequest).listen(0, "127.0.0.1", function() {
        // Listening
        console.log("opened application server on %j", server.address())
            // Launch Go Server
            if(debug) {
                go = spawn("go", ["run", __dirname + "/server/server.go"], {
                    env: {
                        "GOPATH": process.env["GOPATH"],
                        "GOROOT": process.env["GOROOT"],
                        "PATH": process.env["PATH"],
                        "CWD": process.cwd,
                        "MLGBASE": __dirname + "/../",
                        "MLGDATA": dataDirectory(),
                        "MLGPORT": server.address().port,
                        "MLGAPP": applicationDirectory(),
                        "MLGPLATFORM": platform(),
                        "MLGVERSION": VERSION,
                        "MLGDEBUG": "1",
                    },
                });
                logSpawn(go, "SERVER");
            } else {
                go = spawn(__dirname + "/bin/server", [], {
                    env: {
                        "PATH": process.env["PATH"],
                        "CWD": process.cwd,
                        "MLGBASE": __dirname,
                        "MLGDATA": dataDirectory(),
                        "MLGPORT": server.address().port,
                        "MLGAPP": applicationDirectory(),
                        "MLGPLATFORM": platform(),
                        "MLGVERSION": VERSION,
                    },
                });
                logSpawn(go, "SERVER");
            }
    });


    // var protocol = require('protocol');
    // protocol.registerProtocol("melange", function(request) {
    //   console.log(request);
    //   // var url = request.url.substr(7)
    //   // return new protocol.RequestFileJob(path.normalize(__dirname + '/' + url));
    // });
});

app.on('open-url', function(url) {
    console.log("Opened URL!")
        console.log(url);
})

    app.on('will-quit', function() {
        console.log("Will Quit Fired");
        go.kill();
    });
