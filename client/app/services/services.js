'use strict';

/* Services */

var mlgCleanup = function(msg) {
    if(msg.$promise) { delete msg.$promise; }
    if(msg.$resolved) { delete msg.$resolved; }
    return msg;
};

(function() {
  var melangeServices = angular.module('melangeServices', []);
  var cleanup = mlgCleanup;

  // MLG-TILES
  melangeServices.factory('mlgTiles', ['$resource', '$q', function($resource, $q) {
    var tiles = [
      {
        size: "12",
        height: "85",
        url: "http://" + "ch.airdispat.plugins.status" + melangePluginSuffix + "/tile.html",
        click: true,
      },
      {
        size: "6",
        height: "150",
        url: "http://" + "ch.airdispat.plugins.news" + melangePluginSuffix + "/tile.html",
        name: "News",
      },
      {
        size: "6",
        height: "150",
        url: "http://" + "ch.airdispat.plugins.news" + melangePluginSuffix + "/tile.html",
        name: "Family",
      }
    ];

    return {
      all: function() {
        var defer = $q.defer();
        setTimeout(function() {
          defer.resolve(tiles);
        }, 0);
        return defer.promise;
      }
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
      publishMessage: $resource('http://' + melangeAPI + '/messages/new', {}, {create: {method:'POST'}}).create,
      getMessages: $resource('http://' + melangeAPI + '/messages', {}, {query: {method:'GET', isArray:true}}).query,
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
