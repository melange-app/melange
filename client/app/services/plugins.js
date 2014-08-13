'use strict';

(function() {
  var melangeServices = angular.module('melangeServices');
  var endsWith = function(str, suffix) {
      return str.indexOf(suffix, str.length - suffix.length) !== -1;
  };
  var cleanup = mlgCleanup;

  var mlgStatus = {
    id: "00000",
    error: {
      code: 0,
      message: "",
    }
  }

  // MLG-PLUGINS
  melangeServices.factory('mlgPlugins', ['$resource', '$q', 'mlgApi', function($resource, $q, mlgApi) {
    // Plugins Resource
    var plugins = $resource('http://' + melangeAPI + '/plugins', {}, {query: {method:'GET', isArray:true}});

    var allPlugins = {};
    var havePlugins = false;
    var getAllPlugins = function(callback) {
      if(!havePlugins) {
        plugins.query(function(value) {
          for (var index in cleanup(value)) {
            allPlugins[value[index].id] = value[index]
          }
          callback(allPlugins);
        });
      } else {
        callback(allPlugins);
      }
    }

    // --- Plugin Communication
    var receivers = {
      viewerUpdate: function(origin, data, callback, obj) {
        if(obj !== undefined) {
          obj.element.style.height = data["height"] + "px";
        }
        if(data.sendMsg === true) {
          callback({
            type: "viewerMessage",
            context: obj.context,
          })
        }
      },
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
      var frame = undefined;
      for(var p in registeredPlugins) {
        for (var i in registeredPlugins[p]) {
          if(registeredPlugins[p][i].element.contentWindow == e.source) {
            frame = registeredPlugins[p][i];
          }
        }
      }
      var origin = e.origin.substr(7, (e.origin.length - 7 - melangePluginSuffix.length));
      receivers[e.data.type](origin, e.data.context, function(output) {
        messenger(e.source, output, e.origin)
      }, frame);
    }
    window.addEventListener("message", receiveMessage, false);
    // --- Plugin Communication

    var registeredPlugins = {};

    return {
      registerPlugin: function(plugin, elem, type, context) {
        if(registeredPlugins[plugin.id] === undefined) { registeredPlugins[plugin.id] = []; }
        registeredPlugins[plugin.id].push({
          element: elem,
          type: type,
          context: context,
        })
      },
      unregisterPlugin: function(plugin, elem) {
        // Something, something
      },
      all: function() {
        var defer = $q.defer();
        getAllPlugins(function(all) {
          defer.resolve(all);
        });
        return defer.promise;
      },
      viewer: function(msg) {
        var defer = $q.defer();
        getAllPlugins(function(all) {
          // Loop over plugins
          for(var name in all) {
            // Loop over Viewer
            for(var v in all[name].viewers) {
              var works = true;
              var viewer = all[name].viewers[v]
              // Check that Components are correct
              for(var i in viewer["type"]) {
                if(msg.components[viewer["type"][i]] === undefined) {
                  works = false;
                  break;
                }
              }
              // Return the viewer
              if(works) {
                // Remove uneeded components
                for(var i in msg.components) {
                  if(viewer["type"].indexOf(i) == -1) {
                    delete msg.components[i];
                  }
                }

                defer.resolve([all[name], v]);
              }
            }
          }
          defer.reject();
        });
        return defer.promise;
      },
    }
  }]);

})()
