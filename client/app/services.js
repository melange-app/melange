'use strict';

/* Services */
var melangeServices = angular.module('melangeServices', []);

(function() {
  function endsWith(str, suffix) {
      return str.indexOf(suffix, str.length - suffix.length) !== -1;
  }

  // Resources to be Deleted
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

  // MLG-SETUP
  melangeServices.factory('mlgIdentity', ['$resource', function($resource) {
    var saveIdentity = $resource('http://' + melangeAPI + '/identity/new', {}, {save: {method: 'POST'}});
    return {
      profile: {},
      save: function(onsuccess, onerror) {
        console.log("Saving Identity.");
        return saveIdentity.save(this.profile,
          // Success
          function(value, responseHeaders) {
            onsuccess();
          },
          // Error
          function(res) {
            onerror();
          }
        );
      },
      servers: $resource('http://' + melangeAPI + '/servers', {}, {query: {method:'GET', isArray:true}}).query,
      trackers: $resource('http://' + melangeAPI + '/trackers', {}, {query: {method:'GET', isArray:true}}).query,
    };
  }]);


  // MLG-API
  melangeServices.factory('mlgApi', ['$resource', function($resource) {
    return {
      // Contact Management
      lists: function() {
        return ["Friends", "Family"];
      },
      contacts: function() {
        return [
          {
            name: "Hunter Leath",
            favorite: false,
            identities: [
              {
                address: "4073f3dff85852fc8c0c206599b7e221d8c7a77f085a9497",
              },
            ],
            lists: ["Family"],
          },
          {
            name: "Dalton Sherwood",
            favorite: true,
            identities: [
              {
                address: "4073f3dff85852fc8c0c206599b7e221d8c7a77f085a9497",
              },
            ],
            lists: ["Friends"],
          },
        ];
      },
      // Message Management
      getMessages: $resource('http://' + melangeAPI + '/messages', {}, {query: {method:'GET', isArray:true}}).query,
      // Identity Management
      newIdentity: function(id) {},
      identities: $resource('http://' + melangeAPI + '/identity', {}, {query: {method:'GET', isArray:true}}).query,
      setCurrent: $resource('http://' + melangeAPI + '/identity/current', {}, {save: {method:'POST'}}).$save,
      current: $resource('http://' + melangeAPI + '/identity/current', {}, {query: {method:'GET'}}).query,
    }
  }]);

  // MLG-PLUGINS
  melangeServices.factory('mlgPlugins', ['$resource', function($resource) {
    // Plugins Resource
    var plugins = $resource('http://' + melangeAPI + '/plugins', {}, {query: {method:'GET', isArray:true}});

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
