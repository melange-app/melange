'use strict';

(function() {
  /* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('SettingsCtrl', ['$scope', 'mlgIdentity', 'mlgHelper', '$rootScope',
    function($scope, mlgIdentity, mlgHelper, $rootScope) {
      mlgIdentity.list().then(function(data) {
        $scope.identities = data;
      })

      $scope.setCurrentIdentity = function(id) {
          // Set it in the API
          mlgIdentity.setCurrent(id);
      };

      $scope.copy = function(str) {
        if(typeof window.require === "function") {
          var clipboard = require('clipboard');
          clipboard.writeText(str);
        }
      }

    }]);

  melangeControllers.controller('NewIdentityCtrl', ['$scope', '$location', 'mlgIdentity',
    function($scope, $location, mlgIdentity) {
      $scope.profile = mlgIdentity.profile;

      $scope.mailProviders = mlgIdentity.servers();
      $scope.addressProviders = mlgIdentity.trackers();

      $scope.save = function() {
        mlgIdentity.save(function() {
          mlgIdentity.refresh();
          $scope.$emit("mlgRefreshApp");
          $location.path("/settings/identity");
        }, function(err) {
          alert("Cannot save new identity.");
        });
      }
    }]);
})();
