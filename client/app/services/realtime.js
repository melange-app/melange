(function() {
  function wsURL(s) {
    var l = window.location;
    return ((l.protocol === "https:") ? "wss://" : "ws://") + melangeAPI + s;
  }

  var melangeServices = angular.module('melangeServices');
  melangeServices.factory('mlgRealtime', ['$rootScope', 'mlgCandyBar', 'mlgMessages', function($rootScope, mlgCandyBar, mlgMessages) {
    var subscribers = {};

    var connectCandy = mlgCandyBar.setCandy("<p>Connecting to backend...</p>");

    // var conn = new WebSocket(wsURL("/messages"))
    var conn = new WebSocket("ws://api.melange.127.0.0.1.xip.io:7776/realtime");

    conn.onopen = function (event) {
      console.log("Opened realtime connection.");

      $rootScope.$apply(function() {
        mlgCandyBar.stopCandy(connectCandy);
        mlgCandyBar.setCandy("<p><i class='fa fa-spin fa-circle-o-notch'></i> Loading Messages...</p>")
      });

      sendData("startup", "ok");
    };
    conn.onmessage = function(event) {
      // console.log("Received from server")
      // console.log(event.data)
      var msg = JSON.parse(event.data)
      if (msg["type"] == "initDone") {
        $rootScope.$apply(function() {
          console.log("All messages loaded and in position.")
          mlgCandyBar.stopCandy();
          console.log(mlgMessages.getMessages().length);
        })

        return;
      } else if (msg["type"] == "message") {
        $rootScope.$evalAsync(function() {
          mlgMessages.addMessage(msg["data"])
        })

        return;
      } else if (msg["type"] == "waitingForSetup") {
        $rootScope.$apply(function() {
          mlgCandyBar.stopCandy();
        });
      }

      if(msg["type"] in subscribers) {
        var receivers = subscribers[msg["type"]]
        var toRemove = [];
        for(var i in receivers) {
          try {
            receivers[i].callback(msg["data"])
          } catch (err) {
            console.log("Couldn't get to callback.")
            console.log(err)
            toRemove.push(i)
          }
        }

        for (var i in toRemove) {
          delete receivers[toRemove[i]];
        }
        subscribers[msg["type"]] = receivers;
      } else {
        console.log("Got unknown message with type " + msg["type"])
      }
    }

    function sendData(type, data) {
      // Construct a msg object containing the data the server needs to process the message from the chat client.
      var msg = {
        type: type,
        data: data,
      };

      // Send the msg object as a JSON-formatted string.
      conn.send(JSON.stringify(msg));
    }

    return {
      send: sendData,
      subscribe: function(type, callback) {
        if(!(type in subscribers)) {
          subscribers[type] = [];
        }

        subscribers[type].push({
          callback: callback,
        });

        return subscribers[type].length - 1;
      },
      unsubscribe: function(type, id) {
        if(!(type in subscribers && id < subscribers[type].length)) {
          return
        }
        subscribers[type].removeAt(id);
      },
    }
  }]);

  melangeServices.factory('mlgMessages', [function() {
    var messages = [];
    var selfMessages = [];

    var msgCompare = function(a, b) {
        if(a.date < b.date) { return 1; }
        if(a.date > b.date) { return -1; }
        return 0;
    }

    var subscribers = [];

    return {
      addMessage: function(data) {
        console.log("Adding Message " + data.name);
        console.log(data);

        // Add message to global list.
        messages.unshift(data);
        messages.sort(msgCompare);

        // Add message to local list.
        // if(messagesFrom[data.from.fingerprint] === undefined) {
        //   messagesFrom[data.from.fingerprint] = [];
        // }
        // messagesFrom[data.from.fingerprint].unshift(data);
        // messagesFrom[data.from.fingerprint].sort(msgCompare);

        if(data.self) {
          selfMessages.unshift(data);
          selfMessages.sort(msgCompare);
        }

        for(var i in subscribers) {
          try {
            subscribers[i].callback(data);
          } catch (e) {
            console.log("Likely that mlgMessages subscriber left.")
            console.log(e);
            delete subscribers[i];
          }
        }
      },
      subscribe: function(callback) {
        subscribers.push({
          callback: callback,
        });
      },
      getMessages: function(obj) {
        // Only give back self messages.
        if(obj !== undefined && obj.self == true && obj.received == false && obj.public == false) { return selfMessages; }

        return messages;
      },
      getSpecificMessages: function(user) {
        var out = messagesFrom[user];
        if (out === undefined) { return []; }
        return out;
      },
    };
  }]);
})()
