'use strict';

(function() {
  var melangeControllers = angular.module('melangeControllers', []);

  melangeControllers.controller('ApplicationCtrl', ['$scope', '$location', '$route', '$interval', 'mlgIdentity', 'mlgPlugins', 'mlgHelper', 'mlgApi',
    function($scope, $location, $route, $interval, mlgIdentity, mlgPlugins, mlgHelper, mlgApi) {

      mlgPlugins.all().then(function(plugins) {
        $scope.allPlugins = plugins;
      })

      mlgApi.contacts().then(function(data) {
        $scope.contacts = data;
      });

      mlgIdentity.current().then(function(id) {
        $scope.currentIdentity = id;
      });

      mlgIdentity.list().then(function(ids) {
        $scope.allIdentities = ids;
      })

      $scope.syncInProgress = false;
      var sync = function() {
        console.log("Syncing.");
        $scope.syncInProgress = true;
        // loadApp();
        $scope.$broadcast("mlgSyncApp");
        $scope.syncInProgress = false;
      }
      $scope.sync = sync;

      var autoSync = $interval(sync, 300000);
      $scope.$on("$destroy", function() {
        if(autoSync !== undefined) {
          autoSync.cancel();
          autoSync = undefined;
        }
      })

      $scope.$watch(function() { return $location.path(); }, function(path) { $scope.page = path; });
      $scope.reload = $route.reload;

      $scope.switchId = function(id) {
        mlgIdentity.setCurrent(id).$promise.then(function() {
          sync();
        })
      }

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
