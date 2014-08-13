'use strict';

(function() {
  var melangeControllers = angular.module('melangeControllers', []);

  melangeControllers.controller('ApplicationCtrl', ['$scope', '$location', '$route', '$interval', 'mlgIdentity', 'mlgPlugins', 'mlgHelper', 'mlgApi',
    function($scope, $location, $route, $interval, mlgIdentity, mlgPlugins, mlgHelper, mlgApi) {

      var loadApp = function() {
        $scope.allPlugins = mlgHelper.promise([], mlgPlugins.all());

        $scope.contacts = mlgApi.contacts();

        $scope.currentIdentity = mlgHelper.promise({}, mlgIdentity.current());
        $scope.allIdentities = mlgHelper.promise([], mlgIdentity.list());
      }

      $scope.$on("mlgRefreshApp", function(e, args) {
        console.log("Refreshing...");
        loadApp()
      });
      loadApp();

      $scope.syncInProgress = false;
      var sync = function() {
        $scope.syncInProgress = true;
        loadApp();
        $scope.$broadcast("mlgSyncApp");
        $scope.syncInProgress = false;
      }
      $scope.sync = sync;

      var autoSync = $interval(sync, 30000);
      $scope.$on("$destroy", function() {
        if(autoSync !== undefined) {
          autoSync.cancel();
          autoSync = undefined;
        }
      })

      $scope.$watch(function() { return $location.path(); }, function(path) { $scope.page = path; });
      $scope.reload = $route.reload;

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

  melangeControllers.controller('StartupCtrl', ['mlgIdentity', '$location',
    function(mlgIdentity, $location) {

      mlgIdentity.startup().then(
        function(obj) {
          if(obj) {
            $location.path("/dashboard");
          } else {
            $location.path("/setup");

          }
        },
        function(obj) {
          console.log("Error loading startup status.");
          $location.path("/error");
        },
        function(data) {
          console.dir(data);
        });

    }]);

})()
