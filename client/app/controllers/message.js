'use strict';

(function() {
  var melangeControllers = angular.module('melangeControllers');

  melangeControllers.controller('AllCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    // Sync up
    var sync = function() {
      $scope.loading = true;
      mlgApi.getMessages().then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data;
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

  }]);

  melangeControllers.controller('DashboardCtrl', ['$scope', 'mlgPlugins', 'mlgHelper', 'mlgTiles', 'mlgApi', 'mlgRealtime',
  function($scope, mlgPlugins, mlgHelper, mlgTiles, mlgApi, mlgRealtime) {
    // Tile Information
    $scope.editDash = false;
    mlgTiles.all().then(function(tiles) {
      $scope.tiles = tiles;
    })

    var stopFunc = function(e, ui) {
      console.log($scope.tiles);
    }

    // Conditionally enable tile building
    $scope.$watch(function() { return $scope.editDash }, function(data) {
      $scope.sortableOptions = {
        stop: stopFunc,
        disabled: !data,
      }
    })

    mlgPlugins.all().then(function(plugins) {
      $scope.plugins = plugins;
    });

    $scope.adding = false;
    $scope.addTile = function(plugin, tileKey) {
      $scope.tiles.push(mlgTiles.parse(plugin, tileKey));
      $scope.adding = false;
      mlgTiles.update($scope.tiles);
    }

    $scope.deleteTile = function(index) {
      $scope.tiles.splice(index, 1);
      mlgTiles.update($scope.tiles);
    }

    mlgRealtime.subscribe("*", function(msg) {

    });

    // Sync up if needed.
    var sync = function() {
      console.log("Syncing")
      $scope.loading = true;
      mlgApi.getMessages().then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data;
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

    $scope.messageLimit = 50;
    $scope.messageCounter = 0;
    $scope.$watch("messageCounter", function(val) {
        console.log("Updated Counter");
        console.log(val);
    });
  }]);

  melangeControllers.controller('ProfileCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    $scope.newProfile = false;

    mlgApi.currentProfile().then(function(data) {
      console.log(data);

      var image = data.image;
      if(image.indexOf("@") != -1) {
        image = "http://data.local.getmelange.com:7776/" + image;
      }

      $scope.profile = data;
      $scope.profile.image = image;
    },
    function(err) {
      if(err === true) {
        $scope.newProfile = true;
      } else {
        console.log("Couldn't get profile. Something went wrong.")
        console.log(err)
      }
    });

    // Sync up if needed.
    $scope.me = true;
    var sync = function() {
      console.log("Syncing")
      $scope.loading = true;
      mlgApi.getMessages(true, false, false).then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data;
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

  }]);

  melangeControllers.controller('UserProfileCtrl', ['$scope', '$routeParams', 'mlgApi',
  function($scope, $routeParams, mlgApi) {
    $scope.profile = {
      name: $routeParams.alias
    }

    $scope.inContacts = $scope.contacts.reduce(function(acc, val, i, a) {
        if(val.profile.alias == $routeParams.alias)
            $scope.isFollowing = val.subscribed;
        return acc || (val.profile.alias == $routeParams.alias);
    }, false)

    mlgApi.getMessage($routeParams.alias, "profile").then(function(profile) {
      console.log(profile)

      var image = profile.components["airdispat.ch/profile/avatar"].string;
      if(image.indexOf("@") != -1) {
        image = "http://data.local.getmelange.com:7776/" + image;
      } else if (image == "") {
        image = "http://robohash.org/" + profile.from.fingerprint + ".png?bgset=bg2";
      }

      $scope.profile = {
        name: profile.components["airdispat.ch/profile/name"].string,
        description: profile.components["airdispat.ch/profile/description"].string,
        image: image,
      }
    }, function(err) {
      console.log("Got an error getting profile")
      console.log(err)
    });

    // Sync up if needed.
    var sync = function() {
      console.log("Syncing")
      $scope.loading = true;
      mlgApi.getMessagesAtAlias($routeParams.alias).then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data.reverse();
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

  }]);

  melangeControllers.controller('EditProfileCtrl', ['$scope', '$location', 'mlgFile', 'mlgApi',
  function($scope, $location, mlgFile, mlgApi) {

    mlgApi.currentProfile().then(function(data) {
      $scope.profile = data;
      $scope.newProfile = false;
    },
    function(err) {
      $scope.newProfile = true;
      $scope.profile = {};
    });

    $scope.upload = function() {
      $scope.uploading = true;
      mlgFile.upload("airdispat.ch/profile", [], "airdispat.ch/profile/image").then(
        function(data) {
          console.log("Completed upload!")
          console.log(data);
          $scope.uploading = false;
          $scope.avatar = data.url;
          $scope.profile.image = data.user + "/" + data.name;
        },
        function(error) {
          console.log("Error uploading profile.")
          console.log(error);
          $scope.uploading = false;
        },
        function(progress) {

        }
      );
    }

    $scope.save = function() {
      // Save the profile
      mlgApi.updateProfile($scope.profile).then(
        function() {
          $location.path("/profile");
        },
        function(err) {
          console.log("Error updating profile.")
          console.log(err)
        }
      )
    }
  }]);

})();
