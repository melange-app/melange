'use strict';

var melangeControllers = angular.module('melangeControllers', []);

melangeControllers.controller('ApplicationCtrl', ['$scope', '$location', '$route', 'mlgApi', 'mlgPlugins',
  function($scope, $location, $route, api, plugins) {

    $scope.$watch(function() { return $location.path(); }, function(path) { $scope.page = path; });
    $scope.reload = $route.reload;
    $scope.allPlugins = plugins.query();

    $scope.currentIdentity = api.current();
    $scope.allIdentities = api.identities();

    $scope.containerClass = function(page) {
      if (page === undefined) { return }

      if (page.indexOf('/plugin') === 0 || page.indexOf('/settings') === 0) {
        return ['main']
      }

      if (page.indexOf('/setup') === 0 || page.indexOf('/startup') === 0) {
        return ['container-fluid', 'setup-container']
      }

      return ['container-fluid', 'main']
    }

    $scope.outerClass = function(page) {
      if (page === undefined) { return }

      var title = page.substring(1).replace('/', '-');
      return title
    }

  }]);

melangeControllers.controller('StartupCtrl', ['mlgApi', '$location',
  function(mlgApi, $location) {

    mlgApi.current().$promise.then(
      function(obj) {
        $location.path("/dashboard");
      },
      function(obj) {
        if (obj.status == 422) {
          $location.path("/setup");
        } else {
          console.log("Error loading startup status.");
          console.log(obj.status);
          $location.path("/error");
        }
      },
      function(data) {
        console.dir(data);
      });

  }]);
