'use strict';

var notesPlugin = angular.module('notesPlugin', [
  'ngRoute',
  'melangeUi',
  'notesControllers',
]);

notesPlugin.config(['$routeProvider',
  function($routeProvider) {
    // Setup the Application Routes
    $routeProvider.
      // Application Routes
      when('/', {
        templateUrl: 'templates/list.html',
        controller: 'ListCtrl'
      }).
      when('/new', {
        templateUrl: 'templates/new.html',
        controller: 'NewCtrl',
      }).
      when('/profile/:profileId', {
        templateUrl: 'templates/profile.html',
        controller: 'ProfileCtrl',
      }).
      when('/note/:noteId', {
        templateUrl: 'template/note.html',
        controller: 'NoteCtrl',
      }).
      when('/note/:noteId/edit', {
        templateUrl: 'template/edit.html',
        controller: 'EditCtrl',
      }).
      otherwise({
        redirectTo: '/'
      });
  }]);
