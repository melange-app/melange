'use strict';

/* Controllers */
var melangeControllers = angular.module('melangeControllers');


melangeControllers.controller('SettingsCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi, $rootScope) {
    $scope.identities = mlgApi.identities();

    $scope.newIdentity = function() {

    };

  }]);

melangeControllers.controller('NewIdentityCtrl', ['$scope', '$location', 'mlgIdentity',
  function($scope, $location, mlgIdentity) {
    $scope.profile = mlgIdentity.profile;

    $scope.mailProviders = mlgIdentity.servers();
    $scope.addressProviders = mlgIdentity.trackers();

    $scope.save = function() {
      mlgIdentity.save(function() {
        $location.path("/settings/identity");
      }, function(err) {
        alert("Cannot save new identity.");
      });
    }
  }]);
