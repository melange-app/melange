'use strict';

(function() {
  var melangeControllers = angular.module('melangeControllers');

  melangeControllers.controller('PluginCtrl', ['$scope', '$routeParams', 'mlgFlash',
  function($scope, $routeParams, mlgFlash) {
    // Load the Plugin, boys!
    var location = mlgFlash.get("pluginLocation");
    if(location === undefined) { location = $routeParams.action; }
    $scope.pluginUrl = "http://" + $routeParams.pluginid + melangePluginSuffix + "/" + location;
  }]);

  melangeControllers.controller('MarketCtrl', ['$scope', 'mlgPlugins',
  function($scope, mlgPlugins) {
      $scope.loadingStore = true;
      mlgPlugins.allFromStore().then(function(data) {
          $scope.loadingStore = false;
          $scope.store = data;
      });

      $scope.install = function(plugin) {
          plugin.Installing = true;
          mlgPlugins.install({
              "Repository": plugin.Repository,
          }).then(function(data) {
              plugin.Installing = false;
              console.log("Installed " + plugin.Id + " successfully.");
              plugin.Installed = true;
          }, function(data) {
              plugin.Installing = false;
              plugin.Error = true;
          })
      }
  }]);

})();
