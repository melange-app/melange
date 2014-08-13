'use strict';

(function() {
  var melangeControllers = angular.module('melangeControllers');

  melangeControllers.controller('AllCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    // Sync up
    $scope.$on("mlgSyncApp", sync)
    var sync = function() {
      $scope.messages = mlgApi.getMessages();
    }
    sync();
    
  }]);

  melangeControllers.controller('DashboardCtrl', ['$scope', 'mlgHelper', 'mlgTiles', 'mlgApi',
  function($scope, mlgHelper, mlgTiles, mlgApi) {
    $scope.editDash = false;
    $scope.tiles = mlgHelper.promise([], mlgTiles.all());

    // Sync up if needed.
    $scope.$on("mlgSyncApp", sync)
    var sync = function() {
      $scope.newsfeed = mlgApi.getMessages();
    }
    sync();

  }]);

  melangeControllers.controller('ProfileCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    $scope.newProfile = false;

    mlgApi.currentProfile().then(function(data) {
      console.log(data);
      $scope.profile = data;
    },
    function(err) {
      if(err === true) {
        $scope.newProfile = true;
      } else {
        console.log("Couldn't get profile. Something went wrong.")
        console.log(err)
      }
    });

  }]);

  melangeControllers.controller('NewProfileCtrl', ['$scope', '$location', 'mlgApi',
  function($scope, $location, mlgApi) {
    $scope.profile = {};
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
