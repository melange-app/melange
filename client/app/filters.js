'use strict';

/* Filters */

var melangeFilters = angular.module('melangeFilters', [])

melangeFilters.filter('unsafe', ['$sce', function($sce) {
    return function(val) {
        return $sce.trustAsHtml(val);
    };
}]);

melangeFilters.filter('unsafeUrl', ['$sce', function($sce) {
    return function(val) {
        return $sce.trustAsResourceUrl(val);
    };
}]);

melangeFilters.filter('objectsHave', function() {
  return function(val, obj) {
    var newObj = {};
    for (var key in val) {
      if(val[key][obj] !== null) {
        newObj[key] = val[key];
      }
    }
    return newObj;
  }
})

melangeFilters.filter('escapeBackground', function() {
  return function(val) {
    return val.replace("'", "%27")
  }
})
