'use strict';

/* Directives */

var melangeDirectives = angular.module('melangeDirectives', []);

melangeDirectives.directive("mlgTile", function() {
  return {
    templateUrl: "partials/directives/tile.html",
    restrict: "E",
    scope: {
      height: "=height",
      url: "=src",
      name: "=name",
      click: "=clickthrough",
    },
    link: function(element, attrs, scope) {

    },
  }
});

melangeDirectives.directive("modal", function() {
  return {
    templateUrl: "partials/directives/modal.html",
    restrict: "E",
    scope: {
      data: "=data",
      name: "=name",
    },
  }
});

melangeDirectives.directive("message", function() {
  return {
    templateUrl: "partials/directives/messageViewer.html",
    restrict: "E",
    scope: {
      data: "=data",
    },
  }
});
