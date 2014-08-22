'use strict';

(function() {
  /* Controllers */
  var melangeControllers = angular.module('melangeControllers');


  melangeControllers.controller('SettingsCtrl', ['$scope', '$interval', '$sce', 'mlgApi', 'mlgIdentity', 'mlgHelper', '$rootScope',
    function($scope, $interval, $sce, mlgApi, mlgIdentity, mlgHelper, $rootScope) {
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

      $scope.working = false;
      $scope.updateStatus = "Check for Updates"
      $scope.btnType = "btn-primary"
      $scope.canUpdate = window.require;
      var dir = undefined;

      var installUpdate = function() {
        console.log("Starting to update.")
        // Install will shutdown the server, I have to shutdown node
        mlgApi.update.install({
          dir: dir,
        }).$promise.then(
          function(data) {
            console.log("Updated")
          },
          function(err) {
            console.log("Error updating")
            console.log(err)
          }
        );
      }

      var downloadUpdate = function() {
        $scope.working = true;
        $scope.updateStatus = "Downloading...";
        mlgApi.update.download($scope.update).$promise.then(function(obj) {
          $scope.downloadProgress = 0;
          var checker = $interval(function() {
              mlgApi.update.progress().$promise.then(function(obj) {
                if(obj["dir"] !== undefined) {
                  $scope.working = false;
                  $scope.btnType = "btn-danger";
                  $scope.updateStatus = "Install and Restart";
                  $scope.checkForUpdates = installUpdate;
                  $interval.cancel(checker);
                  dir = obj["dir"];
                } else if (obj["progress"] !== undefined) {
                  console.log(obj["progress"])
                  $scope.downloadProgress = (obj["progress"] * 100).toFixed(2);
                }
              })
          }, 500)
        }, function(err) {
          console.log("Error downloading update.");
          console.log(err);
        })
      }

      $scope.checkForUpdates = function() {
        $scope.working = true;
        mlgApi.update.check().$promise.then(function(obj) {
          $scope.working = false;
          $scope.btnType = "btn-success"
          $scope.updateStatus = "Download new Version " + obj.Version;
          $scope.update = obj;
          console.log(obj.Changelog);
          $scope.changelog = $sce.trustAsHtml(obj.Changelog);
          $scope.checkForUpdates = downloadUpdate;
        }, function(obj) {
          $scope.working = false;
          if(obj.status == 422) {
            $scope.updateStatus = "No new update.";
          } else {
            console.log("Error getting update.")
            console.log(obj);
          }
        })
      }

    }]);

  melangeControllers.controller('PluginSettingsCtrl', ['$scope', 'mlgPlugins',
    function($scope, mlgPlugins) {

      $scope.loadingStore = true;
      mlgPlugins.allFromStore().then(function(data) {
        $scope.loadingStore = false;
        $scope.store = data;
      });

      mlgPlugins.all().then(function(data) {
        $scope.plugins = data;
        $scope.hasPlugins = Object.keys(data).length > 0;
      });

      $scope.install = function(plugin) {
        plugin.Installing = true;
        mlgPlugins.install({
          "Repository": plugin.Repository,
        }).then(function(data) {
          plugin.Installing = false;
          console.log("Installed " + plugin.Id + " successfully.");
          plugin.Installed = true;
        })
      }

      $scope.uninstall = function(plugin) {
        var id = plugin.id;
        mlgPlugins.uninstall({
          "Id": id,
        }).then(function(data) {
          console.log("Uninstalled " + id + " successfully.");
        })
      }
    }
  ]);

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
