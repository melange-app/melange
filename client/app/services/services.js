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
  melangeServices.factory('mlgApi', ['$resource', '$q', function($resource, $q) {
    return {
      // Contact Management
      lists: function() {
        return ["Friends", "Family"];
      },
      contacts: $resource('http://' + melangeAPI + '/contacts', {}, {query: {method: 'GET', isArray: true}}).query,
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
