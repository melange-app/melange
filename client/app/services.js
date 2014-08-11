'use strict';

/* Services */
var melangeServices = angular.module('melangeServices', []);

(function() {
  function endsWith(str, suffix) {
      return str.indexOf(suffix, str.length - suffix.length) !== -1;
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
      refresh: function() {
        identities = [];
        getIdentities(function(id) {});
      },
      startup: function() {
        var defer = $q.defer();
        resource.current().$promise.then(
        function(s) { defer.resolve(true) },
        function(obj) {
          if(obj.status == 422) {
            defer.resolve(false);
          } else {
            defer.reject();
          }
        });
        return defer.promise
      },
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
      contacts: $resource('http://' + melangeAPI + '/contacts', {}, {query: {method: 'GET', isArray: true}}).query,
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

    var cleanup = function(msg) {
        if(msg.$promise) { delete msg.$promise; }
        if(msg.$resolved) { delete msg.$resolved; }
    }

    // --- Plugin Communication
    var receivers = {
      createMessage: function(origin, data, callback) {
        mlgApi.publishMessage(data).$promise.then(
          function(data) {
            callback({
              type: "createdMessage",
              context: mlgStatus,
            })
          },
          function(err) {
            console.log(err)
            callback({
              type: "createdMessage",
              context: {error: {code: 1, message: "Something happened. Too lazy to find out what."}},
            })
          }
        );
      },
      findMessages: function(origin, data, callback) {
        mlgApi.getMessages().$promise.then(
          function(msg) {
            cleanup(msg)
            callback({
              type: "foundMessages",
              context: msg,
            });
          },
          function(err) {
            console.log(err)
            callback({
              type: "foundMessages",
              context: {error: {code: 1, message: "Something happened. Too lazy to find out what."}}
            })
          }
        );
      },
      updateMessage: function(origin, data, callback) {
        callback({
          type: "updatedMessage",
          context: mlgStatus,
        })
      },
      downloadMessage: function(origin, data, callback) {
        mlgApi.getMessages().$promise.then(
          function(msg) {
            cleanup(msg);
            callback({
              type: "downloadedMessage",
              context: msg[0],
            });
          },
          function(err) {
            console.log(err)
            callback({
              type: "downloadedMessage",
              context: {error: {code: 1, message: "Something happened. Too lazy to find out what."}}
            })
          }
        );
      },
      downloadPublicMessages: function(origin, data, callback) {
        mlgApi.getMessages().$promise.then(
          function(msg) {
            cleanup(msg);
            callback({
              type: "downloadedPublicMessages",
              context: msg,
            });
          },
          function(err) {
            console.log(err)
            callback({
              type: "downloadedPublicMessages",
              context: {error: {code: 1, message: "Something happened. Too lazy to find out what."}}
            })
          }
        );
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
