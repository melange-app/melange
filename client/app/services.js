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

  melangeServices.factory('mlgHelper', ['$rootScope', function($rootScope) {
    return {
      promise: function(result, func) {
            if (!result || !func) return result;

            var then = function(promise) {
                //see if they sent a resource
                if ('$promise' in promise) {
                    promise.$promise.then(update);
                }
                //see if they sent a promise directly
                else if ('then' in promise) {
                    promise.then(update);
                }
            };

            var update = function(data) {
                if ($.isArray(result)) {
                    //clear result list
                    result.length = 0;
                    //populate result list with data
                    $.each(data, function(i, item) {
                        result.push(item);
                    });
                } else {
                    //clear result object
                    for (var prop in result) {
                        if (prop !== 'load') delete result[prop];
                    }
                    //deep populate result object from data
                    $.extend(true, result, data);
                }
            };

            //see if they sent a function that returns a promise, or a promise itself
            if ($.isFunction(func)) {
                // create load event for reuse
                result.load = function() {
                    then(func());
                };
                result.load();
            } else {
                then(func);
            }

            return result;
        },
      }
  }])

  // MLG-SETUP
  melangeServices.factory('mlgIdentity', ['$resource', '$q', function($resource, $q) {
    var resource = $resource('http://' + melangeAPI + '/identity/:action', {
      action: "",
    }, {
      new: {
        method: 'POST',
        params: {
          action: "new",
        }
      },
      current: {
        method: 'GET',
        params: {
          action: "current",
        }
      },
      setCurrent: {
        method: 'POST',
        params: {
          action: "current",
        }
      },
      list: {
        method: 'GET',
        isArray: true,
      }
    });

    var identities = [];
    var getIdentities = function(callback) {
      if(identities.length == 0) {
        var ids = resource.list({}, function(data) {
          identities = ids;
          callback(identities);
        }, function() {
          alert("Error getting identities.");
        })
      } else {
        callback(identities);
      }
    }

    return {
      current: function() {
        var defer = $q.defer();
        getIdentities(function(id) {
          for(var i in id) {
            if(id[i].Current) {
              defer.resolve(id[i]);
            }
          }
          defer.reject("Cannot get current identity.");
        });
        return defer.promise;
      },
      setCurrent: function(id) {
        return resource.setCurrent(id);
      },
      list: function() {
        var defer = $q.defer();
        getIdentities(function(id) {
          defer.resolve(id);
        });
        return defer.promise;
      },
      profile: {},
      save: function(onsuccess, onerror) {
        console.log("Saving Identity.");
        return resource.new(this.profile,
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
      publishMessage: $resource('http://' + melangeAPI + '/messages/new', {}, {create: {method:'POST'}}).create,
      getMessages: $resource('http://' + melangeAPI + '/messages', {}, {query: {method:'GET', isArray:true}}).query,
    }
  }]);

  // MLG-PLUGINS
  melangeServices.factory('mlgPlugins', ['$resource', 'mlgApi', function($resource, mlgApi) {
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
        mlgApi.publishMessage(data).then(
          function(data) {
            callback({
              type: "createdMessage",
              context: mlgStatus,
            })
          },
          function(err) {
            callback({
              type: "createdMessage",
              context: {error: {code: 1, message: "Something happened. Too lazy to find out what."}},
            })
          }
        );
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
