'use strict';

/* Controllers */

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

      if (page.indexOf('/startup') === 0) {
        return ['container-fluid', 'startup-container']
      }

      return ['container-fluid', 'main']
    }

    $scope.outerClass = function(page) {
      if (page === undefined) { return }

      var title = page.substring(1).replace('/', '-');
      return title
    }

  }]);

melangeControllers.controller('SettingsCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    $scope.identities = mlgApi.identities()
  }]);

melangeControllers.controller('ContactsCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    $scope.lists = mlgApi.lists()
    $scope.contacts = mlgApi.contacts()
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

melangeControllers.controller('StartupCtrl', ['mlgIdentity', '$scope', '$location',
  function(mlgIdentity, $scope, $location) {

    $scope.profile = mlgIdentity.profile;

    $scope.mailProviders = mlgIdentity.servers();
    $scope.addressProviders = mlgIdentity.trackers();

    $scope.finish = function() {
      mlgIdentity.profile.nickname = "Primary";
      mlgIdentity.save(function() {
        $location.path("/dashboard");
      }, function() {
        alert("Error creating Identity.");
      });
    }
  }]);
