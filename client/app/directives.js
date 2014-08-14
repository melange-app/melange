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
      edit: "=edit",
    },
    link: function(element, attrs, scope) {

    },
  }
});

melangeDirectives.directive("mlgAddTile", function() {
  return {
    templateUrl: "partials/directives/addTile.html",
    restrict: "E",
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

melangeDirectives.directive("message", ['mlgPlugins', function(mlgPlugins) {
  return {
    templateUrl: "partials/directives/messageViewer.html",
    restrict: "E",
    scope: {
      data: "=data",
    },
    link: function(scope, elem, attrs) {
      var thePlugin = undefined;
      var theFrame = undefined;

      scope.$on("$destroy", function() {
        mlgPlugins.unregisterPlugin(thePlugin, theFrame)
      });

      scope.register = function() {
        var frame = $(elem).find("iframe");
        if(frame.length === 1) {
          theFrame = frame[0];
          mlgPlugins.registerPlugin(thePlugin, theFrame, "viewer", scope.data);
        }
      }

      scope.$watch("data", function(d) {
        if(typeof d === "string")
          return
        mlgPlugins.viewer(d).then(function(v) {
          thePlugin = v[0];
          scope.plugin = thePlugin;
          scope.templateType = "remote";
          scope.url = "http://" + v[0].id + melangePluginSuffix + "/" + v[0].viewers[v[1]].view;
        }, function() { scope.templateType = "default"; });
      })
    }
  }
}]);
