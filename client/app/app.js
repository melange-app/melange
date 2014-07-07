'use strict';

/* App Module */

var melangePluginSuffix=".plugins.melange.127.0.0.1.xip.io:9001";

var melangeApp = angular.module('melangeApp', [
  'ngRoute',
  'ngResource',
  'melangeServices',
  'melangeControllers',
  'melangeFilters',
  'melangeDirectives',
]);

melangeApp.config(['$routeProvider',
  function($routeProvider) {
    // Setup the Application Routes
    $routeProvider.
      // Application Routes
      when('/dashboard', {
        templateUrl: 'partials/dashboard.html',
        controller: 'DashboardCtrl'
      }).
      when('/profile', {
        templateUrl: 'partials/profile.html',
        controller: 'ProfileCtrl'
      }).
      when('/all', {
        templateUrl: 'partials/all.html',
        controller: 'AllCtrl'
      }).
      // Contact Routes
      when('/contacts', {
        templateUrl: 'partials/contacts.html',
        controller: 'ContactsCtrl'
      }).
      // Plugin Routes
      when('/plugin/:pluginid/:action', {
        templateUrl: 'partials/plugin/loader.html',
        controller: 'PluginCtrl'
      }).
      // Settings Routes
      when('/settings', {
        templateUrl: 'partials/settings/index.html',
        controller: 'SettingsCtrl'
      }).
      // Startup Routes
      when('/startup', {
        templateUrl: 'partials/startup/index.html',
        controller: 'StartupCtrl'
      }).
      // Exisiting Account Routes
      when('/startup/link', {
        templateUrl: 'partials/startup/link.html',
        controller: 'StartupCtrl'
      }).
      // New Account Routes
      when('/startup/new', {
        templateUrl: 'partials/startup/new.html',
        controller: 'StartupCtrl'
      }).
      when('/startup/server', {
        templateUrl: 'partials/startup/server.html',
        controller: 'StartupCtrl'
      }).
      when('/startup/confirm', {
        templateUrl: 'partials/startup/confirm.html',
        controller: 'StartupCtrl'
      }).
      otherwise({
        redirectTo: '/'
      });
  }]);
