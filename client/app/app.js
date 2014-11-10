'use strict';

/* App Module */

var melangeSuffix = ".melange:7776";
var melangePluginSuffix=".plugins" + melangeSuffix;
var melangeAPI ="api" + melangeSuffix;

var melangeApp = angular.module('melangeApp', [
  'ngRoute',
  'ngResource',
  'melangeServices',
  'melangeControllers',
  'melangeFilters',
  'melangeDirectives',
  'ui.sortable',
]);

melangeApp.config(['$routeProvider',
  function($routeProvider) {
    // Setup the Application Routes
    $routeProvider
      // Application Routes
      .when('/dashboard', {
        templateUrl: 'partials/dashboard.html',
        controller: 'DashboardCtrl'
      })
      .when('/all', {
        templateUrl: 'partials/all.html',
        controller: 'AllCtrl'
      })
      // Profile Routes
      .when('/profile', {
        templateUrl: 'partials/profile/profile.html',
        controller: 'ProfileCtrl'
      })
      .when('/profile/edit', {
        templateUrl: 'partials/profile/edit.html',
        controller: 'EditProfileCtrl'
      })
      .when('/user/:alias', {
        templateUrl: 'partials/profile/profile.html',
        controller: 'UserProfileCtrl'
      })
      // Contact Routes
      .when('/contacts', {
        templateUrl: 'partials/contacts.html',
        controller: 'ContactsCtrl'
      })
      // Plugin Routes
      .when('/plugin/:pluginid/:action', {
        templateUrl: 'partials/plugin/loader.html',
        controller: 'PluginCtrl'
      })
      .when('/market', {
        templateUrl: 'partials/market/index.html',
        controller: 'MarketCtrl',
      })
      .when('/market/:action', {
        templateUrl: 'partials/market/index.html',
        controller: 'MarketCtrl',
      })
      // Settings Routes
      .when('/settings', {
        templateUrl: 'partials/settings/index.html',
        controller: 'SettingsCtrl'
      })
      // Identity Settings
      .when('/settings/identity', {
        templateUrl: 'partials/settings/identity.html',
        controller: 'SettingsCtrl'
      })
      .when('/settings/identity/new', {
        templateUrl: 'partials/settings/newIdentity.html',
        controller: 'NewIdentityCtrl'
      })
      .when('/settings/plugins', {
        templateUrl: 'partials/settings/plugins.html',
        controller: 'PluginSettingsCtrl'
      })
      .when('/settings/developer', {
        templateUrl: 'partials/settings/developer.html',
        controller: 'SettingsCtrl',
      })
      .when('/settings/developer/add', {
        templateUrl: 'partials/settings/developerAdd.html',
        controller: 'DeveloperSettingsCtrl',
      })
      .when('/settings/developer/view/:app', {
        templateUrl: 'partials/settings/developerView.html',
        controller: 'SettingsCtrl',
      })
      .when('/settings/developer/view/:app/update', {
        templateUrl: 'partials/settings/developerUpdate.html',
        controller: 'DeveloperSettingsCtrl',
      })
      .when('/settings/advanced', {
        templateUrl: 'partials/settings/advanced.html',
        controller: 'SettingsCtrl'
      })
      // Startup Routes
      .when('/setup', {
        templateUrl: 'partials/setup/index.html',
        controller: 'SetupCtrl'
      })
      // Exisiting Account Routes
      .when('/setup/link', {
        templateUrl: 'partials/setup/link.html',
        controller: 'SetupCtrl'
      })
      // New Account Routes
      .when('/setup/new', {
        templateUrl: 'partials/setup/new.html',
        controller: 'SetupCtrl'
      })
      .when('/setup/server', {
        templateUrl: 'partials/setup/server.html',
        controller: 'SetupCtrl'
      })
      .when('/setup/confirm', {
        templateUrl: 'partials/setup/confirm.html',
        controller: 'SetupCtrl'
      })
      .when('/startup', {
        templateUrl: 'partials/startup.html',
        controller: 'StartupCtrl'
      })
      .otherwise({
        redirectTo: '/dashboard'
      });
  }]);
