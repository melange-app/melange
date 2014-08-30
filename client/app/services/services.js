'use strict';

/* Services */

var mlgCleanup = function(msg) {
    if(msg.$promise) { delete msg.$promise; }
    if(msg.$resolved) { delete msg.$resolved; }
    // Remove Proto Messages
    if(msg.$delete) { delete msg.$delete }
    if(msg.$get) { delete msg.$get }
    if(msg.$query) { delete msg.$query }
    if(msg.$remove) { delete msg.$remove }
    if(msg.$save) { delete msg.$save }
    return msg;
};

// "status-updater": {
//   "name": "Quick Status Updater",
//   "description": "Quickly publish status updates at the top of your dashboard.",
//   "view": "tile.html",
//   "click": true,
//   "size": "100%x100"
// }

(function() {
  var melangeServices = angular.module('melangeServices', []);
  var cleanup = mlgCleanup;

  // MLG-TILES
  melangeServices.factory('mlgTiles', ['$resource', '$q', 'mlgPlugins', function($resource, $q, mlgPlugins) {
    var tilesResource = $resource("http://" + melangeAPI + "/tiles/:action", { action: "" },
    {
      current: {
        method: 'GET',
        params: {
          action: "current",
        },
        isArray: true,
      },
      update: {
        method: 'POST',
        params: {
          action: "update",
        }
      },
    });

    // {
    //   id: 2,
    //   size: "6",
    //   height: "150",
    //   url: "http://" + "ch.airdispat.plugins.news" + melangePluginSuffix + "/tile.html",
    //   name: "Family",
    // }

    var startsWith = function(str, needle) {
      return (str.indexOf(needle) === 0)
    }

    var parse = function(plugin, tileKey) {
      var theTile = plugin.tiles[tileKey];

      var parsedTile = {
        id: plugin.id + "|" + tileKey,
        url: "http://" + plugin.id + melangePluginSuffix + "/" + theTile.view,
        click: theTile.click,
      };

      // if (theTile["title"] !== undefined) {
      //   parsedTile["name"] = theTile["title"];
      // }

      var size = theTile['size'];


      if(startsWith(size, "100%")) {
        var height = size.split("x")[1];
        parsedTile["size"] = "12";
        parsedTile["height"] = height;
      } else if (startsWith(size, "50%")) {
        parsedTile["size"] = "6";
        parsedTile["height"] = "150";
      }

      return parsedTile;
    }

    return {
      all: function() {
        var defer = $q.defer();
        tilesResource.current().$promise.then(function(data) {
          mlgPlugins.all().then(function(plugins) {
            if(Object.keys(plugins).length == 0) {
              defer.resolve([]);
              return;
            }

            data = cleanup(data);
            var tiles = [];
            for(var i in data) {
              var components = data[i].split("|");
              if(angular.isDefined(plugins[components[0]])) {
                var tile = plugins[components[0]].tiles[components[1]]
                tiles.push(parse(plugins[components[0]], components[1]))
              }
            }
            defer.resolve(tiles);
          });
        }, function(err) {
          console.log("Error loading tiles");
          console.log(err)
          defer.reject(err);
        });
        return defer.promise;
      },
      parse: parse,
      update: function(tiles) {
        var update = [];
        for(var i in tiles) {
          update.push(tiles[i].id);
        }

        return tilesResource.update(update).$promise
      },
    }
  }]);

  // MLG-IDENTITY
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

    var current = {};
    var identities = [];
    var currentIdentity = function(defer) {
      getIdentities(function(id) {
        for(var i in id) {
          if(id[i].Current) {
            angular.copy(id[i], current);
            defer.resolve(current);
          }
        }
      });
    };
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
        angular.copy([], identities);
        angular.copy({}, current);
        var defer = $q.defer();
        currentIdentity(defer);
        return defer.promise;
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
        currentIdentity(defer);
        return defer.promise;
      },
      setCurrent: function(id) {
        // Reload the view
        if(id.Current) { return }

        for(var i in identities) {
          identities[i].Current = false;
        }
        id.Current = true;
        angular.copy(id, current);

        return resource.setCurrent({
          fingerprint: id.Fingerprint,
        });
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
  melangeServices.factory('mlgApi', ['$resource', '$q', function($resource, $q) {
    var apiResource = $resource('http://' + melangeAPI + '/:action', {}, {
      // Messages
      getMessage: {
        method: 'POST',
        params: {
          action: "messages/get",
        }
      },
      getMessagesAtAlias: {
        method: 'POST',
        isArray: true,
        params: {
          action: "messages/at",
        }
      },
      // Contacts
      contacts: {
        method: 'GET',
        isArray: true,
        params: {
          action: "contacts",
        }
      },
      updateContact: {
        method: 'POST',
        params: {
          action: "contacts/update",
        }
      },
      // Updates
      updateCheck: {
        method: 'GET',
        params: {
          action: "update",
        }
      },
      downloadUpdate: {
        method: 'POST',
        params: {
          action: "update/download",
        }
      },
      downloadProgress: {
        method: 'POST',
        params: {
          action: "update/download/progress",
        }
      },
      installUpdate: {
        method: 'POST',
        params: {
          action: "update/install",
        }
      },
      // Applications
      installApp: {
        method: 'POST',
        params: {
          action: "app/install",
        }
      },
      uninstallApp: {
        method: 'POST',
        params: {
          action: "app/uninstall",
        }
      }
    })

    var contacts = [];
    var getContacts = function(defer) {
      if(contacts.length == 0) {
        console.log("Getting contacts.");
        apiResource.contacts().$promise.then(
          function(data) {
            for(var i in data) {
              data[i].subscribed = false;
              if(data[i].addresses !== undefined && data[i].addresses.length !== 0) {
                data[i].subscribed = data[i].addresses[0].subscribed
              }
            }
            contacts = data;
            defer.resolve(contacts);
          },
          function(err) {
            console.log("Error getting contacts.")
            console.log(err);
          }
        );
      } else {
        defer.resolve(contacts);
      }
    };

    return {
      update: {
        check: apiResource.updateCheck,
        download: apiResource.downloadUpdate,
        progress: apiResource.downloadProgress,
        install: apiResource.installUpdate,
      },
      // Contact Management
      lists: function() {
        return ["Friends", "Family"];
      },
      contacts: function() {
        var defer = $q.defer();
        console.log("Called contacts.");
        getContacts(defer);
        return defer.promise;
      },
      updateContact: function(contact) {
        return apiResource.updateContact(contact).$promise;
      },
      // Message Management
      publishMessage: function(data) {
        console.log(data);
        return $resource('http://' + melangeAPI + '/messages/new', {}, {create: {method:'POST'}}).create(data);
      },
      getMessage: function(alias, name) {
        return apiResource.getMessage({
          alias: alias,
          name: name,
        }).$promise
      },
      getMessages: function(self, pub, received) {
        var obj = {
          public: pub,
          self: self,
          received: received,
        }
        if(arguments.length === 0) {
          obj = {
            public: true,
            self: true,
            received: true,
          };
        }
        return $resource('http://' + melangeAPI + '/messages', {}, {query: {method:'POST', isArray:true}}).query(obj);
      },
      getMessagesAtAlias: function(alias, onlyPublic) {
        if(arguments.length === 1) { onlyPublic = true; }

        return apiResource.getMessagesAtAlias({
          alias: alias,
          onlyPublic: onlyPublic,
        }).$promise;
      },
      // Profile Management
      updateProfile: function(profile) {
        return $resource('http://' + melangeAPI + '/profile/update', {}, {updateProfile: {method:'POST'}}).updateProfile(profile).$promise
      },
      currentProfile: function() {
        var defer = $q.defer();
        $resource('http://' + melangeAPI + '/profile/current', {}, {get: {method:'GET'}}).get().$promise.then(
          function(data) {
            defer.resolve(data);
          },
          function(err) {
            if(err.status == 422) {
              defer.reject(true);
            }
            defer.reject(err);
          }
        )
        return defer.promise;
      },
    }
  }]);
})()
