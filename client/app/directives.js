'use strict';

/* Directives */

var melangeDirectives = angular.module('melangeDirectives', []);

melangeDirectives.directive("mlgTile", function() {
  return {
    templateUrl: "partials/tile.html",
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
