'use strict';

(function() {
  /* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('SetupCtrl', ['mlgIdentity', '$scope', '$location', 'mlgLink',
  function(mlgIdentity, $scope, $location, mlgLink) {
    $scope.profile = mlgIdentity.profile;

    $scope.mailProviders = mlgIdentity.servers();
    $scope.addressProviders = mlgIdentity.trackers();

    $scope.isLinking = false;
    $scope.linkCode = mlgLink.code;
    $scope.linkRequest = function() {
        $scope.isLinking = true;
        mlgLink.linkRequest($scope.address).then(
            // Success is a full linking of the identity.
            function() {
                $location.path("/dashboard");
            },
            // Error has code
            function(msg) {
                console.log("Error requesting link: " + msg);
                console.dir(msg);
                $scope.linkError = msg;
                $location.path("/setup/link/error");
            },
            function(code) {
                mlgLink.code = code;
                $location.path("/setup/link/wait");
            }
        );
    }

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
