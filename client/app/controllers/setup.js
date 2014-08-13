'use strict';

(function() {
  /* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('SetupCtrl', ['mlgIdentity', '$scope', '$location',
  function(mlgIdentity, $scope, $location) {
    $scope.profile = mlgIdentity.profile;

    $scope.mailProviders = mlgIdentity.servers();
    $scope.addressProviders = mlgIdentity.trackers();

    $scope.finish = function() {
      mlgIdentity.profile.nickname = "Primary";
      mlgIdentity.save(function() {
        mlgIdentity.profile = {};
        mlgIdentity.refresh();
        $scope.$emit("mlgRefreshApp");
        $location.path("/dashboard");
      }, function() {
        alert("Error creating Identity.");
      });
    }
  }]);
})();
