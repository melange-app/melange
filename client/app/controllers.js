'use strict';

/* Controllers */

var melangeControllers = angular.module('melangeControllers', []);

melangeControllers.controller('SidebarCtrl', ['$scope', '$location', '$route', 'mlgPlugins',
  function($scope, $location, $route, plugins) {

    $scope.$watch(function() { return $location.path(); }, function(path) { $scope.page = path; });
    $scope.reload = $route.reload;
    $scope.allPlugins = plugins.query();

    $scope.containerClass = function(page) {
      if (page === undefined) { return }

      if (page.indexOf('/startup') === 0) {
        return ['container']
      } else {
        if (page.indexOf('/plugin') !== 0) {
          return ['container-fluid', 'main']
        }
        return ['main']
      }
    }

    $scope.outerClass = function(page) {
      if (page === undefined) { return }

      var title = page.substring(1).replace('/', '-');
      return title
    }

  }]);

melangeControllers.controller('SettingsCtrl', ['$scope',
  function($scope) {

  }]);

melangeControllers.controller('ContactsCtrl', ['$scope',
  function($scope) {

  }]);

melangeControllers.controller('DashboardCtrl', ['$scope',
  function($scope, $sce) {
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
  }]);

melangeControllers.controller('ProfileCtrl', ['$scope',
  function($scope) {
    $scope.name = "Hunter Leath";
    $scope.img = "http://i.imgur.com/mQtMWjg.jpg";
    $scope.description = "Ipsum anim mollit sunt elit ex reprehenderit consectetur consequat anim irure. Veniam excepteur anim nostrud elit elit exercitation laboris. Cillum sint mollit minim laborum qui ex ipsum exercitation exercitation ex duis. Ex incididunt sunt et aliqua veniam incididunt minim irure proident ad nostrud voluptate exercitation aliqua.";
  }]);

melangeControllers.controller('PluginCtrl', ['$scope', '$routeParams',
  function($scope, $routeParams) {
    // Load the Plugin, boys!
    $scope.pluginUrl = "http://" + $routeParams.pluginid + melangePluginSuffix + "/" + $routeParams.action;
  }]);

melangeControllers.controller('StartupCtrl', ['$scope',
  function($scope) {

    $scope.mailProviders = [{
      name: 'AirDispatch.Me',
      description: 'The first Melange provider.',
      img: 'http://placehold.it/64x64',
    }];

    $scope.addressProviders = [{
      name: 'AirDispatch.Me',
      description: 'The first Melange provider.',
      img: 'http://placehold.it/64x64',
      suffix: 'airdispat.ch',
    },
    {
      name: 'Virginia.edu',
      description: 'Register to be affiliated with UVa.',
      img: 'http://placehold.it/64x64',
      suffix: 'virginia.edu',
    }];

    $scope.selectedAddress = -1;
    $scope.selectedMail = -1;

  }]);
