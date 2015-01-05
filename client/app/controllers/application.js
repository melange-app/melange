'use strict';

(function() {
  var melangeControllers = angular.module('melangeControllers', []);

  melangeControllers.controller('ApplicationCtrl', ['$scope', '$location', '$route', '$interval', 'mlgIdentity', 'mlgPlugins', 'mlgHelper', 'mlgApi', 'mlgCandyBar', 'mlgRealtime',
    function($scope, $location, $route, $interval, mlgIdentity, mlgPlugins, mlgHelper, mlgApi, mlgCandyBar, mlgRealtime) {
      mlgPlugins.all().then(function(plugins) {
        $scope.allPlugins = plugins;
      });

      mlgApi.contacts().then(function(data) {
        $scope.contacts = data;
      });

      mlgIdentity.current().then(function(id) {
        $scope.currentIdentity = id;
      });

      mlgIdentity.list().then(function(ids) {
        $scope.allIdentities = ids;
      })

      $scope.candy = mlgCandyBar.status();

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

        if (
            page.indexOf('/plugin') === 0 ||
            page.indexOf('/settings') === 0 ||
            page.indexOf('/market') === 0 ||
            page.indexOf('/contacts') === 0
        ) {
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

      $scope.hasButtons = (typeof window.require === 'function');
      $scope.appClose = function() {
        var remote = require('remote')
        remote.getCurrentWindow().close();
      }
      $scope.appMinimize = function() {
        var remote = require('remote')
        remote.getCurrentWindow().minimize();
      }
      $scope.appZoom = function() {
        var remote = require('remote')
        var win = remote.getCurrentWindow();
        if(win.isMaximized()) {
          win.unmaximize();
        } else {
          win.maximize();
        }
      }

      // Mobile Sidebar Toggling
      $scope.isShowing = "not-showing";
      $scope.sidebarButtonIcon = "fa-bars";
      $scope.toggleSidebar = function() {
          if ($scope.isShowing == "not-showing") {
              $scope.isShowing = "showing";
              $scope.sidebarButtonIcon = "fa-arrow-left";
          } else {
              $scope.isShowing = "not-showing";
              $scope.sidebarButtonIcon = "fa-bars";
          }
      }
      $scope.closeSidebar = function() {
          $scope.isShowing = "not-showing";
          $scope.sidebarButtonIcon = "fa-bars";
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
