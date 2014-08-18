var app = require('app');  // Module to control application life.
var BrowserWindow = require('browser-window');  // Module to create native browser window.

require('./menu.js')(app)

var VERSION = "0.0.2";

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
  if(process.platform === 'darwin') {
    return process.env.HOME + "/Library/Application Support/Melange";
  } else if (process.platform === 'win32') {
    return process.env.APPDATA + "/Melange"
  } else {
    return process.env.HOME + "/.melange";
  }
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

global["__dirname"] = __dirname
// This method will be called when atom-shell has done everything
// initialization and ready for creating browser windows.
app.on('ready', function() {
  // Create the browser window.
  mainWindow = new BrowserWindow({
    "width": 800,
    "height": 600,
    "title": "Melange",
    "frame": false,
    "min-width": 769,
    "min-height": 600,
  });
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
});

var launched = false;
var readyToLaunch = false;
var continueLaunching = function() {
  mainWindow.loadUrl('http://app.melange.127.0.0.1.xip.io:7776/Index.html#startup');
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
      go = spawn("go", ["run", __dirname + "/go/src/melange/server/server.go"], {
        env: {
          "GOPATH": __dirname + '/go',
          "GOROOT": process.env["GOROOT"],
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
    } else {
      go = spawn(__dirname + "/bin/server", [], {
        env: {
          "GOPATH": __dirname + '/go',
          "GOROOT": process.env["GOROOT"],
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
});

app.on('will-quit', function() {
  console.log("Will Quit Fired");
  go.kill('SIGKILL')
});
