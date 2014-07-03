'use strict';

/* Services */
var melangeServices = angular.module('melangeServices', []);

(function() {
  function endsWith(str, suffix) {
      return str.indexOf(suffix, str.length - suffix.length) !== -1;
  }

  var mlgMessage = {
    id: "000000",
    from: {
      name: "Hunter Leath",
      address: "0x0F",
      avatar: "http://placehold.it/400x400",
    },
    to: {
      name: "Joseph Barrow",
      address: "0x0F",
      avatar: "http://placehold.it/400x400",
    },
    time: new Date(0),
    data: {
      "airdispat.ch/note/title": "Hello, Joe",
      "airdispat.ch/note/body": "This pot roast smells delicious.",
    }
  }

  var mlgStatus = {
    id: "00000",
    error: {
      code: 0,
      message: "",
    }
  }

  melangeServices.factory('mlgApi', ['$resource', function($resource) {

  }]);

  melangeServices.factory('mlgPlugins', ['$resource', function($resource) {
    // Plugins Resource
    var plugins = $resource('/plugins.json', {}, {query: {method:'GET', isArray:true}});

    var allPlugins = {};
    plugins.query(function(value) {
      for (var index in value) {
        allPlugins[value[index].id] = value[index]
      }
    });

    // --- Plugin Communication
    var receivers = {
      createMessage: function(origin, data, callback) {
        callback({
          type: "createdMessage",
          context: mlgStatus,
        })
      },
      findMessages: function(origin, data, callback) {
        console.log(origin)
        console.dir(allPlugins[origin]);
        callback({
          type: "foundMessages",
          context: [mlgMessage],
        })
      },
      updateMessage: function(origin, data, callback) {
        callback({
          type: "updatedMessage",
          context: mlgStatus,
        })
      },
      downloadMessage: function(origin, data, callback) {
        callback({
          type: "downloadedMessage",
          context: mlgMessage,
        })
      },
      downloadPublicMessages: function(origin, data, callback) {
        callback({
          type: "downloadedPublicMessages",
          context: [mlgMessage],
        })
      },
    };
    function messenger(source, data, origin) {
      source.postMessage(data, origin);
    }
    function receiveMessage(e) {
      if (!endsWith(e.origin, melangePluginSuffix))
        return
      if (typeof e.data["context"] !== "object" || typeof e.data["type"] !== "string")
        return
      if (typeof receivers[e.data.type] !== "function") {
        console.log("Couldn't understand message type " + e.data.type)
        return
      }
      var origin = e.origin.substr(7, (e.origin.length - 7 - melangePluginSuffix.length));
      receivers[e.data.type](origin, e.data.context, function(output) {
        messenger(e.source, output, e.origin)
      });
    }
    window.addEventListener("message", receiveMessage, false);
    // --- Plugin Communication

    return plugins
  }]);
})()
