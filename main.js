var app = require('app');  // Module to control application life.
var BrowserWindow = require('browser-window');  // Module to create native browser window.

require('./menu.js')(app)

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


// app.commandLine.appendSwitch("host-rules", "MAP * 127.0.0.1");

var debug = true;
var go = {};

app.on('will-finish-launching', function() {
  // Launch Go Server
  if(debug) {
    go = spawn("go", ["run", __dirname + "/go/src/melange/server/server.go"], {
      env: {
        "GOPATH": __dirname + '/go',
        "GOROOT": process.env["GOROOT"],
        "PATH": process.env["PATH"],
        "CWD": process.cwd,
        "MLGBASE": __dirname,
      },
    });
    logSpawn(go, "PLUGIN");
  } else {
      go = spawn("server/bin/melange");
      logSpawn(go, "PLUGIN");
  }
});

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
  // mainWindow.loadUrl('http://app.melange.127.0.0.1.xip.io:9001/Index.html#startup');
  mainWindow.loadUrl("file://" + __dirname + "/loading.html");

  // Emitted when the window is closed.
  mainWindow.on('closed', function() {
    // Dereference the window object, usually you would store windows
    // in an array if your app supports multi windows, this is the time
    // when you should delete the corresponding element.
    mainWindow = null;
  });

  mainWindow.focus();
});

app.on('will-quit', function() {
  console.log("Will Quit Fired");
  go.kill('SIGKILL')
});
