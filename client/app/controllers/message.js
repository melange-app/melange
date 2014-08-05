'use strict';

var melangeControllers = angular.module('melangeControllers');


melangeControllers.controller('AllCtrl', ['$scope', 'mlgApi',
function($scope, mlgApi) {
    $scope.messages = mlgApi.getMessages();
}]);

melangeControllers.controller('DashboardCtrl', ['$scope', 'mlgApi',
function($scope, mlgApi) {
  $scope.tiles = [
    {
      size: "12",
      height: "100",
      url: "http://" + "ch.airdispat.plugins.status" + melangePluginSuffix + "/tile.html",
      click: true,
    },
    {
      size: "6",
      height: "150",
      url: "http://" + "ch.airdispat.plugins.news" + melangePluginSuffix + "/tile.html",
      name: "News",
    },
    {
      size: "6",
      height: "150",
      url: "http://" + "ch.airdispat.plugins.news" + melangePluginSuffix + "/tile.html",
      name: "Family",
    }
  ];

  $scope.newsfeed = mlgApi.getMessages();
}]);

melangeControllers.controller('ProfileCtrl', ['$scope',
function($scope) {
  $scope.name = "Hunter Leath";
  $scope.img = "http://i.imgur.com/mQtMWjg.jpg";
  $scope.description = "Ipsum anim mollit sunt elit ex reprehenderit consectetur consequat anim irure. Veniam excepteur anim nostrud elit elit exercitation laboris. Cillum sint mollit minim laborum qui ex ipsum exercitation exercitation ex duis. Ex incididunt sunt et aliqua veniam incididunt minim irure proident ad nostrud voluptate exercitation aliqua.";
}]);
