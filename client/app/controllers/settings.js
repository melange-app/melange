'use strict';

(function() {
  /* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('SettingsCtrl', ['$scope', 'mlgIdentity', 'mlgHelper', '$rootScope',
    function($scope, mlgIdentity, mlgHelper, $rootScope) {
      $scope.identities = mlgHelper.promise([], mlgIdentity.list());

      $scope.setCurrentIdentity = function(id) {
          // Reload the view
          for(var i in $scope.identities) {
            $scope.identities[i].Current = false;
          }
          id.Current = true;

          // Set it in the API
          mlgIdentity.setCurrent({
            fingerprint: id.Fingerprint,
          });
          $scope.$emit("mlgRefreshApp");
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
