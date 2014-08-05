'use strict';

var melangeControllers = angular.module('melangeControllers');


melangeControllers.controller('PluginCtrl', ['$scope', '$routeParams',
function($scope, $routeParams) {
  // Load the Plugin, boys!
  $scope.pluginUrl = "http://" + $routeParams.pluginid + melangePluginSuffix + "/" + $routeParams.action;
}]);
